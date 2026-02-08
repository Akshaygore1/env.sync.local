package backup

import (
	"fmt"
	"os"
	"path/filepath"

	"envsync/internal/config"
	"envsync/internal/logging"
)

const maxBackups = 5

func CreateBackup(file string) error {
	if _, err := os.Stat(file); err != nil {
		return nil
	}

	if err := os.MkdirAll(config.BackupDir(), 0o700); err != nil {
		return err
	}

	for i := maxBackups - 1; i >= 1; i-- {
		src := filepath.Join(config.BackupDir(), fmt.Sprintf("secrets.backup.%d", i))
		dst := filepath.Join(config.BackupDir(), fmt.Sprintf("secrets.backup.%d", i+1))
		if _, err := os.Stat(src); err == nil {
			_ = os.Rename(src, dst)
		}
	}

	backupFile := filepath.Join(config.BackupDir(), "secrets.backup.1")
	if err := copyFile(file, backupFile); err != nil {
		return err
	}
	_ = os.Chmod(backupFile, 0o600)
	logging.Log("INFO", "Created backup: secrets.backup.1")
	return nil
}

func RestoreBackup(backupNum int, secretsFile string) error {
	backupFile := filepath.Join(config.BackupDir(), fmt.Sprintf("secrets.backup.%d", backupNum))
	if _, err := os.Stat(backupFile); err != nil {
		logging.Log("ERROR", fmt.Sprintf("Backup %d not found", backupNum))
		return err
	}

	_ = CreateBackup(secretsFile)
	if err := copyFile(backupFile, secretsFile); err != nil {
		return err
	}
	_ = os.Chmod(secretsFile, 0o600)
	logging.Log("SUCCESS", fmt.Sprintf("Restored from backup %d", backupNum))
	return nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.WriteFile(dst, data, 0o600); err != nil {
		return err
	}
	return nil
}
