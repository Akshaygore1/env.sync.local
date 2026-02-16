package mode

import (
	"os"
	"testing"

	"envsync/internal/config"
)

func TestGetModeDefaultWhenNoFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	m := GetMode()
	if m != config.DefaultMode {
		t.Fatalf("GetMode() = %q, want %q", m, config.DefaultMode)
	}
}

func TestSetAndGetMode(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	// Set to trusted-owner-ssh
	if err := SetMode(config.ModeTrustedOwnerSSH, true, false); err != nil {
		t.Fatalf("SetMode() error = %v", err)
	}
	m := GetMode()
	if m != config.ModeTrustedOwnerSSH {
		t.Fatalf("GetMode() = %q, want %q", m, config.ModeTrustedOwnerSSH)
	}

	// Set to dev-plaintext-http (requires --yes)
	if err := SetMode(config.ModeDevPlaintextHTTP, true, false); err != nil {
		t.Fatalf("SetMode() error = %v", err)
	}
	m = GetMode()
	if m != config.ModeDevPlaintextHTTP {
		t.Fatalf("GetMode() = %q, want %q", m, config.ModeDevPlaintextHTTP)
	}
}

func TestSetModeDevPlaintextRequiresYes(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	err := SetMode(config.ModeDevPlaintextHTTP, false, false)
	if err == nil {
		t.Fatal("Expected error for dev-plaintext-http without --yes")
	}
}

func TestSetModeDowngradeFromSecurePeerRequiresYes(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	// First set to secure-peer
	if err := SetMode(config.ModeSecurePeer, true, false); err != nil {
		t.Fatalf("SetMode(secure-peer) error = %v", err)
	}

	// Try downgrade without --yes
	err := SetMode(config.ModeTrustedOwnerSSH, false, false)
	if err == nil {
		t.Fatal("Expected error for downgrade from secure-peer without --yes")
	}
}

func TestSetModeInvalidMode(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	err := SetMode(config.SyncMode("invalid-mode"), true, false)
	if err == nil {
		t.Fatal("Expected error for invalid mode")
	}
}

func TestSetModeSameModeNoOp(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	// Set initial mode
	if err := SetMode(config.ModeTrustedOwnerSSH, true, false); err != nil {
		t.Fatalf("SetMode() error = %v", err)
	}

	// Set same mode again — should be a no-op
	if err := SetMode(config.ModeTrustedOwnerSSH, false, false); err != nil {
		t.Fatalf("SetMode() same mode should not error, got: %v", err)
	}
}

func TestPruneOldMaterialRequiresYes(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if err := SetMode(config.ModeTrustedOwnerSSH, true, false); err != nil {
		t.Fatalf("SetMode() error = %v", err)
	}

	// Try prune without --yes
	err := SetMode(config.ModeSecurePeer, false, true)
	if err == nil {
		t.Fatal("Expected error for --prune-old-material without --yes")
	}
}

func TestPruneOldMaterialFromSecurePeer(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Set up secure-peer mode with some TLS material
	if err := SetMode(config.ModeSecurePeer, true, false); err != nil {
		t.Fatalf("SetMode() error = %v", err)
	}
	if err := os.MkdirAll(config.TLSDir(), 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(config.TLSKeyFile(), []byte("fake-key"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Switch to trusted-owner-ssh with prune
	if err := SetMode(config.ModeTrustedOwnerSSH, true, true); err != nil {
		t.Fatalf("SetMode(prune) error = %v", err)
	}

	// TLS dir should be gone
	if _, err := os.Stat(config.TLSDir()); err == nil {
		t.Fatal("Expected TLS dir to be removed after prune")
	}
}

func TestModeDescription(t *testing.T) {
	desc := ModeDescription(config.ModeSecurePeer)
	if desc == "" || desc == "Unknown mode" {
		t.Fatalf("ModeDescription(secure-peer) = %q, want non-empty description", desc)
	}
}
