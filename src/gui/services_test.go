package main

import (
	"envsync/internal/config"
	"testing"
)

func TestAppGetVersion(t *testing.T) {
	app := NewApp()
	v := app.GetVersion()
	if v == "" {
		t.Error("GetVersion returned empty string")
	}
}

func TestAppGetConfigPaths(t *testing.T) {
	app := NewApp()
	paths := app.GetConfigPaths()
	if paths.ConfigDir == "" {
		t.Error("ConfigDir is empty")
	}
	if paths.BackupDir == "" {
		t.Error("BackupDir is empty")
	}
	if paths.LogDir == "" {
		t.Error("LogDir is empty")
	}
	if paths.KeysDir == "" {
		t.Error("KeysDir is empty")
	}
	if paths.SecretsFile == "" {
		t.Error("SecretsFile is empty")
	}
}

func TestModeServiceGetMode(t *testing.T) {
	svc := &ModeService{}
	info, err := svc.GetMode()
	if err != nil {
		t.Fatalf("GetMode error: %v", err)
	}
	if info.Current == "" {
		t.Error("Current mode is empty")
	}
	if info.Description == "" {
		t.Error("Mode description is empty")
	}
}

func TestModeServiceGetAvailableModes(t *testing.T) {
	svc := &ModeService{}
	modes := svc.GetAvailableModes()
	if len(modes) == 0 {
		t.Error("No available modes returned")
	}
	if len(modes) < 3 {
		t.Errorf("Expected at least 3 modes, got %d", len(modes))
	}
	for _, m := range modes {
		if m.Current == "" {
			t.Error("Mode has empty current field")
		}
		if m.Description == "" {
			t.Error("Mode has empty description")
		}
		if m.Transport == "" {
			t.Error("Mode has empty transport")
		}
		if m.Encryption == "" {
			t.Error("Mode has empty encryption")
		}
	}
}

func TestStatusServiceGetFileStatus(t *testing.T) {
	svc := &StatusService{}
	status, err := svc.GetFileStatus()
	if err != nil {
		t.Fatalf("GetFileStatus error: %v", err)
	}
	if status.Path == "" {
		t.Error("Path is empty")
	}
}

func TestStatusServiceGetServerStatus(t *testing.T) {
	svc := &StatusService{}
	status, err := svc.GetServerStatus()
	if err != nil {
		t.Fatalf("GetServerStatus error: %v", err)
	}
	if status.Port == "" {
		t.Error("Port is empty")
	}
}

func TestStatusServiceIsServerRunning(t *testing.T) {
	svc := &StatusService{}
	_ = svc.IsServerRunning()
}

func TestCronServiceGetCronStatus(t *testing.T) {
	svc := &CronService{}
	info, err := svc.GetCronStatus()
	if err != nil {
		t.Fatalf("GetCronStatus error: %v", err)
	}
	_ = info.Installed
}

func TestBackupServiceGetBackupDir(t *testing.T) {
	svc := &BackupService{}
	dir := svc.GetBackupDir()
	if dir == "" {
		t.Error("Backup dir is empty")
	}
}

func TestBackupServiceListBackups(t *testing.T) {
	svc := &BackupService{}
	backups, err := svc.ListBackups()
	if err != nil {
		t.Fatalf("ListBackups error: %v", err)
	}
	_ = backups
}

func TestLogServiceGetRecentLogs(t *testing.T) {
	svc := &LogService{}
	logs, _ := svc.GetRecentLogs(50)
	_ = logs
}

func TestModeTransportEncryption(t *testing.T) {
	tests := []struct {
		mode       config.SyncMode
		transport  string
		encryption string
	}{
		{config.ModeDevPlaintextHTTP, "HTTP", "none"},
		{config.ModeTrustedOwnerSSH, "SSH/SCP", "optional AGE"},
		{config.ModeSecurePeer, "HTTPS (mTLS)", "AGE (mandatory)"},
	}

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			transport, encryption := modeTransportEncryption(tt.mode)
			if transport != tt.transport {
				t.Errorf("transport: want %q, got %q", tt.transport, transport)
			}
			if encryption != tt.encryption {
				t.Errorf("encryption: want %q, got %q", tt.encryption, encryption)
			}
		})
	}
}

func TestModeFeatures(t *testing.T) {
	tests := []struct {
		mode     config.SyncMode
		minFeats int
	}{
		{config.ModeDevPlaintextHTTP, 3},
		{config.ModeTrustedOwnerSSH, 3},
		{config.ModeSecurePeer, 3},
	}

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			features := modeFeatures(tt.mode)
			if len(features) < tt.minFeats {
				t.Errorf("expected at least %d features, got %d", tt.minFeats, len(features))
			}
		})
	}
}
