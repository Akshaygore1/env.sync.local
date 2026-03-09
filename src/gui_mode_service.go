//go:build gui

package main

import (
	"envsync/internal/config"
	"envsync/internal/mode"
)

// ModeService provides mode management
type ModeService struct{}

// GetMode returns current mode info
func (m *ModeService) GetMode() (ModeInfo, error) {
	current := mode.GetMode()
	transport, encryption := modeTransportEncryption(current)
	return ModeInfo{
		Current:     string(current),
		Description: mode.ModeDescription(current),
		Features:    modeFeatures(current),
		Transport:   transport,
		Encryption:  encryption,
	}, nil
}

// SetMode changes the operation mode
func (m *ModeService) SetMode(newMode string, pruneOldMaterial bool) error {
	return mode.SetMode(config.SyncMode(newMode), true, pruneOldMaterial)
}

// GetAvailableModes returns all available modes
func (m *ModeService) GetAvailableModes() []ModeInfo {
	var modes []ModeInfo
	for _, sm := range config.ValidSyncModes() {
		transport, encryption := modeTransportEncryption(sm)
		modes = append(modes, ModeInfo{
			Current:     string(sm),
			Description: mode.ModeDescription(sm),
			Features:    modeFeatures(sm),
			Transport:   transport,
			Encryption:  encryption,
		})
	}
	return modes
}

func modeFeatures(m config.SyncMode) []string {
	switch m {
	case config.ModeDevPlaintextHTTP:
		return []string{"Plaintext storage", "HTTP transport", "Debug only"}
	case config.ModeTrustedOwnerSSH:
		return []string{"Plaintext or encrypted storage", "SSH transport", "Zero-touch onboarding"}
	case config.ModeSecurePeer:
		return []string{"Mandatory AGE encryption", "mTLS transport", "Invitation-based onboarding"}
	default:
		return nil
	}
}
