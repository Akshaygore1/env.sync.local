//go:build gui

package main

import (
	"envsync/internal/backup"
	"envsync/internal/config"
	"os"
	"path/filepath"
	"time"
)

// BackupService provides backup management
type BackupService struct{}

// ListBackups returns all available backups
func (b *BackupService) ListBackups() ([]BackupEntry, error) {
	backupDir := config.BackupDir()
	var entries []BackupEntry

	files, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return entries, nil
		}
		return nil, err
	}

	for i, f := range files {
		if f.IsDir() {
			continue
		}
		info, err := f.Info()
		if err != nil {
			continue
		}
		entries = append(entries, BackupEntry{
			Number:    i + 1,
			Timestamp: info.ModTime().Format(time.RFC3339),
			Size:      info.Size(),
			Path:      filepath.Join(backupDir, f.Name()),
		})
	}

	return entries, nil
}

// Restore restores a backup by number
func (b *BackupService) Restore(n int) error {
	return backup.RestoreBackup(n, config.SecretsFile())
}

// GetBackupDir returns the backup directory path
func (b *BackupService) GetBackupDir() string {
	return config.BackupDir()
}
