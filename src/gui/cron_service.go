package main

import (
	"envsync/internal/cron"
	"fmt"
	"os/exec"
	"strings"
)

// CronService provides cron job management
type CronService struct{}

// CronInfo represents cron job status
type CronInfo struct {
	Installed bool `json:"installed"`
	Interval  int  `json:"interval"`
}

// GetCronStatus returns current cron configuration
func (c *CronService) GetCronStatus() (CronInfo, error) {
	info := CronInfo{}

	cmd := exec.Command("crontab", "-l")
	output, err := cmd.Output()
	if err != nil {
		return info, nil
	}

	for _, line := range strings.Split(string(output), "\n") {
		if strings.Contains(line, "env-sync") {
			info.Installed = true
			info.Interval = parseCronInterval(line)
			break
		}
	}

	return info, nil
}

// InstallCron installs a periodic sync cron job
func (c *CronService) InstallCron(intervalMinutes int) error {
	if intervalMinutes <= 0 {
		intervalMinutes = 30
	}

	execPath, err := exec.LookPath("env-sync")
	if err != nil {
		return fmt.Errorf("env-sync binary not found in PATH")
	}

	return cron.Install(execPath, intervalMinutes)
}

// RemoveCron removes the periodic sync cron job
func (c *CronService) RemoveCron() error {
	return cron.Remove()
}

func parseCronInterval(line string) int {
	if len(line) > 2 && line[0] == '*' && line[1] == '/' {
		n := 0
		for i := 2; i < len(line) && line[i] >= '0' && line[i] <= '9'; i++ {
			n = n*10 + int(line[i]-'0')
		}
		if n > 0 {
			return n
		}
	}
	return 30
}
