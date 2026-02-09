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

// GetRemoteSecretsPath fetches the secrets file path from the remote host using env-sync path
func GetRemoteSecretsPath(host string) (string, error) {
	args := []string{"ssh", "-n", "-o", "BatchMode=yes", "-o", "ConnectTimeout=3", "-o", "StrictHostKeyChecking=" + HostKeyCheckingMode(), host, "env-sync", "path"}
	logging.LogCommand(args...)
	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get remote path: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func FetchViaSCP(host string, remotePath string, dest string) error {
	// Try to get the remote path dynamically using env-sync path
	// This handles OS-specific home directory differences
	dynamicPath, err := GetRemoteSecretsPath(host)
	if err == nil && dynamicPath != "" {
		remotePath = dynamicPath
	}
	// If we can't get the dynamic path, fall back to the passed remotePath

	remote := fmt.Sprintf("%s:%s", host, remotePath)
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
