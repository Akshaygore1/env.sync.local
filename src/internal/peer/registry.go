package peer

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"envsync/internal/config"
	"envsync/internal/logging"
)

// PeerState represents the authorization state of a peer.
type PeerState string

const (
	StatePending  PeerState = "pending"
	StateApproved PeerState = "approved"
	StateRevoked  PeerState = "revoked"
)

// Capability represents a peer's capability.
type Capability string

const (
	CapRead             Capability = "read"
	CapRequestReencrypt Capability = "request_reencrypt"
)

// Peer represents a known peer in the registry.
type Peer struct {
	ID                   string       `json:"id"`
	Hostname             string       `json:"hostname"`
	TransportFingerprint string       `json:"transport_fingerprint"`
	AGEPubkey            string       `json:"age_pubkey"`
	State                PeerState    `json:"state"`
	Capabilities         []Capability `json:"capabilities"`
	AddedAt              string       `json:"added_at"`
	UpdatedAt            string       `json:"updated_at"`
}

// Registry holds the collection of known peers.
type Registry struct {
	Peers []Peer `json:"peers"`
}

// LoadRegistry reads the peer registry from disk.
func LoadRegistry() (*Registry, error) {
	data, err := os.ReadFile(config.PeerRegistryFile())
	if err != nil {
		if os.IsNotExist(err) {
			return &Registry{}, nil
		}
		return nil, err
	}
	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("corrupt peer registry: %w", err)
	}
	return &reg, nil
}

// SaveRegistry writes the peer registry to disk.
func SaveRegistry(reg *Registry) error {
	if err := os.MkdirAll(filepath.Dir(config.PeerRegistryFile()), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(config.PeerRegistryFile(), data, 0o600)
}

// AddPeer adds a new peer to the registry in pending state.
func (r *Registry) AddPeer(p Peer) error {
	for _, existing := range r.Peers {
		if existing.ID == p.ID {
			return fmt.Errorf("peer %s already exists", p.ID)
		}
	}
	now := time.Now().UTC().Format(time.RFC3339)
	p.AddedAt = now
	p.UpdatedAt = now
	if p.State == "" {
		p.State = StatePending
	}
	if p.Capabilities == nil {
		p.Capabilities = []Capability{}
	}
	r.Peers = append(r.Peers, p)
	return nil
}

// ApprovePeer moves a peer from pending to approved state and grants default capabilities.
func (r *Registry) ApprovePeer(peerID string) error {
	for i, p := range r.Peers {
		if p.ID == peerID {
			if p.State == StateApproved {
				return nil // already approved
			}
			if p.State == StateRevoked {
				return fmt.Errorf("peer %s is revoked; remove and re-add to approve", peerID)
			}
			r.Peers[i].State = StateApproved
			r.Peers[i].UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			r.Peers[i].Capabilities = []Capability{CapRead, CapRequestReencrypt}
			logging.Log("SUCCESS", "Approved peer: "+peerID)
			return nil
		}
	}
	return fmt.Errorf("peer %s not found", peerID)
}

// RevokePeer moves a peer to revoked state and clears capabilities.
func (r *Registry) RevokePeer(peerID string) error {
	for i, p := range r.Peers {
		if p.ID == peerID {
			r.Peers[i].State = StateRevoked
			r.Peers[i].UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			r.Peers[i].Capabilities = []Capability{}
			logging.Log("SUCCESS", "Revoked peer: "+peerID)
			return nil
		}
	}
	return fmt.Errorf("peer %s not found", peerID)
}

// GetPeer returns a peer by ID.
func (r *Registry) GetPeer(peerID string) (*Peer, error) {
	for _, p := range r.Peers {
		if p.ID == peerID {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("peer %s not found", peerID)
}

// GetPeerByHostname returns a peer by hostname.
func (r *Registry) GetPeerByHostname(hostname string) (*Peer, error) {
	for _, p := range r.Peers {
		if p.Hostname == hostname {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("peer with hostname %s not found", hostname)
}

// GetPeerByFingerprint returns a peer by transport fingerprint.
func (r *Registry) GetPeerByFingerprint(fingerprint string) (*Peer, error) {
	for _, p := range r.Peers {
		if p.TransportFingerprint == fingerprint {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("peer with fingerprint not found")
}

// ListApprovedPeers returns all approved peers.
func (r *Registry) ListApprovedPeers() []Peer {
	approved := make([]Peer, 0)
	for _, p := range r.Peers {
		if p.State == StateApproved {
			approved = append(approved, p)
		}
	}
	sort.Slice(approved, func(i, j int) bool {
		return approved[i].Hostname < approved[j].Hostname
	})
	return approved
}

// ListPendingPeers returns all peers awaiting approval.
func (r *Registry) ListPendingPeers() []Peer {
	pending := make([]Peer, 0)
	for _, p := range r.Peers {
		if p.State == StatePending {
			pending = append(pending, p)
		}
	}
	return pending
}

// RemovePeer removes a peer from the registry.
func (r *Registry) RemovePeer(peerID string) error {
	for i, p := range r.Peers {
		if p.ID == peerID {
			r.Peers = append(r.Peers[:i], r.Peers[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("peer %s not found", peerID)
}

// HasCapability checks if a peer has a specific capability.
func (p *Peer) HasCapability(cap Capability) bool {
	for _, c := range p.Capabilities {
		if c == cap {
			return true
		}
	}
	return false
}

// IsAuthorized checks if a peer is approved and has the required capability.
func (r *Registry) IsAuthorized(peerID string, cap Capability) bool {
	p, err := r.GetPeer(peerID)
	if err != nil {
		return false
	}
	return p.State == StateApproved && p.HasCapability(cap)
}

// GetApprovedAGEPubkeys returns all AGE public keys from approved peers.
func (r *Registry) GetApprovedAGEPubkeys() []string {
	pubkeys := make([]string, 0)
	for _, p := range r.Peers {
		if p.State == StateApproved && p.AGEPubkey != "" {
			pubkeys = append(pubkeys, p.AGEPubkey)
		}
	}
	sort.Strings(pubkeys)
	return pubkeys
}

// UpdatePeer updates an existing peer's information.
func (r *Registry) UpdatePeer(peerID string, updates func(*Peer)) error {
	for i, p := range r.Peers {
		if p.ID == peerID {
			updates(&r.Peers[i])
			r.Peers[i].UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			return nil
		}
	}
	return errors.New("peer not found: " + peerID)
}
