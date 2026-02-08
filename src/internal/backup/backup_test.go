package backup

import (
	"os"
	"path/filepath"
	"testing"

	"envsync/internal/config"
)

func TestCreateBackupRotation(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	secretsFile := filepath.Join(home, ".secrets.env")
	if err := os.WriteFile(secretsFile, []byte("first"), 0o600); err != nil {
		t.Fatalf("write secrets file: %v", err)
	}

	if err := CreateBackup(secretsFile); err != nil {
		t.Fatalf("CreateBackup: %v", err)
	}

	backup1 := filepath.Join(config.BackupDir(), "secrets.backup.1")
	data, err := os.ReadFile(backup1)
	if err != nil {
		t.Fatalf("read backup1: %v", err)
	}
	if string(data) != "first" {
		t.Fatalf("backup1 content mismatch: %q", string(data))
	}

	if err := os.WriteFile(secretsFile, []byte("second"), 0o600); err != nil {
		t.Fatalf("rewrite secrets file: %v", err)
	}

	if err := CreateBackup(secretsFile); err != nil {
		t.Fatalf("CreateBackup second: %v", err)
	}

	data, err = os.ReadFile(backup1)
	if err != nil {
		t.Fatalf("read backup1 after rotation: %v", err)
	}
	if string(data) != "second" {
		t.Fatalf("backup1 content after rotation mismatch: %q", string(data))
	}

	backup2 := filepath.Join(config.BackupDir(), "secrets.backup.2")
	data, err = os.ReadFile(backup2)
	if err != nil {
		t.Fatalf("read backup2: %v", err)
	}
	if string(data) != "first" {
		t.Fatalf("backup2 content mismatch: %q", string(data))
	}
}
