package mtlstransport

import (
	"testing"

	"envsync/internal/identity"
)

func TestNewServerTLSConfigRequiresIdentity(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, err := NewServerTLSConfig()
	if err == nil {
		t.Fatal("Expected error when no identity exists")
	}
}

func TestNewServerTLSConfigWithIdentity(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, err := identity.EnsureIdentity("test-server.local")
	if err != nil {
		t.Fatalf("EnsureIdentity() error = %v", err)
	}

	cfg, err := NewServerTLSConfig()
	if err != nil {
		t.Fatalf("NewServerTLSConfig() error = %v", err)
	}

	if len(cfg.Certificates) != 1 {
		t.Fatalf("Certificates count = %d, want 1", len(cfg.Certificates))
	}

	if cfg.MinVersion != 0x0304 { // TLS 1.3
		t.Fatalf("MinVersion = %x, want 0x0304 (TLS 1.3)", cfg.MinVersion)
	}
}

func TestNewClientTLSConfigWithIdentity(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, err := identity.EnsureIdentity("test-client.local")
	if err != nil {
		t.Fatalf("EnsureIdentity() error = %v", err)
	}

	cfg, err := NewClientTLSConfig("some-peer")
	if err != nil {
		t.Fatalf("NewClientTLSConfig() error = %v", err)
	}

	if len(cfg.Certificates) != 1 {
		t.Fatalf("Certificates count = %d, want 1", len(cfg.Certificates))
	}
}

func TestPinnedFingerprintReturnsEmptyWithoutTrustedCert(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	fp := pinnedFingerprint("nonexistent-peer")
	if fp != "" {
		t.Fatalf("Expected empty fingerprint, got %q", fp)
	}
}

func TestPinnedFingerprintMatchesTrustedCert(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	id, err := identity.GenerateTransportIdentity("peer.local")
	if err != nil {
		t.Fatalf("GenerateTransportIdentity() error = %v", err)
	}

	if err := identity.SaveTrustedCert("peer-host", id.CertPEM); err != nil {
		t.Fatalf("SaveTrustedCert() error = %v", err)
	}

	fp := pinnedFingerprint("peer-host")
	if fp == "" {
		t.Fatal("Expected non-empty fingerprint for trusted cert")
	}
	if fp != identity.Fingerprint(id.Certificate) {
		t.Fatal("Fingerprint doesn't match expected value")
	}
}
