package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"envsync/internal/backup"
	"envsync/internal/config"
	"envsync/internal/cron"
	"envsync/internal/discovery"
	"envsync/internal/keys"
	"envsync/internal/logging"
	"envsync/internal/metadata"
	"envsync/internal/secrets"
	"envsync/internal/server"
	svcmgr "envsync/internal/service"
	syncer "envsync/internal/sync"
	sshtransport "envsync/internal/transport/ssh"
)

func Run(argv []string) int {
	if len(argv) == 0 {
		return 1
	}
	base := filepath.Base(argv[0])
	args := argv[1:]

	// Handle global --verbose flag before routing to subcommands
	args = handleGlobalFlags(args)

	switch base {
	case "env-sync-client":
		return runSync(args, "env-sync-client", true)
	case "env-sync-discover":
		return runDiscover(args, "env-sync-discover")
	case "env-sync-serve":
		return runServe(args, "env-sync-serve")
	case "env-sync-key":
		return runKey(args)
	case "env-sync-load":
		return runLoad(args)
	}

	if len(args) == 0 {
		return runSync(args, "env-sync sync", false)
	}

	if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		showHelp()
		return 0
	}

	if args[0] == "version" || args[0] == "--version" {
		fmt.Println("env-sync version " + config.Version)
		return 0
	}

	// Handle legacy -v as version (but --verbose takes precedence in handleGlobalFlags)
	if args[0] == "-v" && !config.IsVerbose() {
		fmt.Println("env-sync version " + config.Version)
		return 0
	}

	if strings.HasPrefix(args[0], "-") {
		return runSync(args, "env-sync sync", false)
	}

	command := args[0]
	args = args[1:]

	switch command {
	case "sync", "s":
		return runSync(args, "env-sync sync", false)
	case "serve", "server":
		return runServe(args, "env-sync serve")
	case "discover", "d":
		return runDiscover(args, "env-sync discover")
	case "status", "st":
		return runStatus(args)
	case "init", "i":
		return runInit(args)
	case "restore", "r":
		return runRestore(args)
	case "cron", "c":
		return runCron(args)
	case "key", "k":
		return runKey(args)
	case "load", "l":
		return runLoad(args)
	case "add", "a":
		return runAdd(args)
	case "remove", "rm", "delete", "del":
		return runRemove(args)
	case "list", "ls":
		return runList(args)
	case "show", "get":
		return runShow(args)
	case "path", "p":
		return runPath(args)
	case "service", "svc":
		return runServiceManagement(args)
	case "help", "--help", "-h":
		showHelp()
		return 0
	case "version", "--version", "-v":
		fmt.Println("env-sync version " + config.Version)
		return 0
	default:
		logging.Log("ERROR", "Unknown command: "+command)
		showHelp()
		return 1
	}
}

// handleGlobalFlags processes global flags (like --verbose) and returns remaining args
func handleGlobalFlags(args []string) []string {
	remaining := []string{}
	for _, arg := range args {
		if arg == "--verbose" {
			config.SetVerbose(true)
		} else {
			remaining = append(remaining, arg)
		}
	}
	return remaining
}

func showHelp() {
	fmt.Print(`env-sync - Distributed secrets synchronization tool

Usage: env-sync [global options] [command] [options]

Global Options:
  --verbose                Print verbose output including commands executed

SECURITY NOTICE:
  By default, env-sync uses SCP (SSH) for secure peer-to-peer synchronization.
  This requires SSH keys to be set up between machines.
  
  HTTP mode (--insecure-http) is available as a fallback but is INSECURE:
  • Secrets transmitted in plaintext
  • Accessible to any device on the network
  • No authentication required

Commands:
  sync [options]           Sync secrets from network (default)
    Options:
      -a, --all            Sync from all discovered peers
      -f, --force          Force sync even if local is newer
      -q, --quiet          Quiet mode
      --insecure-http      Use insecure HTTP instead of SCP (not recommended)
      -h, --help           Show help

  serve [options]          Start HTTP server (for HTTP mode only)
    Options:
      -p, --port PORT      Port to listen on (default: 5739)
      -d, --daemon         Run as daemon service (auto-restart, 30m sync, mDNS)
      -q, --quiet          Quiet mode
      -h, --help           Show help

  discover [options]       Discover peers on network
    Options:
      -t, --timeout SECS   Discovery timeout (default: 5)
      -q, --quiet          Only output hostnames
      -v, --verbose        Verbose output
      --ssh                Filter for hosts with SSH access
      -h, --help           Show help

  status                   Show sync status and peer information

  init                     Initialize new secrets file

  restore [n]              Restore from backup (n=1-5, default: 1)

  key [subcommand]         Manage AGE encryption keys
    Subcommands:
      show                 Show this machine's public key
      export               Export public key
      import <pubkey>      Import a peer's public key
      list                 List all cached public keys
      request-access       Request access for new machine
      grant-access         Grant access to requesting machine

  load [options]           Load secrets for shell integration
    Options:
      --format <env|json>  Output format (default: env)

  add <KEY="value">         Add or update a secret key-value pair
    Example: env-sync add OPENAI_API_KEY="sk-..."

  remove <KEY>              Remove a secret key from the file
    Example: env-sync remove OPENAI_API_KEY

  list                     List all secret keys (values hidden)

  show <KEY>                Show value of a specific key

  path [options]            Show paths to env-sync files
    Options:
      --backup             Show backup directory path

  cron [options]           Setup cron job for periodic sync
    Options:
      --install            Install cron job (30min interval)
      --remove             Remove cron job
      --show               Show current cron job

  service [subcommand]     Manage env-sync background service
    Subcommands:
      stop                 Stop the running service
      restart              Restart the service
      uninstall            Uninstall the service completely

  help                     Show this help message

Examples:
  env-sync                           # Run sync via SCP (secure, default)
  env-sync serve -d                  # Start HTTP server (for HTTP mode)
  env-sync discover                  # Find peers on network
  env-sync status                    # Show current status
  env-sync init                      # Create new secrets file

Setup Instructions:
  1. Ensure SSH keys are set up between machines:
     ssh-copy-id other-machine.local
  
  2. Initialize with encryption:
     env-sync init --encrypted
     env-sync                         # Uses SCP with encryption
  
  3. Adding a new machine to existing fleet:
     # On new machine (D):
     env-sync init --encrypted
     env-sync key request-access --trigger beelink.local
     env-sync                         # Now can decrypt and sync
  
  4. Shell integration (add to .bashrc):
     eval "$(env-sync load 2>/dev/null)"
  
  5. To use insecure HTTP mode (not recommended):
     env-sync serve -d                # Start HTTP server
     env-sync --insecure-http         # Sync via HTTP

Files:
  ~/.secrets.env           # Main secrets file
  ~/.config/env-sync/      # Configuration directory
  ~/.config/env-sync/backups/  # Backup files

Environment Variables:
  ENV_SYNC_QUIET=true      # Suppress output
  ENV_SYNC_PORT=5739       # Server port
  ENV_SYNC_STRICT_SSH=true # Require known_hosts entries for SSH

`)
}

