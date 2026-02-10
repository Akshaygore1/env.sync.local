package syncer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"envsync/internal/config"
)

func writeSSHScript(t *testing.T, dir string, output string) {
	t.Helper()
	sshScript := filepath.Join(dir, "ssh")
	script := "#!/bin/sh\n"
	if output != "" {
		script += "echo \"" + output + "\"\n"
	}
	if err := os.WriteFile(sshScript, []byte(script), 0o700); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func TestCachePeerPubkeyCachesKey(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	binDir := filepath.Join(tempDir, "bin")
	if err := os.MkdirAll(binDir, 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	pubkey := "age1testpubkey"
	writeSSHScript(t, binDir, pubkey)

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	cachePeerPubkey("peer.local")

	if _, err := os.Stat(config.AgeKnownHostsDir()); err != nil {
		t.Fatalf("AgeKnownHostsDir() error = %v", err)
	}

	cachedPath := filepath.Join(config.AgeKnownHostsDir(), "peer.local.pub")
	data, err := os.ReadFile(cachedPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if strings.TrimSpace(string(data)) != pubkey {
		t.Fatalf("Cached pubkey = %q, want %q", strings.TrimSpace(string(data)), pubkey)
	}
}

func TestCachePeerPubkeySkipsInvalidKey(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	binDir := filepath.Join(tempDir, "bin")
	if err := os.MkdirAll(binDir, 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	writeSSHScript(t, binDir, "not-a-pubkey")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	cachePeerPubkey("peer.local")

	cachedPath := filepath.Join(config.AgeKnownHostsDir(), "peer.local.pub")
	if _, err := os.Stat(cachedPath); err == nil {
		t.Fatal("Expected no cached pubkey for invalid key")
	}
}

func TestCachePeerPubkeySkipsEmptyKey(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	binDir := filepath.Join(tempDir, "bin")
	if err := os.MkdirAll(binDir, 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	writeSSHScript(t, binDir, "")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	cachePeerPubkey("peer.local")

	cachedPath := filepath.Join(config.AgeKnownHostsDir(), "peer.local.pub")
	if _, err := os.Stat(cachedPath); err == nil {
		t.Fatal("Expected no cached pubkey for empty key")
	}
}

func TestEnsureRegisteredWithPeerSuccess(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	// Generate a local key so GetLocalPubkey returns something
	if err := os.MkdirAll(config.AgeKeyDir(), 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(config.AgePubKeyFile(), []byte("age1testlocalpubkey"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Create a mock ssh script that echoes ENVSYNC_REGISTER_SUCCESS
	binDir := filepath.Join(tempDir, "bin")
	if err := os.MkdirAll(binDir, 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	sshScript := filepath.Join(binDir, "ssh")
	script := "#!/bin/sh\necho 'ENVSYNC_REGISTER_SUCCESS'\n"
	if err := os.WriteFile(sshScript, []byte(script), 0o700); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	err := ensureRegisteredWithPeer("peer.local")
	if err != nil {
		t.Fatalf("ensureRegisteredWithPeer() error = %v", err)
	}
}

func TestEnsureRegisteredWithPeerNoLocalKey(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	// No key file created, so GetLocalPubkey returns ""
	err := ensureRegisteredWithPeer("peer.local")
	if err == nil {
		t.Fatal("Expected error when no local key exists")
	}
	if !strings.Contains(err.Error(), "no local key found") {
		t.Fatalf("Expected 'no local key found' error, got: %v", err)
	}
}

func TestEnsureRegisteredWithPeerSSHFailure(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	// Generate a local key
	if err := os.MkdirAll(config.AgeKeyDir(), 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(config.AgePubKeyFile(), []byte("age1testlocalpubkey"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Create a mock ssh script that fails
	binDir := filepath.Join(tempDir, "bin")
	if err := os.MkdirAll(binDir, 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	sshScript := filepath.Join(binDir, "ssh")
	script := "#!/bin/sh\nexit 1\n"
	if err := os.WriteFile(sshScript, []byte(script), 0o700); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	err := ensureRegisteredWithPeer("peer.local")
	if err == nil {
		t.Fatal("Expected error when SSH fails")
	}
}
