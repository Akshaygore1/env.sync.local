package keys

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"envsync/internal/crypto/age"
)

func TestDecryptSecretsFile_FailsOnInvalidEncryption(t *testing.T) {
	// Setup temp home directory
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	// Generate a key pair for the "local" user
	if err := age.GenerateKey(); err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	// Generate a different key pair for the "remote" user
	remoteDir := filepath.Join(tempDir, "remote")
	if err := os.MkdirAll(filepath.Join(remoteDir, ".config", "env-sync", "keys"), 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	t.Setenv("HOME", remoteDir)
	if err := age.GenerateKey(); err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	remotePubkey := age.GetLocalPubkey()

	// Switch back to local
	t.Setenv("HOME", tempDir)

	// Create a secrets file encrypted with the remote key only
	secretsFile := filepath.Join(tempDir, "test_secrets.env")
	value1, err := age.EncryptValue("secret-value-1", []string{remotePubkey})
	if err != nil {
		t.Fatalf("EncryptValue() error = %v", err)
	}
	value2, err := age.EncryptValue("secret-value-2", []string{remotePubkey})
	if err != nil {
		t.Fatalf("EncryptValue() error = %v", err)
	}

	content := `# === ENV_SYNC_METADATA ===
# VERSION: 2.0.0
# TIMESTAMP: 2025-02-08T15:30:45Z
# HOST: remote.local
# MODIFIED: 2025-02-08T15:30:45Z
# ENCRYPTED: true
# RECIPIENTS: ` + remotePubkey + `
# === END_METADATA ===

OPENAI_API_KEY="` + value1 + `" # ENVSYNC_UPDATED_AT=2025-02-08T15:30:45Z
DATABASE_URL="` + value2 + `" # ENVSYNC_UPDATED_AT=2025-02-08T14:20:10Z

# === ENV_SYNC_FOOTER ===
# VERSION: 2.0.0
# TIMESTAMP: 2025-02-08T15:30:45Z
# HOST: remote.local
# === END_FOOTER ===`

	if err := os.WriteFile(secretsFile, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Attempt to decrypt - should fail because we don't have the remote key
	outputFile := filepath.Join(tempDir, "decrypted.env")
	err = DecryptSecretsFile(secretsFile, outputFile)

	if err == nil {
		t.Fatal("DecryptSecretsFile() expected error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to decrypt") {
		t.Errorf("DecryptSecretsFile() error = %v, want error containing 'failed to decrypt'", err)
	}

	if !strings.Contains(err.Error(), "OPENAI_API_KEY") {
		t.Errorf("DecryptSecretsFile() error = %v, want error mentioning 'OPENAI_API_KEY'", err)
	}
}

func TestDecryptSecretsFile_SucceedsWithCorrectKey(t *testing.T) {
	// Setup temp home directory
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	// Generate a key pair
	if err := age.GenerateKey(); err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	pubkey := age.GetLocalPubkey()

	// Create a secrets file encrypted with our key
	secretsFile := filepath.Join(tempDir, "test_secrets.env")
	value1, err := age.EncryptValue("secret-value-1", []string{pubkey})
	if err != nil {
		t.Fatalf("EncryptValue() error = %v", err)
	}
	value2, err := age.EncryptValue("secret-value-2", []string{pubkey})
	if err != nil {
		t.Fatalf("EncryptValue() error = %v", err)
	}

	content := `# === ENV_SYNC_METADATA ===
# VERSION: 2.0.0
# TIMESTAMP: 2025-02-08T15:30:45Z
# HOST: local.local
# MODIFIED: 2025-02-08T15:30:45Z
# ENCRYPTED: true
# RECIPIENTS: ` + pubkey + `
# === END_METADATA ===

OPENAI_API_KEY="` + value1 + `" # ENVSYNC_UPDATED_AT=2025-02-08T15:30:45Z
DATABASE_URL="` + value2 + `" # ENVSYNC_UPDATED_AT=2025-02-08T14:20:10Z

# === ENV_SYNC_FOOTER ===
# VERSION: 2.0.0
# TIMESTAMP: 2025-02-08T15:30:45Z
# HOST: local.local
# === END_FOOTER ===`

	if err := os.WriteFile(secretsFile, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Decrypt - should succeed
	outputFile := filepath.Join(tempDir, "decrypted.env")
	err = DecryptSecretsFile(secretsFile, outputFile)

	if err != nil {
		t.Fatalf("DecryptSecretsFile() error = %v, want nil", err)
	}

	// Verify decrypted content
	decrypted, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	decryptedStr := string(decrypted)
	if !strings.Contains(decryptedStr, `OPENAI_API_KEY="secret-value-1"`) {
		t.Errorf("Decrypted content missing OPENAI_API_KEY, got: %s", decryptedStr)
	}
	if !strings.Contains(decryptedStr, `DATABASE_URL="secret-value-2"`) {
		t.Errorf("Decrypted content missing DATABASE_URL, got: %s", decryptedStr)
	}
}

func TestRecipientsContain_UsesCommaDelimitedMatches(t *testing.T) {
	recipients := "age1abc123, age1def456"

	if !RecipientsContain(recipients, "age1def456") {
		t.Fatal("RecipientsContain() expected true for exact recipient match")
	}

	if RecipientsContain(recipients, "age1abc") {
		t.Fatal("RecipientsContain() expected false for substring match")
	}
}