func runSync(args []string, usageName string, isClient bool) int {
	opts := syncer.Options{}
	remaining := []string{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-a", "--all":
			opts.AllPeers = true
		case "-f", "--force":
			opts.Force = true
		case "--force-pull":
			opts.ForcePull = true
		case "-q", "--quiet":
			opts.Quiet = true
		case "--insecure-http":
			opts.InsecureHTTP = true
		case "-h", "--help":
			if isClient {
				fmt.Println("Usage: " + usageName + " [options] [hostname]")
			} else {
				fmt.Println("Usage: " + usageName + " [options]")
			}
			fmt.Println("")
			fmt.Println("Syncs secrets from peers using SCP (SSH) by default.")
			fmt.Println("")
			fmt.Println("Options:")
			fmt.Println("  -a, --all              Sync from all discovered peers")
			fmt.Println("  -f, --force            Force sync even if local is newer")
			fmt.Println("  --force-pull           Force overwrite all local secrets from specified host")
			fmt.Println("  -q, --quiet            Quiet mode")
			fmt.Println("  --insecure-http        Use insecure HTTP instead of SCP (not recommended)")
			if isClient {
				fmt.Println("  hostname               Specific hostname to sync from")
			}
			fmt.Println("")
			fmt.Println("Examples:")
			if isClient {
				fmt.Println("  env-sync-client                           # Sync from newest peer via SCP")
				fmt.Println("  env-sync-client hostname.local            # Sync from specific host via SCP")
				fmt.Println("  env-sync-client --force-pull hostname.local   # Force overwrite from specific host")
				fmt.Println("  env-sync-client --insecure-http           # Sync using HTTP (INSECURE)")
			} else {
				fmt.Println("  env-sync sync                           # Sync via SCP (secure)")
				fmt.Println("  env-sync sync mbp16.local               # Sync from specific host")
				fmt.Println("  env-sync sync --force-pull mbp16.local  # Force overwrite from specific host")
				fmt.Println("  env-sync sync --insecure-http           # Sync via HTTP (INSECURE)")
			}
			return 0
		default:
			if strings.HasPrefix(args[i], "-") {
				logging.Log("ERROR", fmt.Sprintf("Unknown option: %s", args[i]))
				return 1
			}
			remaining = append(remaining, args[i])
		}
	}
	if len(remaining) > 1 {
		logging.Log("ERROR", "Multiple hostnames specified")
		return 1
	}
	if len(remaining) == 1 {
		opts.TargetHost = remaining[0]
	}

	if opts.ForcePull && opts.TargetHost == "" {
		logging.Log("ERROR", "--force-pull requires a specific hostname")
		logging.Log("INFO", "Usage: env-sync sync --force-pull hostname.local")
		return 1
	}

	if err := syncer.Run(opts); err != nil {
		return 1
	}
	return 0
}

func runServe(args []string, usageName string) int {
	opts := server.Options{Port: config.EnvSyncPort()}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-p", "--port":
			if i+1 >= len(args) {
				logging.Log("ERROR", fmt.Sprintf("Unknown option: %s", args[i]))
				return 1
			}
			opts.Port = args[i+1]
			i++
		case "-d", "--daemon":
			opts.Daemon = true
		case "--service":
			opts.Service = true
		case "-q", "--quiet":
			opts.Quiet = true
			_ = os.Setenv("ENV_SYNC_QUIET", "true")
		case "-h", "--help":
			fmt.Println("Usage: " + usageName + " [options]")
			fmt.Println("Options:")
			fmt.Println("  -p, --port PORT    Port to listen on")
			fmt.Println("  -d, --daemon       Run as daemon service (auto-restart, 30m sync, mDNS)")
			fmt.Println("  -q, --quiet        Quiet mode")
			return 0
		default:
			logging.Log("ERROR", fmt.Sprintf("Unknown option: %s", args[i]))
			return 1
		}
	}

	if _, err := os.Stat(config.SecretsFile()); err != nil {
		logging.Log("ERROR", "Secrets file not found: "+config.SecretsFile())
		logging.Log("INFO", "Run 'env-sync init' to create one")
		return 1
	}

	if err := server.Run(opts); err != nil {
		return 1
	}
	return 0
}

func runDiscover(args []string, usageName string) int {
	opts, _, err := discovery.ParseOptions(args)
	if err != nil {
		if err.Error() == "help" {
			fmt.Println("Usage: " + usageName + " [options]")
			fmt.Println("Options:")
			fmt.Println("  -t, --timeout SECONDS  Discovery timeout (default: 5)")
			fmt.Println("  -q, --quiet            Only output hostnames")
			fmt.Println("  -v, --verbose          Verbose output")
			fmt.Println("  --ssh                  Filter for hosts with SSH access")
			fmt.Println("  --collect-keys         Collect and cache peer public keys")
			fmt.Println("  --pubkeys              Show discovered public keys")
			return 0
		}
		logging.Log("ERROR", fmt.Sprintf("Unknown option: %s", err.Error()))
		return 1
	}
	peers, err := discovery.Discover(opts)
	if err != nil {
		return 1
	}
	if opts.Quiet {
		if len(peers) > 0 {
			fmt.Println(strings.Join(peers, "\n"))
		}
		return 0
	}
	return 0
}

