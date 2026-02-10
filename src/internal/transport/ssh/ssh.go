package sshtransport

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"envsync/internal/logging"
	"envsync/internal/secrets"
)

func TestSSH(host string) error {
	args := []string{"ssh", "-n", "-o", "BatchMode=yes", "-o", "ConnectTimeout=3", "-o", "StrictHostKeyChecking=" + HostKeyCheckingMode(), host, "echo", "OK"}
	logging.LogCommand(args...)
	cmd := exec.Command(args[0], args[1:]...)
	return cmd.Run()
}

func FetchViaSCP(host string, remotePath string, dest string) error {
	// Always use .secrets.env without path prefix
	// SCP command runs relative to home folder, so this will work regardless of where env-sync is installed
	remote := fmt.Sprintf("%s:.secrets.env", host)
	args := []string{"scp", "-o", "BatchMode=yes", "-o", "ConnectTimeout=5", "-o", "StrictHostKeyChecking=" + HostKeyCheckingMode(), remote, dest}
	logging.LogCommand(args...)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = nil
	return cmd.Run()
}

// RegisterPubkeyWithPeer SSHs into a remote host, stores the local pubkey
// in their known_hosts directory, and triggers re-encryption so the remote
// secrets include this machine as a recipient.
func RegisterPubkeyWithPeer(host string, localPubkey string, localHostname string) error {
	script := `mkdir -p ~/.config/env-sync/keys/known_hosts && printf %s "$1" > ~/.config/env-sync/keys/known_hosts/$2.pub && env-sync 2>/dev/null || true && echo 'ENVSYNC_REGISTER_SUCCESS'`
	args := []string{
		"ssh",
		"-o", "BatchMode=yes",
		"-o", "ConnectTimeout=5",
		"-o", "StrictHostKeyChecking=" + HostKeyCheckingMode(),
		host,
		"bash", "-c", script,
		"bash", localPubkey, localHostname,
	}
	logging.LogCommand(args...)
	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("SSH command failed: %w", err)
	}
	if !strings.Contains(string(output), "ENVSYNC_REGISTER_SUCCESS") {
		return fmt.Errorf("remote registration did not succeed")
	}
	return nil
}

func Hostname() string {
	return strings.TrimSpace(secrets.GetHostname())
}

func HostKeyCheckingMode() string {
	if strings.EqualFold(os.Getenv("ENV_SYNC_STRICT_SSH"), "true") {
		return "yes"
	}
	return "accept-new"
}
