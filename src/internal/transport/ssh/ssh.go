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

func Hostname() string {
	return strings.TrimSpace(secrets.GetHostname())
}

func HostKeyCheckingMode() string {
	if strings.EqualFold(os.Getenv("ENV_SYNC_STRICT_SSH"), "true") {
		return "yes"
	}
	return "accept-new"
}