func runStatus(_ []string) int {
	fmt.Println("=== env-sync Status ===")
	fmt.Println("")

	fmt.Println("Local Secrets File:")
	if _, err := os.Stat(config.SecretsFile()); err == nil {
		version := metadata.GetFileVersion(config.SecretsFile())
		timestamp := metadata.GetFileTimestamp(config.SecretsFile())
		host := metadata.GetFileHost(config.SecretsFile())
		checksum := metadata.GetFileChecksum(config.SecretsFile())
		fmt.Println("  File: " + config.SecretsFile())
		fmt.Println("  Version: " + version)
		fmt.Println("  Created: " + timestamp)
		fmt.Println("  Host: " + host)
		if checksum != "" {
			fmt.Printf("  Checksum: %s...\n", truncate(checksum, 16))
		}
		fmt.Println("")
	} else {
		fmt.Println("  File: " + config.SecretsFile() + " (NOT FOUND)")
		fmt.Println("  Run 'env-sync init' to create one")
		fmt.Println("")
	}

	fmt.Println("Local Server:")
	if health, err := fetchHealth("localhost"); err == nil {
		fmt.Println("  Status: Running")
		fmt.Println("  Port: " + config.EnvSyncPort())
		fmt.Println("  Version: " + health.Version)
	} else {
		fmt.Println("  Status: Not running")
		fmt.Println("  Run 'env-sync serve -d' to start")
	}
	fmt.Println("")

	fmt.Println("Discovered Peers:")
	peers, err := discovery.Discover(discovery.Options{Timeout: 5 * time.Second, Quiet: true})
	if err == nil && len(peers) > 0 {
		fmt.Printf("  Found %d peer(s):\n", len(peers))
		for _, peer := range peers {
			if health, err := fetchHealth(peer); err == nil {
				fmt.Printf("    ✓ %s (v%s)\n", peer, health.Version)
			} else {
				fmt.Printf("    ✗ %s (unreachable)\n", peer)
			}
		}
	} else {
		fmt.Println("  No peers found on local network")
	}
	fmt.Println("")

	fmt.Println("Backups:")
	if _, err := os.Stat(config.BackupDir()); err == nil {
		files, _ := filepath.Glob(filepath.Join(config.BackupDir(), "secrets.backup.*"))
		fmt.Printf("  Available backups: %d\n", len(files))
		if len(files) > 0 {
			type backupInfo struct {
				path string
				time time.Time
			}
			entries := []backupInfo{}
			for _, file := range files {
				info, err := os.Stat(file)
				if err != nil {
					continue
				}
				entries = append(entries, backupInfo{path: file, time: info.ModTime()})
			}
			sort.Slice(entries, func(i, j int) bool {
				return entries[i].time.After(entries[j].time)
			})
			for i := 0; i < len(entries) && i < 5; i++ {
				fmt.Printf("    - %s (%s)\n", filepath.Base(entries[i].path), entries[i].time.Format("2006-01-02 15:04:05"))
			}
		}
	} else {
		fmt.Println("  No backups yet")
	}
	return 0
}

func runInit(args []string) int {
	encrypted := false
	encryptExisting := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--encrypted":
			encrypted = true
		case "--encrypt-existing":
			encryptExisting = true
		case "--help":
			fmt.Println("Usage: env-sync init [options]")
			fmt.Println("Options:")
			fmt.Println("  --encrypted         Initialize with encryption enabled (generates AGE key)")
			fmt.Println("  --encrypt-existing  Encrypt an existing plaintext secrets file")
			fmt.Println("  --help              Show this help")
			return 0
		default:
			logging.Log("ERROR", fmt.Sprintf("Unknown option: %s", args[i]))
			return 1
		}
	}

	if encrypted || encryptExisting {
		if err := keys.CheckAgeInstalled(); err != nil {
			logging.Log("ERROR", "AGE is required for encryption. Install it first.")
			return 1
		}
	}

	if encryptExisting {
		return handleEncryptExisting()
	}

	if _, err := os.Stat(config.SecretsFile()); err == nil {
		logging.Log("WARN", "Secrets file already exists: "+config.SecretsFile())
		fmt.Print("Overwrite? [y/N] ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(answer)
		if !strings.HasPrefix(strings.ToLower(answer), "y") {
			logging.Log("INFO", "Cancelled")
			return 0
		}
		_ = backup.CreateBackup(config.SecretsFile())
	}

	if encrypted {
		_ = keys.GenerateKey()
	}

	if err := secrets.InitSecretsFile(config.SecretsFile(), ""); err != nil {
		return 1
	}

	if encrypted {
		pubkey := keys.GetLocalPubkey()
		if err := metadata.EnsureEncryptedMetadata(config.SecretsFile(), secrets.GetHostname()); err != nil {
			return 1
		}
		publicKeys := keys.GetAllKnownPublicKeys()
		if err := metadata.EnsurePublicKeysMetadata(config.SecretsFile(), publicKeys); err != nil {
			return 1
		}
		if err := secrets.UpdateMetadata(config.SecretsFile(), ""); err != nil {
			return 1
		}
		logging.Log("SUCCESS", "Initialized encrypted secrets file at "+config.SecretsFile())
		logging.Log("INFO", "Your AGE public key: "+pubkey)
		logging.Log("INFO", "Edit the file to add your secrets (use 'env-sync add'), then run 'env-sync'")
		return 0
	}

	logging.Log("SUCCESS", "Initialized secrets file at "+config.SecretsFile())
	logging.Log("INFO", "Edit the file to add your secrets, then run 'env-sync'")
	return 0
}

