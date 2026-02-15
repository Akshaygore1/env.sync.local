package mode

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"envsync/internal/config"
	"envsync/internal/logging"
)

// GetMode reads the current sync mode from the mode config file.
// Returns the default mode (trusted-owner-ssh) if no config exists.
func GetMode() config.SyncMode {
	data, err := os.ReadFile(config.ModeConfigFile())
	if err != nil {
		return config.DefaultMode
	}
	m := strings.TrimSpace(string(data))
	if config.IsValidSyncMode(m) {
		return config.SyncMode(m)
	}
	return config.DefaultMode
}

// SetMode writes the sync mode to the config file.
// It validates the mode and performs safety checks for mode transitions.
func SetMode(newMode config.SyncMode, yes bool, pruneOldMaterial bool) error {
	if !config.IsValidSyncMode(string(newMode)) {
		return fmt.Errorf("invalid mode: %s (valid modes: %s)",
			newMode, strings.Join(modeStrings(), ", "))
	}

	currentMode := GetMode()
	if currentMode == newMode {
		logging.Log("INFO", fmt.Sprintf("Already in mode: %s", newMode))
		return nil
	}

	// Safety: warn on downgrade from secure-peer
	if currentMode == config.ModeSecurePeer && newMode != config.ModeSecurePeer {
		logging.Log("WARNING", "╔════════════════════════════════════════════════════════════╗")
		logging.Log("WARNING", "║  SECURITY DOWNGRADE: Switching away from secure-peer mode ║")
		logging.Log("WARNING", "║  Peer-to-peer encryption and authentication will be lost  ║")
		logging.Log("WARNING", "╚════════════════════════════════════════════════════════════╝")
		if !yes {
			return errors.New("security downgrade requires --yes flag to confirm")
		}
	}

	// Safety: warn when using dev-plaintext-http
	if newMode == config.ModeDevPlaintextHTTP {
		logging.Log("WARNING", "╔════════════════════════════════════════════════════════════════╗")
		logging.Log("WARNING", "║  ⚠️  DEV MODE: Secrets will be transmitted in PLAINTEXT       ║")
		logging.Log("WARNING", "║  This mode is for LOCAL DEBUGGING ONLY. Never use in prod.   ║")
		logging.Log("WARNING", "╚════════════════════════════════════════════════════════════════╝")
		if !yes {
			return errors.New("dev-plaintext-http mode requires --yes flag to confirm")
		}
	}

	if pruneOldMaterial {
		if !yes {
			return errors.New("--prune-old-material requires --yes flag")
		}
		if err := pruneModeMaterial(currentMode); err != nil {
			logging.Log("WARN", fmt.Sprintf("Failed to prune old material: %v", err))
		}
	}

	if err := os.MkdirAll(filepath.Dir(config.ModeConfigFile()), 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(config.ModeConfigFile(), []byte(string(newMode)+"\n"), 0o600); err != nil {
		return err
	}

	logging.Log("SUCCESS", fmt.Sprintf("Mode set to: %s", newMode))
	return nil
}

// pruneModeMaterial removes security material from the previous mode.
func pruneModeMaterial(oldMode config.SyncMode) error {
	switch oldMode {
	case config.ModeSecurePeer:
		logging.Log("INFO", "Pruning secure-peer material (TLS certs, peer registry)...")
		_ = os.RemoveAll(config.TLSDir())
		_ = os.RemoveAll(config.PeerRegistryDir())
		logging.Log("SUCCESS", "Pruned secure-peer material")
	case config.ModeTrustedOwnerSSH:
		logging.Log("INFO", "Pruning trusted-owner-ssh material (AGE keys, known hosts)...")
		// Note: not removing AGE keys by default as they may be needed for decryption
		_ = os.RemoveAll(config.AgeKnownHostsDir())
		logging.Log("SUCCESS", "Pruned trusted-owner-ssh material")
	case config.ModeDevPlaintextHTTP:
		// Nothing to prune for dev mode
	}
	return nil
}

// ModeDescription returns a human-readable description of a mode.
func ModeDescription(m config.SyncMode) string {
	switch m {
	case config.ModeDevPlaintextHTTP:
		return "Development/debug only — plaintext HTTP, no auth"
	case config.ModeTrustedOwnerSSH:
		return "Same-owner devices — SCP/SSH transport, mutual trust"
	case config.ModeSecurePeer:
		return "Cross-owner collaboration — HTTPS+mTLS, encrypted storage, peer auth"
	default:
		return "Unknown mode"
	}
}

func modeStrings() []string {
	modes := config.ValidSyncModes()
	result := make([]string, len(modes))
	for i, m := range modes {
		result[i] = string(m)
	}
	return result
}
