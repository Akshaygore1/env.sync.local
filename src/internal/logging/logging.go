package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"envsync/internal/config"
)

const (
	red    = "\033[0;31m"
	green  = "\033[0;32m"
	yellow = "\033[1;33m"
	blue   = "\033[0;34m"
	reset  = "\033[0m"
)

func Log(level string, message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	if err := os.MkdirAll(config.LogDir(), 0o700); err == nil {
		logFile := filepath.Join(config.LogDir(), "env-sync.log")
		if file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600); err == nil {
			_, _ = fmt.Fprintf(file, "[%s] [%s] %s\n", timestamp, level, message)
			_ = file.Close()
		}
	}

	if strings.EqualFold(os.Getenv("ENV_SYNC_QUIET"), "true") {
		return
	}

	switch level {
	case "ERROR":
		fmt.Fprintf(os.Stderr, "%sERROR:%s %s\n", red, reset, message)
	case "WARN":
		fmt.Fprintf(os.Stderr, "%sWARN:%s %s\n", yellow, reset, message)
	case "INFO":
		fmt.Fprintf(os.Stderr, "%sINFO:%s %s\n", blue, reset, message)
	case "SUCCESS":
		fmt.Fprintf(os.Stderr, "%sSUCCESS:%s %s\n", green, reset, message)
	default:
		fmt.Fprintf(os.Stderr, "%s\n", message)
	}
}
