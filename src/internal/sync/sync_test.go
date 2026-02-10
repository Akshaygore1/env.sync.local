package syncer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"envsync/internal/config"
)

func TestCachePeerPubkeyCachesKey(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	binDir := filepath.Join(tempDir, "bin")
	if err := os.MkdirAll(binDir, 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	pubkey := "age1testpubkey"
	sshScript := filepath.Join(binDir, "ssh")
	script := "#!/bin/sh\necho \"" + pubkey + "\"\n"
	if err := os.WriteFile(sshScript, []byte(script), 0o700); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	cachePeerPubkey("peer.local")

	cachedPath := filepath.Join(config.AgeKnownHostsDir(), "peer.local.pub")
	data, err := os.ReadFile(cachedPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if strings.TrimSpace(string(data)) != pubkey {
		t.Fatalf("Cached pubkey = %q, want %q", strings.TrimSpace(string(data)), pubkey)
	}
}
