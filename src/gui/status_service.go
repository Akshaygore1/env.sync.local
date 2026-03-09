package main

import (
	"envsync/internal/config"
	"envsync/internal/discovery"
	"envsync/internal/keys"
	"envsync/internal/metadata"
	"envsync/internal/mode"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// StatusService provides system status information
type StatusService struct{}

// FileStatus describes the secrets file state
type FileStatus struct {
	Exists    bool   `json:"exists"`
	Path      string `json:"path"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
	Host      string `json:"host"`
	Encrypted bool   `json:"encrypted"`
	ModTime   string `json:"modTime"`
}

// ServerStatus describes the background server state
type ServerStatus struct {
	Running bool   `json:"running"`
	Port    string `json:"port"`
	PID     int    `json:"pid"`
}

// PeerStatus describes a discovered peer
type PeerStatus struct {
	Hostname  string `json:"hostname"`
	SSHAccess bool   `json:"sshAccess"`
	HasPubKey bool   `json:"hasPubKey"`
}

// BackupEntry describes a backup file
type BackupEntry struct {
	Number    int    `json:"number"`
	Timestamp string `json:"timestamp"`
	Size      int64  `json:"size"`
	Path      string `json:"path"`
}

// ModeInfo describes the current operating mode
type ModeInfo struct {
	Current     string   `json:"current"`
	Description string   `json:"description"`
	Features    []string `json:"features"`
	Transport   string   `json:"transport"`
	Encryption  string   `json:"encryption"`
}

// StatusInfo aggregates all status data
type StatusInfo struct {
	SecretsFile FileStatus    `json:"secretsFile"`
	Server      ServerStatus  `json:"server"`
	Peers       []PeerStatus  `json:"peers"`
	Backups     []BackupEntry `json:"backups"`
	Mode        ModeInfo      `json:"mode"`
}

// GetStatus returns full system status
func (s *StatusService) GetStatus() (StatusInfo, error) {
	fileStatus, _ := s.GetFileStatus()
	serverStatus, _ := s.GetServerStatus()
	modeInfo := getModeInfo()
	backups := listBackupEntries()

	return StatusInfo{
		SecretsFile: fileStatus,
		Server:      serverStatus,
		Mode:        modeInfo,
		Backups:     backups,
	}, nil
}

// GetFileStatus returns secrets file info
func (s *StatusService) GetFileStatus() (FileStatus, error) {
	file := config.SecretsFile()
	status := FileStatus{Path: file}

	info, err := os.Stat(file)
	if err != nil {
		status.Exists = false
		return status, nil
	}

	status.Exists = true
	status.ModTime = info.ModTime().Format(time.RFC3339)
	status.Version = metadata.GetFileVersion(file)
	status.Timestamp = metadata.GetFileTimestamp(file)
	status.Host = metadata.GetFileHost(file)
	status.Encrypted = keys.IsFileEncrypted(file)

	return status, nil
}

// GetServerStatus returns HTTP server status
func (s *StatusService) GetServerStatus() (ServerStatus, error) {
	status := ServerStatus{Port: config.EnvSyncPort()}

	pidFile := config.ServerPidFile()
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return status, nil
	}

	var pid int
	if _, err := fmt.Sscanf(strings.TrimSpace(string(data)), "%d", &pid); err == nil {
		process, err := os.FindProcess(pid)
		if err == nil && process.Signal(nil) == nil {
			status.Running = true
			status.PID = pid
		}
	}

	return status, nil
}

// IsServerRunning checks if the background server is running
func (s *StatusService) IsServerRunning() bool {
	status, _ := s.GetServerStatus()
	return status.Running
}

// GetFileModTime returns the secrets file modification time for change detection
func (s *StatusService) GetFileModTime() (string, error) {
	info, err := os.Stat(config.SecretsFile())
	if err != nil {
		return "", err
	}
	return info.ModTime().Format(time.RFC3339Nano), nil
}

// DiscoverPeers runs mDNS discovery and returns peer status
func (s *StatusService) DiscoverPeers() ([]PeerStatus, error) {
	peers, err := discovery.Discover(discovery.Options{
		Timeout: 5 * time.Second,
		Quiet:   true,
	})
	if err != nil {
		return nil, err
	}

	var result []PeerStatus
	for _, hostname := range peers {
		result = append(result, PeerStatus{
			Hostname:  hostname,
			HasPubKey: keys.GetCachedPubkey(hostname) != "",
		})
	}
	return result, nil
}

func getModeInfo() ModeInfo {
	currentMode := mode.GetMode()
	transport, encryption := modeTransportEncryption(currentMode)
	return ModeInfo{
		Current:     string(currentMode),
		Description: mode.ModeDescription(currentMode),
		Transport:   transport,
		Encryption:  encryption,
	}
}

func modeTransportEncryption(m config.SyncMode) (string, string) {
	switch m {
	case config.ModeDevPlaintextHTTP:
		return "HTTP", "none"
	case config.ModeTrustedOwnerSSH:
		return "SSH/SCP", "optional AGE"
	case config.ModeSecurePeer:
		return "HTTPS (mTLS)", "AGE (mandatory)"
	default:
		return "unknown", "unknown"
	}
}

func listBackupEntries() []BackupEntry {
	backupDir := config.BackupDir()
	var entries []BackupEntry

	files, err := os.ReadDir(backupDir)
	if err != nil {
		return entries
	}

	for i, f := range files {
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

	return entries
}
