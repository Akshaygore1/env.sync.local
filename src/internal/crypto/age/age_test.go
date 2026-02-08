package age

import "testing"

func TestEncryptDecryptRoundTrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if err := GenerateKey(); err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	pubkey := GetLocalPubkey()
	if pubkey == "" {
		t.Fatal("GetLocalPubkey() returned empty string")
	}

	encrypted, err := EncryptValue("super-secret", []string{pubkey})
	if err != nil {
		t.Fatalf("EncryptValue() error = %v", err)
	}

	decrypted, err := DecryptValue(encrypted)
	if err != nil {
		t.Fatalf("DecryptValue() error = %v", err)
	}
	if decrypted != "super-secret" {
		t.Fatalf("DecryptValue() = %q, want %q", decrypted, "super-secret")
	}
}
