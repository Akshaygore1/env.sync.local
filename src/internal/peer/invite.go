package peer

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"envsync/internal/config"
	"envsync/internal/logging"
)

// Invite represents a one-time enrollment token for a new peer.
type Invite struct {
	Token                string `json:"token"`
	CreatedBy            string `json:"created_by"`
	TransportFingerprint string `json:"transport_fingerprint"`
	ExpiresAt            string `json:"expires_at"`
	Used                 bool   `json:"used"`
}

// CreateInvite generates a new enrollment invite with expiry.
func CreateInvite(hostID string, fingerprint string, expiry time.Duration) (*Invite, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	invite := &Invite{
		Token:                hex.EncodeToString(tokenBytes),
		CreatedBy:            hostID,
		TransportFingerprint: fingerprint,
		ExpiresAt:            time.Now().UTC().Add(expiry).Format(time.RFC3339),
		Used:                 false,
	}

	if err := saveInvite(invite); err != nil {
		return nil, err
	}

	logging.Log("SUCCESS", "Created enrollment invite (expires: "+invite.ExpiresAt+")")
	return invite, nil
}

// ValidateInvite checks if an invite token is valid and not expired.
func ValidateInvite(token string) (*Invite, error) {
	invite, err := loadInvite(token)
	if err != nil {
		return nil, fmt.Errorf("invalid or unknown invite token")
	}

	if invite.Used {
		return nil, fmt.Errorf("invite token already used")
	}

	expiry, err := time.Parse(time.RFC3339, invite.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("invalid expiry format")
	}

	if time.Now().UTC().After(expiry) {
		return nil, fmt.Errorf("invite token expired at %s", invite.ExpiresAt)
	}

	return invite, nil
}

// MarkInviteUsed marks an invite as used.
func MarkInviteUsed(token string) error {
	invite, err := loadInvite(token)
	if err != nil {
		return err
	}
	invite.Used = true
	return saveInvite(invite)
}

func saveInvite(invite *Invite) error {
	if err := os.MkdirAll(config.InviteDir(), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(invite, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(config.InviteDir(), invite.Token+".json"), data, 0o600)
}

func loadInvite(token string) (*Invite, error) {
	data, err := os.ReadFile(filepath.Join(config.InviteDir(), token+".json"))
	if err != nil {
		return nil, err
	}
	var invite Invite
	if err := json.Unmarshal(data, &invite); err != nil {
		return nil, err
	}
	return &invite, nil
}

// CleanExpiredInvites removes expired invite files.
func CleanExpiredInvites() {
	files, _ := filepath.Glob(filepath.Join(config.InviteDir(), "*.json"))
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		var invite Invite
		if err := json.Unmarshal(data, &invite); err != nil {
			continue
		}
		expiry, err := time.Parse(time.RFC3339, invite.ExpiresAt)
		if err != nil {
			continue
		}
		if time.Now().UTC().After(expiry) {
			_ = os.Remove(file)
		}
	}
}
