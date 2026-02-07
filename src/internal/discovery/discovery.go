package discovery

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"

	"envsync/internal/config"
	"envsync/internal/keys"
	"envsync/internal/logging"
	httptransport "envsync/internal/transport/http"
	sshtransport "envsync/internal/transport/ssh"
)

type Options struct {
	Timeout     time.Duration
	Quiet       bool
	Verbose     bool
	FilterSSH   bool
	CollectKeys bool
	ShowPubkeys bool
}

func Discover(opts Options) ([]string, error) {
	if !opts.Quiet {
		logging.Log("INFO", fmt.Sprintf("Discovering env-sync peers (timeout: %ds)...", int(opts.Timeout.Seconds())))
	}

	var peers []string
	var err error

	switch runtime.GOOS {
	case "linux":
		peers, err = discoverAvahi(opts.Timeout)
	case "darwin":
		peers, err = discoverDnssd(opts.Timeout)
	default:
		peers, err = discoverFallback(opts.Timeout)
	}

	if err != nil {
		return nil, err
	}

	peers = uniqueSorted(peers)

	if opts.FilterSSH {
		filtered := make([]string, 0)
		for _, peer := range peers {
			if sshtransport.TestSSH(peer) == nil {
				filtered = append(filtered, peer)
			}
		}
		peers = filtered
	}

	if opts.Quiet {
		return peers, nil
	}

	if len(peers) == 0 {
		logging.Log("WARN", "No env-sync peers found on local network")
		fmt.Println("")
		return peers, nil
	}

	logging.Log("SUCCESS", fmt.Sprintf("Found %d peer(s):", len(peers)))
	for _, peer := range peers {
		info := ""
		if opts.Verbose {
			health, err := FetchHealth(peer)
			if err == nil && health.Version != "" {
				info = fmt.Sprintf(" (version: %s)", health.Version)
			}
		}
		if opts.CollectKeys || opts.ShowPubkeys {
			pubkey := fetchPubkey(peer)
			if pubkey != "" {
				_ = keys.CachePeerPubkey(peer, pubkey)
				if opts.ShowPubkeys {
					info = info + fmt.Sprintf(" [pubkey: %s...]", truncate(pubkey, 20))
				}
			}
		}
		fmt.Printf("  - %s%s\n", peer, info)
	}

	if opts.CollectKeys {
		logging.Log("INFO", "Collecting public keys from peers...")
		collected := 0
		for _, peer := range peers {
			pubkey := fetchPubkey(peer)
			if pubkey != "" {
				_ = keys.CachePeerPubkey(peer, pubkey)
				logging.Log("SUCCESS", "Cached public key from "+peer)
				collected++
			} else {
				logging.Log("WARN", "Could not fetch public key from "+peer)
			}
		}
		if collected > 0 {
			logging.Log("SUCCESS", fmt.Sprintf("Collected %d public key(s)", collected))
			logging.Log("INFO", "Run 'env-sync' to sync with new recipients")
		}
	}

	return peers, nil
}

func discoverAvahi(timeout time.Duration) ([]string, error) {
	if _, err := exec.LookPath("avahi-browse"); err != nil {
		logging.Log("ERROR", "avahi-browse not found. Install with: sudo apt-get install avahi-utils")
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "avahi-browse", "-t", "-r", "-p", config.Service)
	output, _ := cmd.Output()
	lines := strings.Split(string(output), "\n")
	peers := []string{}
	for _, line := range lines {
		if !strings.HasPrefix(line, "=") {
			continue
		}
		fields := strings.Split(line, ";")
		if len(fields) < 7 {
			continue
		}
		hostname := strings.TrimSuffix(fields[6], ".")
		if hostname != "" {
			peers = append(peers, hostname)
		}
	}
	return peers, nil
}

func discoverDnssd(timeout time.Duration) ([]string, error) {
	if _, err := exec.LookPath("dns-sd"); err != nil {
		logging.Log("ERROR", "dns-sd not found. This should be built into macOS.")
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "dns-sd", "-B", config.Service, "local.")
	output, _ := cmd.Output()
	lines := strings.Split(string(output), "\n")
	peers := []string{}
	for _, line := range lines {
		if !strings.HasPrefix(strings.TrimSpace(line), "+") || !strings.Contains(line, config.Service) {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		name := fields[len(fields)-1]
		infoCtx, cancelInfo := context.WithTimeout(context.Background(), 2*time.Second)
		infoCmd := exec.CommandContext(infoCtx, "dns-sd", "-L", name, config.Service, "local.")
		infoOutput, _ := infoCmd.Output()
		cancelInfo()
		for _, infoLine := range strings.Split(string(infoOutput), "\n") {
			if !strings.HasPrefix(strings.TrimSpace(infoLine), "+") {
				continue
			}
			parts := strings.Fields(infoLine)
			if len(parts) >= 7 {
				target := strings.TrimSuffix(parts[6], ".")
				if target != "" {
					peers = append(peers, target)
				}
			}
		}
	}
	return peers, nil
}

func discoverFallback(timeout time.Duration) ([]string, error) {
	logging.Log("WARN", "Using fallback discovery method (limited functionality)")

	commonHosts := []string{
		"beelink.local",
		"mbp16.local",
		"razer.local",
		"macbook.local",
		"macbook-pro.local",
		"macbook-air.local",
		"imac.local",
		"mac-mini.local",
		"ubuntu.local",
		"debian.local",
		"fedora.local",
		"linux.local",
		"raspberrypi.local",
		"pi.local",
	}

	peers := []string{}
	for _, host := range commonHosts {
		if host == sshtransport.Hostname() {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		ping := exec.CommandContext(ctx, "ping", "-c", "1", "-W", "1", host)
		_ = ping.Run()
		cancel()
		health, err := FetchHealth(host)
		if err == nil && health.Status == "ok" {
			peers = append(peers, host)
		}
	}
	return peers, nil
}

type HealthResponse = httptransport.HealthResponse

func FetchHealth(host string) (HealthResponse, error) {
	return httptransport.FetchHealth(host)
}

func fetchPubkey(host string) string {
	cmd := exec.Command("ssh", "-n", "-o", "ConnectTimeout=3", "-o", "StrictHostKeyChecking=accept-new", host, "cat ~/.config/env-sync/keys/age_key.pub")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func uniqueSorted(values []string) []string {
	seen := map[string]bool{}
	unique := []string{}
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		unique = append(unique, value)
	}
	sort.Strings(unique)
	return unique
}

func truncate(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max]
}

func ParseOptions(args []string) (Options, []string, error) {
	opts := Options{Timeout: 5 * time.Second}
	remaining := []string{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-t", "--timeout":
			if i+1 >= len(args) {
				return opts, remaining, errors.New("timeout required")
			}
			duration, err := time.ParseDuration(args[i+1] + "s")
			if err != nil {
				return opts, remaining, err
			}
			opts.Timeout = duration
			i++
		case "-q", "--quiet":
			opts.Quiet = true
		case "-v", "--verbose":
			opts.Verbose = true
		case "--ssh":
			opts.FilterSSH = true
		case "--collect-keys":
			opts.CollectKeys = true
		case "--pubkeys":
			opts.ShowPubkeys = true
		case "-h", "--help":
			return opts, remaining, errors.New("help")
		default:
			if strings.HasPrefix(args[i], "-") {
				return opts, remaining, fmt.Errorf("unknown option: %s", args[i])
			}
			remaining = append(remaining, args[i])
		}
	}
	return opts, remaining, nil
}
