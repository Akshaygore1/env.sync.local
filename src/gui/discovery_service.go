package main

import (
	"envsync/internal/discovery"
	"envsync/internal/keys"
	"time"
)

// DiscoveryService provides peer discovery operations
type DiscoveryService struct{}

// DiscoveredPeer represents a peer found on the network
type DiscoveredPeer struct {
	Hostname  string `json:"hostname"`
	Address   string `json:"address"`
	Port      int    `json:"port"`
	SSHAccess bool   `json:"sshAccess"`
	HasPubKey bool   `json:"hasPubKey"`
	Reachable bool   `json:"reachable"`
}

// Discover finds peers via mDNS
func (d *DiscoveryService) Discover(timeoutSec int) ([]DiscoveredPeer, error) {
	timeout := time.Duration(timeoutSec) * time.Second
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	peers, err := discovery.Discover(discovery.Options{
		Timeout: timeout,
		Quiet:   true,
	})
	if err != nil {
		return nil, err
	}

	var result []DiscoveredPeer
	for _, hostname := range peers {
		result = append(result, DiscoveredPeer{
			Hostname:  hostname,
			Reachable: true,
			HasPubKey: keys.GetCachedPubkey(hostname) != "",
		})
	}
	return result, nil
}

// DiscoverSSH finds peers and filters to SSH-reachable only
func (d *DiscoveryService) DiscoverSSH(timeoutSec int) ([]DiscoveredPeer, error) {
	timeout := time.Duration(timeoutSec) * time.Second
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	peers, err := discovery.Discover(discovery.Options{
		Timeout:   timeout,
		Quiet:     true,
		FilterSSH: true,
	})
	if err != nil {
		return nil, err
	}

	var result []DiscoveredPeer
	for _, hostname := range peers {
		result = append(result, DiscoveredPeer{
			Hostname:  hostname,
			SSHAccess: true,
			Reachable: true,
			HasPubKey: keys.GetCachedPubkey(hostname) != "",
		})
	}
	return result, nil
}

// CollectKeys discovers peers and collects their public keys
func (d *DiscoveryService) CollectKeys(timeoutSec int) (int, error) {
	timeout := time.Duration(timeoutSec) * time.Second
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	peers, err := discovery.Discover(discovery.Options{
		Timeout:     timeout,
		Quiet:       true,
		CollectKeys: true,
	})
	if err != nil {
		return 0, err
	}

	return len(peers), nil
}
