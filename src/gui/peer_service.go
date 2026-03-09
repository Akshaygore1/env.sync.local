package main

import (
	"envsync/internal/config"
	"envsync/internal/identity"
	"envsync/internal/keys"
	"envsync/internal/mode"
	"envsync/internal/peer"
	"fmt"
	"time"
)

// PeerService provides peer registry management
type PeerService struct{}

// PeerInfo represents a registered peer
type PeerInfo struct {
	ID             string `json:"id"`
	Hostname       string `json:"hostname"`
	State          string `json:"state"`
	TLSFingerprint string `json:"tlsFingerprint"`
	AGEPubKey      string `json:"agePubKey"`
	AddedAt        string `json:"addedAt"`
	UpdatedAt      string `json:"updatedAt"`
}

// InviteInfo represents an enrollment invitation
type InviteInfo struct {
	Token       string `json:"token"`
	CreatedBy   string `json:"createdBy"`
	Fingerprint string `json:"fingerprint"`
	ExpiresAt   string `json:"expiresAt"`
	Command     string `json:"command"`
}

// TrustInfo represents local trust identity
type TrustInfo struct {
	LocalHostname  string     `json:"localHostname"`
	TLSFingerprint string     `json:"tlsFingerprint"`
	CertValidUntil string     `json:"certValidUntil"`
	AGEPublicKey   string     `json:"agePublicKey"`
	TrustedPeers   []PeerInfo `json:"trustedPeers"`
}

// ListPeers returns all registered peers
func (p *PeerService) ListPeers() ([]PeerInfo, error) {
	reg, err := peer.LoadRegistry()
	if err != nil {
		return nil, fmt.Errorf("failed to load peer registry: %w", err)
	}

	var result []PeerInfo
	for _, pr := range reg.Peers {
		result = append(result, peerToInfo(pr))
	}
	return result, nil
}

// ListPending returns pending peers
func (p *PeerService) ListPending() ([]PeerInfo, error) {
	reg, err := peer.LoadRegistry()
	if err != nil {
		return nil, fmt.Errorf("failed to load peer registry: %w", err)
	}

	var result []PeerInfo
	for _, pr := range reg.ListPendingPeers() {
		result = append(result, peerToInfo(pr))
	}
	return result, nil
}

// CreateInvite creates an enrollment invitation
func (p *PeerService) CreateInvite(expiryHours int) (InviteInfo, error) {
	if expiryHours <= 0 {
		expiryHours = 24
	}

	id, err := identity.EnsureIdentity("")
	if err != nil {
		return InviteInfo{}, fmt.Errorf("failed to load identity: %w", err)
	}

	fingerprint := identity.Fingerprint(id.Certificate)
	hostname := id.Certificate.Subject.CommonName

	invite, err := peer.CreateInvite(hostname, fingerprint, time.Duration(expiryHours)*time.Hour)
	if err != nil {
		return InviteInfo{}, fmt.Errorf("failed to create invite: %w", err)
	}

	return InviteInfo{
		Token:       invite.Token,
		CreatedBy:   hostname,
		Fingerprint: fingerprint,
		ExpiresAt:   invite.ExpiresAt,
		Command:     fmt.Sprintf("env-sync peer request %s %s", hostname, invite.Token),
	}, nil
}

// ApprovePeer approves a pending peer
func (p *PeerService) ApprovePeer(peerID string) error {
	reg, err := peer.LoadRegistry()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if err := reg.ApprovePeer(peerID); err != nil {
		return fmt.Errorf("failed to approve peer: %w", err)
	}

	return peer.SaveRegistry(reg)
}

// RevokePeer revokes a peer's access
func (p *PeerService) RevokePeer(peerID string) error {
	reg, err := peer.LoadRegistry()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if err := reg.RevokePeer(peerID); err != nil {
		return fmt.Errorf("failed to revoke peer: %w", err)
	}

	return peer.SaveRegistry(reg)
}

// GetTrustInfo returns local trust identity details
func (p *PeerService) GetTrustInfo() (TrustInfo, error) {
	info := TrustInfo{}

	id, err := identity.LoadIdentity()
	if err != nil {
		return info, fmt.Errorf("no transport identity: %w", err)
	}

	info.LocalHostname = id.Certificate.Subject.CommonName
	info.TLSFingerprint = identity.Fingerprint(id.Certificate)
	info.CertValidUntil = id.Certificate.NotAfter.Format(time.RFC3339)
	info.AGEPublicKey = keys.GetLocalPubkey()

	reg, err := peer.LoadRegistry()
	if err == nil {
		for _, pr := range reg.ListApprovedPeers() {
			info.TrustedPeers = append(info.TrustedPeers, peerToInfo(pr))
		}
	}

	return info, nil
}

// IsSecurePeerMode returns whether the current mode is secure-peer
func (p *PeerService) IsSecurePeerMode() bool {
	return mode.GetMode() == config.ModeSecurePeer
}

func peerToInfo(pr peer.Peer) PeerInfo {
	return PeerInfo{
		ID:             pr.ID,
		Hostname:       pr.Hostname,
		State:          string(pr.State),
		TLSFingerprint: pr.TransportFingerprint,
		AGEPubKey:      pr.AGEPubkey,
		AddedAt:        pr.AddedAt,
		UpdatedAt:      pr.UpdatedAt,
	}
}
