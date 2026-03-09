package main

import (
	"bufio"
	"envsync/internal/config"
	"os"
	"path/filepath"
	"strings"
)

// LogService provides log viewing capabilities
type LogService struct{}

// LogEntry represents a single log line
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

// GetRecentLogs returns recent log entries
func (l *LogService) GetRecentLogs(count int) ([]LogEntry, error) {
	if count <= 0 {
		count = 100
	}

	logDir := config.LogDir()
	files, err := os.ReadDir(logDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	if len(files) == 0 {
		return nil, nil
	}

	latestFile := files[len(files)-1]
	logFile := filepath.Join(logDir, latestFile.Name())

	return readLastNLogLines(logFile, count)
}

// GetLogFile returns the log directory path
func (l *LogService) GetLogFile() string {
	return config.LogDir()
}

func readLastNLogLines(file string, n int) ([]LogEntry, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var allLines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}

	start := 0
	if len(allLines) > n {
		start = len(allLines) - n
	}
	lines := allLines[start:]

	var entries []LogEntry
	for _, line := range lines {
		entries = append(entries, parseLogLine(line))
	}

	return entries, nil
}

func parseLogLine(line string) LogEntry {
	entry := LogEntry{Message: line}

	parts := strings.SplitN(line, " ", 3)
	if len(parts) >= 3 {
		entry.Timestamp = parts[0]
		level := strings.ToUpper(strings.TrimSpace(parts[1]))
		switch level {
		case "ERROR", "WARN", "INFO", "DEBUG", "SUCCESS":
			entry.Level = level
			entry.Message = parts[2]
		default:
			entry.Level = "INFO"
		}
	}

	return entry
}