func handleEncryptExisting() int {
	if _, err := os.Stat(config.SecretsFile()); err != nil {
		logging.Log("ERROR", "No existing secrets file to encrypt: "+config.SecretsFile())
		return 1
	}

	if keys.IsFileEncrypted(config.SecretsFile()) {
		logging.Log("INFO", "Secrets file is already encrypted")
		return 0
	}

	if _, err := os.Stat(config.AgeKeyFile()); err != nil {
		_ = keys.GenerateKey()
	}

	content, err := secrets.GetSecretsContent(config.SecretsFile())
	if err != nil {
		return 1
	}

	timestamp := secrets.GetTimestamp()
	pubkey := keys.GetLocalPubkey()
	var newLines []string
	linePattern := regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)=(.*)`)

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			newLines = append(newLines, line)
			continue
		}
		if matches := linePattern.FindStringSubmatch(line); len(matches) > 0 {
			key := matches[1]
			val := strings.TrimSpace(matches[2])
			val = strings.Trim(val, "\"")
			val = strings.Trim(val, "'")
			encrypted, err := keys.EncryptValue(val, []string{pubkey})
			if err != nil {
				return 1
			}
			newLines = append(newLines, fmt.Sprintf("%s=\"%s\" # ENVSYNC_UPDATED_AT=%s", key, encrypted, timestamp))
		} else {
			newLines = append(newLines, line)
		}
	}

	_ = backup.CreateBackup(config.SecretsFile())
	if err := secrets.SetSecretsContent(config.SecretsFile(), strings.TrimSpace(strings.Join(newLines, "\n"))); err != nil {
		return 1
	}

	if err := metadata.EnsureEncryptedMetadata(config.SecretsFile(), secrets.GetHostname()); err != nil {
		return 1
	}
	publicKeys := keys.GetAllKnownPublicKeys()
	if err := metadata.EnsurePublicKeysMetadata(config.SecretsFile(), publicKeys); err != nil {
		return 1
	}

	if err := secrets.UpdateMetadata(config.SecretsFile(), ""); err != nil {
		return 1
	}

	logging.Log("SUCCESS", "Encrypted existing secrets file")
	logging.Log("INFO", "Run 'env-sync' to sync with peers and add them as recipients")
	return 0
}

func runRestore(args []string) int {
	backupNum := 1
	if len(args) > 0 {
		value, err := strconv.Atoi(args[0])
		if err != nil {
			logging.Log("ERROR", fmt.Sprintf("Invalid backup number: %s (must be 1-5)", args[0]))
			return 1
		}
		backupNum = value
	}
	if backupNum < 1 || backupNum > 5 {
		logging.Log("ERROR", fmt.Sprintf("Invalid backup number: %d (must be 1-5)", backupNum))
		return 1
	}
	if err := backup.RestoreBackup(backupNum, config.SecretsFile()); err != nil {
		return 1
	}
	return 0
}

func runCron(args []string) int {
	action := "show"
	for _, arg := range args {
		switch arg {
		case "--install":
			action = "install"
		case "--remove":
			action = "remove"
		case "--show":
			action = "show"
		default:
			logging.Log("ERROR", fmt.Sprintf("Unknown option: %s", arg))
			return 1
		}
	}

	exe, err := os.Executable()
	if err != nil {
		return 1
	}

	switch action {
	case "install":
		if err := cron.Install(exe); err != nil {
			return 1
		}
	case "remove":
		if err := cron.Remove(); err != nil {
			return 1
		}
	case "show":
		if err := cron.Show(); err != nil {
			return 1
		}
	}
	return 0
}

func runAdd(args []string) int {
	if len(args) < 1 {
		fmt.Println("Usage: env-sync add KEY=\"value\"")
		fmt.Println("Example: env-sync add OPENAI_API_KEY=\"sk-...\"")
		return 1
	}

	input := args[0]
	if !regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`).MatchString(input) {
		logging.Log("ERROR", "Invalid format. Use: KEY=\"value\" or KEY=value")
		return 1
	}

	key := strings.SplitN(input, "=", 2)[0]
	value := strings.SplitN(input, "=", 2)[1]
	value = strings.Trim(value, "\"")
	value = strings.Trim(value, "'")

	if _, err := os.Stat(config.SecretsFile()); err != nil {
		logging.Log("ERROR", "Secrets file not found. Run 'env-sync init' first.")
		return 1
	}

	timestamp := secrets.GetTimestamp()
	finalValue := value

	if keys.IsFileEncrypted(config.SecretsFile()) {
		if !keys.CanDecryptFile(config.SecretsFile()) {
			logging.Log("ERROR", "Cannot modify encrypted file - you don't have access")
			return 1
		}
		recipients := keys.GetRecipientsFromFile(config.SecretsFile())
		encrypted, err := keys.EncryptValue(value, recipients)
		if err != nil {
			logging.Log("ERROR", "Failed to encrypt value")
			return 1
		}
		finalValue = encrypted
	}

	_ = backup.CreateBackup(config.SecretsFile())
	content, _ := secrets.GetSecretsContent(config.SecretsFile())
	lines := []string{}
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, key+"=") || strings.HasPrefix(line, "export "+key+"=") {
			continue
		}
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}

	lines = append(lines, fmt.Sprintf("%s=\"%s\" # ENVSYNC_UPDATED_AT=%s", key, finalValue, timestamp))

	if err := secrets.SetSecretsContent(config.SecretsFile(), strings.Join(lines, "\n")); err != nil {
		return 1
	}
	logging.Log("SUCCESS", "Added/updated key: "+key)
	return 0
}

