package identity

import (
	"os"
	"testing"

	"envsync/internal/config"
)

func TestGenerateTransportIdentity(t *testing.T) {
	id, err := GenerateTransportIdentity("test-host.local")
	if err != nil {
		t.Fatalf("GenerateTransportIdentity() error = %v", err)
	}
	if id.Key == nil {
		t.Fatal("Key is nil")
	}
	if id.Certificate == nil {
		t.Fatal("Certificate is nil")
	}
	if len(id.CertPEM) == 0 {
		t.Fatal("CertPEM is empty")
	}
	if len(id.KeyPEM) == 0 {
		t.Fatal("KeyPEM is empty")
	}
	if id.Certificate.Subject.CommonName != "test-host.local" {
		t.Fatalf("CommonName = %q, want %q", id.Certificate.Subject.CommonName, "test-host.local")
	}
}

func TestGenerateTransportIdentityEmptyHostname(t *testing.T) {
	_, err := GenerateTransportIdentity("")
	if err == nil {
		t.Fatal("Expected error for empty hostname")
	}
}

func TestFingerprint(t *testing.T) {
	id, err := GenerateTransportIdentity("test-host.local")
	if err != nil {
		t.Fatalf("GenerateTransportIdentity() error = %v", err)
	}
	fp := Fingerprint(id.Certificate)
	if fp == "" {
		t.Fatal("Fingerprint is empty")
	}
	// SHA-256 fingerprint should be 32 bytes = 64 hex chars + 31 colons = 95 chars
	if len(fp) != 95 {
		t.Fatalf("Fingerprint length = %d, want 95", len(fp))
	}
}

func TestSaveAndLoadIdentity(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	id, err := GenerateTransportIdentity("test-host.local")
	if err != nil {
		t.Fatalf("GenerateTransportIdentity() error = %v", err)
	}

	if err := SaveIdentity(id); err != nil {
		t.Fatalf("SaveIdentity() error = %v", err)
	}

	loaded, err := LoadIdentity()
	if err != nil {
		t.Fatalf("LoadIdentity() error = %v", err)
	}

	// Compare fingerprints - they should match
	origFP := Fingerprint(id.Certificate)
	loadedFP := Fingerprint(loaded.Certificate)
	if origFP != loadedFP {
		t.Fatalf("Fingerprints don't match: %q vs %q", origFP, loadedFP)
	}
}

func TestIdentityExists(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if IdentityExists() {
		t.Fatal("IdentityExists() should be false before generation")
	}

	id, err := GenerateTransportIdentity("test-host.local")
	if err != nil {
		t.Fatalf("GenerateTransportIdentity() error = %v", err)
	}
	if err := SaveIdentity(id); err != nil {
		t.Fatalf("SaveIdentity() error = %v", err)
	}

	if !IdentityExists() {
		t.Fatal("IdentityExists() should be true after save")
	}
}

func TestEnsureIdentityGeneratesOnFirstCall(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	id, err := EnsureIdentity("test-host.local")
	if err != nil {
		t.Fatalf("EnsureIdentity() error = %v", err)
	}
	fp1 := Fingerprint(id.Certificate)

	// Second call should load the same identity
	id2, err := EnsureIdentity("test-host.local")
	if err != nil {
		t.Fatalf("EnsureIdentity() second call error = %v", err)
	}
	fp2 := Fingerprint(id2.Certificate)

	if fp1 != fp2 {
		t.Fatalf("Expected same fingerprint on second call: %q vs %q", fp1, fp2)
	}
}

func TestSaveAndLoadTrustedCert(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	id, err := GenerateTransportIdentity("peer-host.local")
	if err != nil {
		t.Fatalf("GenerateTransportIdentity() error = %v", err)
	}

	if err := SaveTrustedCert("peer-host", id.CertPEM); err != nil {
		t.Fatalf("SaveTrustedCert() error = %v", err)
	}

	loaded, err := LoadTrustedCert("peer-host")
	if err != nil {
		t.Fatalf("LoadTrustedCert() error = %v", err)
	}

	if Fingerprint(loaded) != Fingerprint(id.Certificate) {
		t.Fatal("Loaded trusted cert fingerprint doesn't match")
	}

	// Verify file permissions
	info, err := os.Stat(config.TLSDir())
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if info.Mode().Perm() != 0o700 {
		t.Fatalf("TLSDir permissions = %o, want 700", info.Mode().Perm())
	}
}
