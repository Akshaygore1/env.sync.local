package cron

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"envsync/internal/logging"
)

func Install(executable string) error {
	cronLine := fmt.Sprintf("*/30 * * * * %s --quiet sync >/dev/null 2>&1", executable)
	existing := readCrontab()
	filtered := filterEnvSync(existing)
	filtered = append(filtered, cronLine)
	if err := writeCrontab(filtered); err != nil {
		return err
	}
	logging.Log("SUCCESS", "Installed cron job for 30-minute sync")
	logging.Log("INFO", "Next sync will happen automatically")
	return nil
}

func Remove() error {
	existing := readCrontab()
	filtered := filterEnvSync(existing)
	if len(filtered) == len(existing) {
		logging.Log("INFO", "No cron job found")
		return nil
	}
	if err := writeCrontab(filtered); err != nil {
		return err
	}
	logging.Log("SUCCESS", "Removed cron job")
	return nil
}

func Show() error {
	existing := readCrontab()
	filtered := filterEnvSync(existing)
	if len(filtered) == len(existing) {
		logging.Log("INFO", "No cron job installed")
		fmt.Println("Run 'env-sync cron --install' to set up periodic sync")
		return nil
	}
	fmt.Println("Current cron job:")
	for _, line := range existing {
		if strings.Contains(line, "env-sync") {
			fmt.Println(line)
		}
	}
	return nil
}

func readCrontab() []string {
	cmd := exec.Command("crontab", "-l")
	output, err := cmd.Output()
	if err != nil {
		return []string{}
	}
	lines := strings.Split(strings.TrimSuffix(string(output), "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return []string{}
	}
	return lines
}

func writeCrontab(lines []string) error {
	cmd := exec.Command("crontab", "-")
	input := bytes.NewBufferString(strings.Join(lines, "\n") + "\n")
	cmd.Stdin = input
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func filterEnvSync(lines []string) []string {
	filtered := []string{}
	for _, line := range lines {
		if strings.Contains(line, "env-sync") {
			continue
		}
		filtered = append(filtered, line)
	}
	return filtered
}