func runRemove(args []string) int {
	if len(args) < 1 {
		fmt.Println("Usage: env-sync remove KEY")
		fmt.Println("Example: env-sync remove OPENAI_API_KEY")
		return 1
	}
	key := args[0]
	if !regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`).MatchString(key) {
		logging.Log("ERROR", "Invalid key name: "+key)
		return 1
	}

	if _, err := os.Stat(config.SecretsFile()); err != nil {
		logging.Log("ERROR", "Secrets file not found. Run 'env-sync init' first.")
		return 1
	}

	if keys.IsFileEncrypted(config.SecretsFile()) && !keys.CanDecryptFile(config.SecretsFile()) {
		logging.Log("ERROR", "Cannot modify encrypted file - you don't have access")
		return 1
	}

	content, _ := secrets.GetSecretsContent(config.SecretsFile())
	if !strings.Contains(content, key+"=") {
		logging.Log("WARN", "Key not found: "+key)
		return 0
	}

	_ = backup.CreateBackup(config.SecretsFile())
	lines := []string{}
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, key+"=") || strings.HasPrefix(line, "export "+key+"=") {
			continue
		}
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}

	if err := secrets.SetSecretsContent(config.SecretsFile(), strings.Join(lines, "\n")); err != nil {
		return 1
	}
	logging.Log("SUCCESS", "Removed key: "+key)
	return 0
}

func runShow(args []string) int {
	if len(args) < 1 {
		fmt.Println("Usage: env-sync show KEY")
		fmt.Println("Example: env-sync show OPENAI_API_KEY")
		return 1
	}
	key := args[0]
	if _, err := os.Stat(config.SecretsFile()); err != nil {
		logging.Log("ERROR", "Secrets file not found. Run 'env-sync init' first.")
		return 1
	}
	content, _ := secrets.GetSecretsContent(config.SecretsFile())
	var line string
	for _, entry := range strings.Split(content, "\n") {
		if strings.HasPrefix(entry, key+"=") {
			line = entry
			break
		}
	}
	if line == "" {
		logging.Log("ERROR", "Key not found: "+key)
		return 1
	}

	trimmed := strings.TrimPrefix(line, key+"=")
	if strings.HasPrefix(trimmed, "\"") {
		trimmed = strings.TrimPrefix(trimmed, "\"")
		if end := strings.Index(trimmed, "\""); end >= 0 {
			val := trimmed[:end]
			if keys.IsFileEncrypted(config.SecretsFile()) {
				decrypted, err := keys.DecryptValue(val)
				if err != nil {
					logging.Log("ERROR", "Failed to decrypt value")
					return 1
				}
				fmt.Println(decrypted)
				return 0
			}
			fmt.Println(val)
			return 0
		}
	}
	val := strings.SplitN(trimmed, "#", 2)[0]
	fmt.Println(strings.TrimSpace(val))
	return 0
}

func runList(_ []string) int {
	if _, err := os.Stat(config.SecretsFile()); err != nil {
		logging.Log("ERROR", "Secrets file not found. Run 'env-sync init' first.")
		return 1
	}
	content, _ := secrets.GetSecretsContent(config.SecretsFile())
	keysSet := map[string]bool{}
	for _, line := range strings.Split(content, "\n") {
		if !regexp.MustCompile(`^[A-Za-z_]`).MatchString(line) {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) > 0 {
			keysSet[parts[0]] = true
		}
	}
	if len(keysSet) == 0 {
		fmt.Println("No secrets found in file.")
		return 0
	}
	fmt.Println("Secrets keys:")
	sortedKeys := make([]string, 0, len(keysSet))
	for key := range keysSet {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)
	for _, key := range sortedKeys {
		fmt.Printf("  • %s\n", key)
	}
	fmt.Println("")
	fmt.Printf("Total: %d keys\n", len(sortedKeys))
	return 0
}

func runLoad(args []string) int {
	format := "env"
	keyFilter := ""
	decryptOnly := false

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--format":
			if i+1 < len(args) {
				format = args[i+1]
				i++
			}
		case "--key":
			if i+1 < len(args) {
				keyFilter = args[i+1]
				i++
			}
		case "--decrypt-only":
			decryptOnly = true
		case "--quiet":
			_ = os.Setenv("ENV_SYNC_QUIET", "true")
		case "--help":
			fmt.Println("Usage: env-sync load [options]")
			fmt.Println("")
			fmt.Println("Load and decrypt secrets for shell integration.")
			fmt.Println("Outputs export statements suitable for eval.")
			fmt.Println("")
			fmt.Println("Options:")
			fmt.Println("  --format <env|json>    Output format (default: env)")
			fmt.Println("  --key <name>           Load only specific key")
			fmt.Println("  --decrypt-only         Output decrypted content without parsing")
			fmt.Println("  --quiet                Suppress output")
			fmt.Println("  --help                 Show this help")
			fmt.Println("")
			fmt.Println("Examples:")
			fmt.Println("  eval \"$(env-sync load)\"                    # Load all secrets")
			fmt.Println("  eval \"$(env-sync load --quiet)\"            # Silent load")
			fmt.Println("  env-sync load --format json                # Output as JSON")
			fmt.Println("  env-sync load --key OPENAI_API_KEY         # Load specific key")
			fmt.Println("")
			fmt.Println("Shell Integration:")
			fmt.Println("  Add this to your ~/.bashrc or ~/.zshrc:")
			fmt.Println("  eval \"$(env-sync load 2>/dev/null)\"")
			return 0
		default:
			logging.Log("ERROR", fmt.Sprintf("Unknown option: %s", args[i]))
			return 1
		}
	}

	if _, err := os.Stat(config.SecretsFile()); err != nil {
		if !strings.EqualFold(os.Getenv("ENV_SYNC_QUIET"), "true") {
			logging.Log("ERROR", "Secrets file not found: "+config.SecretsFile())
		}
		return 1
	}

	contentFile := config.SecretsFile()
	tempFile := ""
	if keys.IsFileEncrypted(config.SecretsFile()) {
		if !keys.CanDecryptFile(config.SecretsFile()) {
			if !strings.EqualFold(os.Getenv("ENV_SYNC_QUIET"), "true") {
				logging.Log("ERROR", "Cannot decrypt secrets file - not in recipient list")
				logging.Log("INFO", "Run 'env-sync key request-access --trigger-all' to request access")
			}
			return 1
		}
		tmp, err := os.CreateTemp("", "env-sync-decrypt")
		if err != nil {
			return 1
		}
		tempFile = tmp.Name()
		_ = tmp.Close()
		if err := keys.DecryptSecretsFile(config.SecretsFile(), tempFile); err != nil {
			if !strings.EqualFold(os.Getenv("ENV_SYNC_QUIET"), "true") {
				logging.Log("ERROR", "Failed to decrypt secrets file")
			}
			_ = os.Remove(tempFile)
			return 1
		}
		contentFile = tempFile
	}
	if tempFile != "" {
		defer os.Remove(tempFile)
	}

	data, err := os.ReadFile(contentFile)
	if err != nil {
		return 1
	}

	lines := strings.Split(string(data), "\n")

	switch format {
	case "env":
		if decryptOnly {
			fmt.Print(string(data))
			return 0
		}
		for _, line := range lines {
			if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
				continue
			}
			if matches := regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`).MatchString(line); !matches {
				continue
			}
			key := strings.SplitN(line, "=", 2)[0]
			if keyFilter != "" && key != keyFilter {
				continue
			}
			fmt.Println("export " + line)
		}
	case "json":
		output := map[string]string{}
		for _, line := range lines {
			if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
				continue
			}
			if matches := regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`).MatchString(line); !matches {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			key := parts[0]
			value := strings.Trim(parts[1], "\"")
			value = strings.Trim(value, "'")
			if keyFilter != "" && key != keyFilter {
				continue
			}
			output[key] = value
		}
		payload, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(payload))
	default:
		logging.Log("ERROR", "Unknown format: "+format)
		return 1
	}

	return 0
}

func runPath(args []string) int {
	showBackup := false
	for _, arg := range args {
		switch arg {
		case "--backup":
			showBackup = true
		case "--help":
			fmt.Println("Usage: env-sync path [options]")
			fmt.Println("")
			fmt.Println("Show paths to env-sync files and directories.")
			fmt.Println("")
			fmt.Println("Options:")
			fmt.Println("  --backup    Show backup directory path")
			fmt.Println("")
			fmt.Println("Examples:")
			fmt.Println("  env-sync path           # Show secrets file path")
			fmt.Println("  env-sync path --backup  # Show backup directory path")
			fmt.Println("")
			fmt.Println("Usage in scripts:")
			fmt.Println("  scp host:$(env-sync path) /tmp/  # Copy secrets file")
			return 0
		default:
			logging.Log("ERROR", fmt.Sprintf("Unknown option: %s", arg))
			return 1
		}
	}

	if showBackup {
		fmt.Println(config.BackupDir())
	} else {
		fmt.Println(config.SecretsFile())
	}
	return 0
}

func runKey(args []string) int {
	if len(args) == 0 {
		showKeyHelp()
		return 0
	}

	switch args[0] {
	case "show", "s":
		return runKeyShow(args[1:])
	case "export", "e":
		return runKeyExport(args[1:])
	case "import", "i":
		return runKeyImport(args[1:])
	case "list", "ls", "l":
		return runKeyList(args[1:])
	case "request-access", "request", "req":
		return runKeyRequestAccess(args[1:])
	case "grant-access", "grant":
		return runKeyGrantAccess(args[1:])
	case "approve-requests", "approve":
		return runKeyApprove(args[1:])
	case "remove", "rm", "r":
		return runKeyRemove(args[1:], false)
	case "revoke":
		return runKeyRemove(args[1:], true)
	case "help", "--help", "-h":
		showKeyHelp()
		return 0
	default:
		logging.Log("ERROR", "Unknown subcommand: "+args[0])
		showKeyHelp()
		return 1
	}
}

func showKeyHelp() {
	fmt.Print(`env-sync key - Manage AGE encryption keys

