package peer

import (
	"testing"
	"time"
)

func TestCreateAndValidateInvite(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	invite, err := CreateInvite("host-a", "aa:bb:cc", 1*time.Hour)
	if err != nil {
		t.Fatalf("CreateInvite() error = %v", err)
	}
	if invite.Token == "" {
		t.Fatal("Token is empty")
	}
	if invite.Used {
		t.Fatal("New invite should not be marked as used")
	}

	validated, err := ValidateInvite(invite.Token)
	if err != nil {
		t.Fatalf("ValidateInvite() error = %v", err)
	}
	if validated.CreatedBy != "host-a" {
		t.Fatalf("CreatedBy = %q, want %q", validated.CreatedBy, "host-a")
	}
}

func TestValidateInviteExpired(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	// Create invite with 0 duration (already expired)
	invite, err := CreateInvite("host-a", "aa:bb:cc", -1*time.Hour)
	if err != nil {
		t.Fatalf("CreateInvite() error = %v", err)
	}

	_, err = ValidateInvite(invite.Token)
	if err == nil {
		t.Fatal("Expected error for expired invite")
	}
}

func TestMarkInviteUsed(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	invite, err := CreateInvite("host-a", "aa:bb:cc", 1*time.Hour)
	if err != nil {
		t.Fatalf("CreateInvite() error = %v", err)
	}

	if err := MarkInviteUsed(invite.Token); err != nil {
		t.Fatalf("MarkInviteUsed() error = %v", err)
	}

	_, err = ValidateInvite(invite.Token)
	if err == nil {
		t.Fatal("Expected error for used invite")
	}
}

func TestValidateInviteInvalidToken(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, err := ValidateInvite("nonexistent-token")
	if err == nil {
		t.Fatal("Expected error for invalid token")
	}
}
