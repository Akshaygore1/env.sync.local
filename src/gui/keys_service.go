package main

import (
	"envsync/internal/config"
	"envsync/internal/keys"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// KeysService provides key management operations
type KeysService struct{}

// KeyInfo represents a key entry
type KeyInfo struct {
	Hostname    string `json:"hostname"`
	PublicKey   string `json:"publicKey"`
	Fingerprint string `json:"fingerprint"`
	IsLocal     bool   `json:"isLocal"`
}

// AccessRequest represents a pending key access request
type AccessRequest struct {
	Hostname  string `json:"hostname"`
	PublicKey string `json:"publicKey"`
	Timestamp string `json:"timestamp"`
}

// GetLocalKey returns the local machine's key info
func (k *KeysService) GetLocalKey() (KeyInfo, error) {
	pubkey := keys.GetLocalPubkey()
	if pubkey == "" {
		return KeyInfo{}, fmt.Errorf("no local key found; run init first")
	}

	hostname, _ := os.Hostname()
	return KeyInfo{
		Hostname:  hostname,
		PublicKey: pubkey,
		IsLocal:   true,
	}, nil
}

// GetPrivateKey returns the private key content (sensitive!)
func (k *KeysService) GetPrivateKey() (string, error) {
	keyPath := keys.GetLocalKeyPath()
	data, err := os.ReadFile(keyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read private key: %w", err)
	}
	return string(data), nil
}

// ExportPublicKey returns the public key as a string
func (k *KeysService) ExportPublicKey() (string, error) {
	pubkey := keys.GetLocalPubkey()
	if pubkey == "" {
		return "", fmt.Errorf("no local key found")
	}
	return pubkey, nil
}

// ImportKey imports a peer's public key
func (k *KeysService) ImportKey(pubkey, hostname string) error {
	if !keys.ValidatePubkey(pubkey) {
		return fmt.Errorf("invalid public key format")
	}
	return keys.CachePeerPubkey(hostname, pubkey)
}

// ImportFromPeer fetches and imports a key from a remote peer via SSH
func (k *KeysService) ImportFromPeer(hostname string) error {
	pubkey := keys.GetCachedPubkey(hostname)
	if pubkey != "" {
		return fmt.Errorf("key already cached for %s", hostname)
	}
	return fmt.Errorf("import from peer requires SSH access; use CLI: env-sync key import --from %s", hostname)
}

// ListKeys returns all cached peer keys
func (k *KeysService) ListKeys() ([]KeyInfo, error) {
	allKeys := keys.GetAllKnownPublicKeys()
	hostname, _ := os.Hostname()

	var result []KeyInfo
	for h, pub := range allKeys {
		result = append(result, KeyInfo{
			Hostname:  h,
			PublicKey: pub,
			IsLocal:   h == hostname,
		})
	}
	return result, nil
}

// RemoveKey removes a cached peer key
func (k *KeysService) RemoveKey(hostname string) error {
	return keys.RemovePeerPubkey(hostname)
}

// RevokeKey removes a peer's key
func (k *KeysService) RevokeKey(hostname string) error {
	return keys.RemovePeerPubkey(hostname)
}

// GetPendingRequests returns pending access requests
func (k *KeysService) GetPendingRequests() ([]AccessRequest, error) {
	requestsDir := config.RequestsDir()
	var requests []AccessRequest

	files, err := os.ReadDir(requestsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return requests, nil
		}
		return nil, err
	}

	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".pub") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(requestsDir, f.Name()))
		if err != nil {
			continue
		}

		hostname := strings.TrimSuffix(f.Name(), ".pub")
		info, _ := f.Info()
		timestamp := ""
		if info != nil {
			timestamp = info.ModTime().Format("2006-01-02T15:04:05Z")
		}

		requests = append(requests, AccessRequest{
			Hostname:  hostname,
			PublicKey: strings.TrimSpace(string(data)),
			Timestamp: timestamp,
		})
	}

	return requests, nil
}

// ApproveRequest approves a pending access request
func (k *KeysService) ApproveRequest(hostname string) error {
	requestsDir := config.RequestsDir()
	reqFile := filepath.Join(requestsDir, hostname+".pub")

	data, err := os.ReadFile(reqFile)
	if err != nil {
		return fmt.Errorf("no pending request from %s", hostname)
	}

	pubkey := strings.TrimSpace(string(data))
	if err := keys.CachePeerPubkey(hostname, pubkey); err != nil {
		return fmt.Errorf("failed to cache key: %w", err)
	}

	return os.Remove(reqFile)
}

// DenyRequest denies and removes a pending request
func (k *KeysService) DenyRequest(hostname string) error {
	requestsDir := config.RequestsDir()
	return os.Remove(filepath.Join(requestsDir, hostname+".pub"))
}
