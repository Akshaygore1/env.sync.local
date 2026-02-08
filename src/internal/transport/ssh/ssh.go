package sshtransport

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"envsync/internal/secrets"
)

func TestSSH(host string) error {
	cmd := exec.Command("ssh", "-n", "-o", "BatchMode=yes", "-o", "ConnectTimeout=3", "-o", "StrictHostKeyChecking="+HostKeyCheckingMode(), host, "echo", "OK")
	return cmd.Run()
}

func FetchViaSCP(host string, remotePath string, dest string) error {
	remote := fmt.Sprintf("%s:%s", host, remotePath)
	cmd := exec.Command("scp", "-o", "BatchMode=yes", "-o", "ConnectTimeout=5", "-o", "StrictHostKeyChecking="+HostKeyCheckingMode(), remote, dest)
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