Usage: env-sync key [subcommand] [options]

Subcommands:
  show [options]                 Show this machine's key
    Options:
      --private                  ⚠️ Show private key (be careful!)

  export [options]               Export public key
    Options:
      --qr                       Export as QR code (if qrencode available)

  import <pubkey>                Import a peer's public key
  import --from <hostname>       Import from a peer via SSH

  list                           List all cached public keys
  list --local                   Show only this machine's key

  request-access [options]       Request access for new machine
    Options:
      --trigger <hostname>       SSH into host and trigger sync (immediate)
      --trigger-all              Trigger sync on all online hosts
      --from <hostname>          Request from specific host (approval required)
      --all                      Request from all discovered hosts

  grant-access [options]         Grant access to a requesting machine
    Options:
      --to <hostname>            Hostname of machine to grant access
      --pubkey <key>             Public key of machine to grant access

  approve-requests               Interactively approve pending requests

  remove <hostname>              Remove a peer's public key
  revoke <hostname>              Remove key and re-encrypt without them

Examples:
  env-sync key show                        # Show your public key
  env-sync key export                      # Export your pubkey
  env-sync key import beelink age1xyz...   # Import beelink's pubkey
  env-sync key request-access --trigger-all  # Request immediate access
  env-sync key list                        # List known pubkeys

