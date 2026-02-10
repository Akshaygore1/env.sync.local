package metadata

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateChecksumMatchesGenerateChecksum(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "secrets.env")
	content := strings.Join([]string{
		"# === ENV_SYNC_METADATA ===",
		"# VERSION: 1.0.0",
		"# TIMESTAMP: 2024-01-01T00:00:00Z",
		"# HOST: example.local",
		"# MODIFIED: 2024-01-01T00:00:00Z",
		"# CHECKSUM: ",
		"# === END_METADATA ===",
		"",
		"FOO=\"bar\"",
		"",
		"# === ENV_SYNC_FOOTER ===",
		"# VERSION: 1.0.0",
		"# TIMESTAMP: 2024-01-01T00:00:00Z",
		"# HOST: example.local",
		"# === END_FOOTER ===",
		"",
	}, "\n")

	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if err := UpdateChecksum(file); err != nil {
		t.Fatalf("UpdateChecksum: %v", err)
	}

	stored := GetFileChecksum(file)
	if stored == "" {
		t.Fatalf("expected checksum to be set")
	}
	if len(stored) != 64 {
		t.Fatalf("expected checksum length 64, got %d", len(stored))
	}

	generated, err := GenerateChecksum(file)
	if err != nil {
		t.Fatalf("GenerateChecksum: %v", err)
	}
	if stored != generated {
		t.Fatalf("checksum mismatch: stored=%s generated=%s", stored, generated)
	}
}

func TestEnsureEncryptedMetadataAndClear(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "secrets.env")
	content := strings.Join([]string{
		"# === ENV_SYNC_METADATA ===",
		"# VERSION: 1.0.0",
		"# TIMESTAMP: 2024-01-01T00:00:00Z",
		"# HOST: oldhost.local",
		"# MODIFIED: 2024-01-01T00:00:00Z",
		"# CHECKSUM: ",
		"# === END_METADATA ===",
		"",
		"FOO=\"bar\"",
		"",
		"# === ENV_SYNC_FOOTER ===",
		"# VERSION: 1.0.0",
		"# TIMESTAMP: 2024-01-01T00:00:00Z",
		"# HOST: oldhost.local",
		"# === END_FOOTER ===",
		"",
	}, "\n")

	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if err := EnsureEncryptedMetadata(file, "newhost.local"); err != nil {
		t.Fatalf("EnsureEncryptedMetadata: %v", err)
	}
	publicKeys := map[string]string{
		"newhost.local": "age1aaa",
		"peer.local":    "age1bbb",
	}
	if err := EnsurePublicKeysMetadata(file, publicKeys); err != nil {
		t.Fatalf("EnsurePublicKeysMetadata: %v", err)
	}
	if err := EnsureEncryptedMetadata(file, "newhost.local"); err != nil {
		t.Fatalf("EnsureEncryptedMetadata (idempotent): %v", err)
	}
	if err := EnsurePublicKeysMetadata(file, publicKeys); err != nil {
		t.Fatalf("EnsurePublicKeysMetadata (idempotent): %v", err)
	}

	updated, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	text := string(updated)
	if !strings.Contains(text, "# HOST: newhost.local") {
		t.Fatalf("expected host to be updated")
	}
	if strings.Count(text, "# ENCRYPTED: true") != 1 {
		t.Fatalf("expected single ENCRYPTED line")
	}
	if strings.Count(text, "# PUBLIC_KEYS: newhost.local:age1aaa,peer.local:age1bbb") != 1 {
		t.Fatalf("expected single PUBLIC_KEYS line")
	}

	if err := ClearEncryptedMetadata(file); err != nil {
		t.Fatalf("ClearEncryptedMetadata: %v", err)
	}

	cleared, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read file after clear: %v", err)
	}
	clearedText := string(cleared)
	if strings.Contains(clearedText, "# ENCRYPTED:") || strings.Contains(clearedText, "# PUBLIC_KEYS:") {
		t.Fatalf("expected encrypted metadata to be removed")
	}
	if !strings.Contains(clearedText, "# HOST: newhost.local") {
		t.Fatalf("expected host to remain updated")
	}
}