`)
}

func runKeyShow(args []string) int {
	showPrivate := false
	for _, arg := range args {
		switch arg {
		case "--private":
			showPrivate = true
		case "--help":
			fmt.Println("Usage: env-sync key show [options]")
			fmt.Println("Options:")
			fmt.Println("  --private    Show private key (be careful!)")
			return 0
		default:
			logging.Log("ERROR", fmt.Sprintf("Unknown option: %s", arg))
			return 1
		}
	}

	if showPrivate {
		data, err := os.ReadFile(config.AgeKeyFile())
		if err != nil {
			logging.Log("ERROR", "No private key found at "+config.AgeKeyFile())
			return 1
		}
		logging.Log("WARN", "⚠️  PRIVATE KEY - Keep this secret!")
		fmt.Print(string(data))
		return 0
	}

	pubkey := keys.GetLocalPubkey()
	if pubkey == "" {
		logging.Log("INFO", "No public key found. Generate one with: env-sync init --encrypted")
		return 1
	}
	fmt.Println("Public Key: " + pubkey)
	fmt.Println("Hostname: " + secrets.GetHostname())
	fmt.Println("")
	fmt.Println("To share with peers:")
	fmt.Println("  env-sync key import --pubkey \"" + pubkey + "\" " + secrets.GetHostname())
	return 0
}

func runKeyExport(args []string) int {
	useQR := false
	for _, arg := range args {
		switch arg {
		case "--qr":
			useQR = true
		case "--help":
			fmt.Println("Usage: env-sync key export [options]")
			fmt.Println("Options:")
			fmt.Println("  --qr    Display as QR code")
			return 0
		default:
			logging.Log("ERROR", fmt.Sprintf("Unknown option: %s", arg))
			return 1
		}
	}

	pubkey := keys.GetLocalPubkey()
	if pubkey == "" {
		logging.Log("ERROR", "No public key found. Generate one with: env-sync init --encrypted")
		return 1
	}

	if useQR {
		if _, err := exec.LookPath("qrencode"); err != nil {
			logging.Log("ERROR", "qrencode not installed. Install with: brew install qrencode")
			return 1
		}
		args := []string{"qrencode", "-t", "ANSIUTF8"}
		logging.LogCommand(args...)
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdin = strings.NewReader(pubkey)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Run()
		return 0
	}

	fmt.Println(pubkey)
	return 0
}

func runKeyImport(args []string) int {
	fromHost := ""
	pubkey := ""
	hostname := ""

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--from":
			if i+1 < len(args) {
				fromHost = args[i+1]
				i++
			}
		case "--help":
			fmt.Println("Usage: env-sync key import [options] <pubkey> [hostname]")
			fmt.Println("Options:")
			fmt.Println("  --from <hostname>    Import from peer via SSH")
			fmt.Println("")
			fmt.Println("Examples:")
			fmt.Println("  env-sync key import age1xyz... beelink.local")
			fmt.Println("  env-sync key import --from beelink.local")
			return 0
		default:
			if strings.HasPrefix(args[i], "-") {
				logging.Log("ERROR", fmt.Sprintf("Unknown option: %s", args[i]))
				return 1
			}
			if pubkey == "" {
				pubkey = args[i]
			} else if hostname == "" {
				hostname = args[i]
			}
		}
	}

	if fromHost != "" {
		logging.Log("INFO", "Fetching public key from "+fromHost+"...")
		args := []string{
			"ssh",
			"-o", "ConnectTimeout=5",
			"-o", "StrictHostKeyChecking=" + sshtransport.HostKeyCheckingMode(),
			fromHost,
			"cat ~/.config/env-sync/keys/age_key.pub",
		}
		logging.LogCommand(args...)
		cmd := exec.Command(args[0], args[1:]...)
		output, err := cmd.Output()
		if err != nil {
			logging.Log("ERROR", "Failed to fetch pubkey from "+fromHost)
			return 1
		}
		pubkey = strings.TrimSpace(string(output))
		hostname = strings.TrimSuffix(fromHost, ".local")
	}

	if pubkey == "" {
		logging.Log("ERROR", "No public key provided")
		return 1
	}
	if hostname == "" {
		logging.Log("ERROR", "Hostname required. Usage: env-sync key import <pubkey> <hostname>")
		return 1
	}

	if !keys.ValidatePubkey(pubkey) {
		logging.Log("ERROR", "Invalid AGE public key format")
		return 1
	}

	if err := keys.CachePeerPubkey(hostname, pubkey); err != nil {
		return 1
	}
	logging.Log("SUCCESS", "Imported public key for "+hostname)
	syncer.ReencryptLocal()
	logging.Log("INFO", "Run 'env-sync' to sync secrets with peers")
	return 0
}

func runKeyList(args []string) int {
	localOnly := false
	for _, arg := range args {
		switch arg {
		case "--local":
			localOnly = true
		case "--help":
			fmt.Println("Usage: env-sync key list [options]")
			fmt.Println("Options:")
			fmt.Println("  --local    Show only this machine's key")
			return 0
		default:
			logging.Log("ERROR", fmt.Sprintf("Unknown option: %s", arg))
			return 1
		}
	}

	pubkey := keys.GetLocalPubkey()
	if pubkey != "" {
		fmt.Println("Local Key (" + secrets.GetHostname() + "):")
		fmt.Println("  " + pubkey)
		fmt.Println("")
	}

	if !localOnly {
		fmt.Println("Known Peers:")
		entries := keys.ListCachedPeers()
		if len(entries) == 0 {
			fmt.Println("  (none)")
			return 0
		}
		for _, entry := range entries {
			fmt.Println("  " + entry)
		}
	}
	return 0
}

func runKeyRequestAccess(args []string) int {
	triggerHost := ""
	triggerAll := false
	fromHosts := []string{}
	requestAll := false

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--trigger":
			if i+1 < len(args) {
				triggerHost = args[i+1]
				i++
			}
		case "--trigger-all":
			triggerAll = true
		case "--from":
			if i+1 < len(args) {
				fromHosts = append(fromHosts, args[i+1])
				i++
			}
		case "--all":
			requestAll = true
		case "--help":
			fmt.Println("Usage: env-sync key request-access [options]")
			fmt.Println("Options:")
			fmt.Println("  --trigger <hostname>    SSH into host and trigger re-encryption")
			fmt.Println("  --trigger-all           Trigger on all online hosts")
			fmt.Println("  --from <hostname>       Request from specific host (needs approval)")
			fmt.Println("  --all                   Request from all discovered hosts")
			fmt.Println("")
			fmt.Println("Examples:")
			fmt.Println("  env-sync key request-access --trigger beelink.local")
			fmt.Println("  env-sync key request-access --trigger-all")
			return 0
		default:
			logging.Log("ERROR", fmt.Sprintf("Unknown option: %s", args[i]))
			return 1
		}
	}

	localPubkey := keys.GetLocalPubkey()
	if localPubkey == "" {
		logging.Log("ERROR", "No local key found. Generate with: env-sync init --encrypted")
		return 1
	}

	if triggerHost != "" || triggerAll {
		hosts := []string{}
		if triggerAll {
			logging.Log("INFO", "Discovering peers to trigger...")
			peers, err := discovery.Discover(discovery.Options{Timeout: 5 * time.Second, Quiet: true})
			if err == nil {
				hosts = append(hosts, peers...)
			}
		} else {
			hosts = append(hosts, triggerHost)
		}

		if len(hosts) == 0 {
			logging.Log("WARN", "No hosts found to trigger")
			return 1
		}

		triggered := 0
		for _, host := range hosts {
			logging.Log("INFO", "Triggering re-encryption on "+host+"...")
			args := []string{
				"ssh",
				"-o", "ConnectTimeout=5",
				"-o", "StrictHostKeyChecking=" + sshtransport.HostKeyCheckingMode(),
				host,
				"bash", "-c",
				"mkdir -p ~/.config/env-sync/keys/known_hosts && printf %s \"$1\" > ~/.config/env-sync/keys/known_hosts/$2.pub && env-sync 2>/dev/null || true && echo 'SUCCESS'",
				"bash", localPubkey, secrets.GetHostname(),
			}
			logging.LogCommand(args...)
			cmd := exec.Command(args[0], args[1:]...)
			output, _ := cmd.Output()
			if strings.Contains(string(output), "SUCCESS") {
				logging.Log("SUCCESS", "Triggered re-encryption on "+host)
				triggered++
			} else {
				logging.Log("WARN", "Failed to trigger "+host)
			}
		}

		if triggered > 0 {
			logging.Log("SUCCESS", fmt.Sprintf("Successfully triggered %d host(s)", triggered))
			logging.Log("INFO", "Run 'env-sync' to download encrypted secrets")
			return 0
		}
		logging.Log("ERROR", "Failed to trigger any hosts")
		return 1
	}

	if len(fromHosts) > 0 || requestAll {
		if requestAll {
			peers, err := discovery.Discover(discovery.Options{Timeout: 5 * time.Second, Quiet: true})
			if err == nil {
				fromHosts = append(fromHosts, peers...)
			}
		}

		requested := 0
		for _, host := range fromHosts {
			logging.Log("INFO", "Sending access request to "+host+"...")
			payload := fmt.Sprintf(`{
  "hostname": "%s",
  "pubkey": "%s",
  "timestamp": "%s"
}
`, secrets.GetHostname(), localPubkey, secrets.GetTimestamp())
			args := []string{
				"ssh",
				"-o", "ConnectTimeout=5",
				"-o", "StrictHostKeyChecking=" + sshtransport.HostKeyCheckingMode(),
				host,
				"bash", "-c",
				"mkdir -p ~/.config/env-sync/requests && cat > ~/.config/env-sync/requests/$1.request && echo 'REQUEST_SENT'",
				"bash", secrets.GetHostname(),
			}
			logging.LogCommand(args...)
			cmd := exec.Command(args[0], args[1:]...)
			cmd.Stdin = strings.NewReader(payload)
			output, _ := cmd.Output()
			if strings.Contains(string(output), "REQUEST_SENT") {
				logging.Log("SUCCESS", "Access request sent to "+host)
				logging.Log("INFO", "An admin must run 'env-sync key approve-requests' on "+host)
				requested++
			} else {
				logging.Log("WARN", "Failed to send request to "+host)
			}
		}

		if requested == 0 {
			logging.Log("ERROR", "Failed to send any access requests")
			return 1
		}
		return 0
	}

	return 0
}

func runKeyGrantAccess(args []string) int {
	targetHost := ""
	pubkey := ""

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--to":
			if i+1 < len(args) {
				targetHost = args[i+1]
				i++
			}
		case "--pubkey":
			if i+1 < len(args) {
				pubkey = args[i+1]
				i++
			}
		case "--help":
			fmt.Println("Usage: env-sync key grant-access [options]")
			fmt.Println("Options:")
			fmt.Println("  --to <hostname>      Hostname of machine to grant access")
			fmt.Println("  --pubkey <key>       Public key of machine")
			return 0
		default:
			logging.Log("ERROR", fmt.Sprintf("Unknown option: %s", args[i]))
			return 1
		}
	}

	if targetHost == "" || pubkey == "" {
		logging.Log("ERROR", "Both --to and --pubkey are required")
		return 1
	}

	if err := keys.CachePeerPubkey(targetHost, pubkey); err != nil {
		return 1
	}
	logging.Log("INFO", "Granting access to "+targetHost+"...")
	_ = syncer.Run(syncer.Options{Quiet: true})
	logging.Log("SUCCESS", "Access granted to "+targetHost)
	logging.Log("INFO", "Secrets re-encrypted with new recipient")
	return 0
}

func runKeyApprove(_ []string) int {
	requestsDir := config.RequestsDir()
	if _, err := os.Stat(requestsDir); err != nil {
		logging.Log("INFO", "No pending requests")
		return 0
	}

	files, _ := filepath.Glob(filepath.Join(requestsDir, "*.request"))
	if len(files) == 0 {
		logging.Log("INFO", "No pending requests")
		return 0
	}

	logging.Log("INFO", "Pending access requests:")
	fmt.Println("")

	for _, file := range files {
		hostname := strings.TrimSuffix(filepath.Base(file), ".request")
		var request struct {
			Pubkey    string `json:"pubkey"`
			Timestamp string `json:"timestamp"`
		}
		if data, err := os.ReadFile(file); err == nil {
			_ = json.Unmarshal(data, &request)
		}

		fmt.Println("  Hostname: " + hostname)
		fmt.Println("  Public Key: " + request.Pubkey)
		fmt.Println("  Requested: " + request.Timestamp)
		fmt.Println("")

		fmt.Printf("Grant access to %s? [Y/n] ", hostname)
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(answer)

		if answer == "" || strings.HasPrefix(strings.ToLower(answer), "y") {
			_ = keys.CachePeerPubkey(hostname, request.Pubkey)
			_ = os.Remove(file)
			logging.Log("SUCCESS", "Access granted to "+hostname)
		} else {
			logging.Log("INFO", "Request denied for "+hostname)
			_ = os.Remove(file)
		}
		fmt.Println("")
	}

	logging.Log("INFO", "Re-encrypting secrets with new recipients...")
	_ = syncer.Run(syncer.Options{Quiet: true})
	return 0
}

func runKeyRemove(args []string, revoke bool) int {
	if len(args) == 0 || args[0] == "--help" {
		fmt.Println("Usage: env-sync key remove <hostname>")
		fmt.Println("       env-sync key revoke <hostname>  (also re-encrypts without them)")
		return 0
	}
	hostname := args[0]
	_ = keys.RemovePeerPubkey(hostname)
	if revoke {
		logging.Log("INFO", "Re-encrypting secrets without "+hostname+"...")
		_ = syncer.Run(syncer.Options{Quiet: true})
		logging.Log("SUCCESS", "Revoked access for "+hostname)
		return 0
	}
	logging.Log("SUCCESS", "Removed public key for "+hostname)
	return 0
}

func fetchHealth(host string) (discovery.HealthResponse, error) {
	return discovery.FetchHealth(host)
}

func truncate(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max]
}

func runServiceManagement(args []string) int {
	if len(args) == 0 {
		showServiceHelp()
		return 0
	}

	switch args[0] {
	case "stop":
		return runServiceStop()
	case "restart":
		return runServiceRestart()
	case "uninstall":
		return runServiceUninstall()
	case "help", "--help", "-h":
		showServiceHelp()
		return 0
	default:
		logging.Log("ERROR", "Unknown subcommand: "+args[0])
		showServiceHelp()
		return 1
	}
}

func runServiceStop() int {
	wasStopped, err := svcmgr.StopIfRunning()
	if err != nil {
		logging.Log("ERROR", fmt.Sprintf("Failed to stop service: %v", err))
		return 1
	}
	if wasStopped {
		logging.Log("SUCCESS", "Service stopped")
	} else {
		logging.Log("INFO", "Service is not running")
	}
	return 0
}

func runServiceRestart() int {
	_, err := svcmgr.StopIfRunning()
	if err != nil {
		logging.Log("ERROR", fmt.Sprintf("Failed to stop service: %v", err))
		return 1
	}
	if err := svcmgr.RestartIfNeeded(true); err != nil {
		logging.Log("ERROR", fmt.Sprintf("Failed to restart service: %v", err))
		return 1
	}
	return 0
}

func runServiceUninstall() int {
	wasUninstalled, err := svcmgr.UninstallIfInstalled()
	if err != nil {
		logging.Log("ERROR", fmt.Sprintf("Failed to uninstall service: %v", err))
		return 1
	}
	if wasUninstalled {
		logging.Log("SUCCESS", "Service uninstalled")
	} else {
		logging.Log("INFO", "Service is not installed")
	}
	return 0
}

func showServiceHelp() {
	fmt.Print(`env-sync service - Manage env-sync background service

Usage: env-sync service [subcommand]

Subcommands:
  stop         Stop the running service
  restart      Restart the service
  uninstall    Uninstall the service completely

Note: The service is created when you run 'env-sync serve -d'
      Use 'systemctl --user status env-sync' (Linux) or
      'launchctl print gui/$(id -u)/env-sync' (macOS) to check status.
`)
}
