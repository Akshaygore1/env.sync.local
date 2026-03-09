package main

import (
	syncer "envsync/internal/sync"
)

// SyncService provides sync operations
type SyncService struct{}

// SyncResult represents the outcome of a sync operation
type SyncResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Source  string `json:"source"`
}

// SyncAll syncs with all discovered peers
func (s *SyncService) SyncAll() (SyncResult, error) {
	err := syncer.Run(syncer.Options{
		AllPeers: true,
		Quiet:    true,
	})
	if err != nil {
		return SyncResult{Success: false, Message: err.Error()}, nil
	}
	return SyncResult{Success: true, Message: "Synced with all peers"}, nil
}

// SyncFrom syncs from a specific peer
func (s *SyncService) SyncFrom(hostname string) (SyncResult, error) {
	err := syncer.Run(syncer.Options{
		TargetHost: hostname,
		Quiet:      true,
	})
	if err != nil {
		return SyncResult{Success: false, Message: err.Error(), Source: hostname}, nil
	}
	return SyncResult{Success: true, Message: "Synced from " + hostname, Source: hostname}, nil
}

// ForcePull force-pulls from a specific peer
func (s *SyncService) ForcePull(hostname string) (SyncResult, error) {
	err := syncer.Run(syncer.Options{
		TargetHost: hostname,
		ForcePull:  true,
		Quiet:      true,
	})
	if err != nil {
		return SyncResult{Success: false, Message: err.Error(), Source: hostname}, nil
	}
	return SyncResult{Success: true, Message: "Force pulled from " + hostname, Source: hostname}, nil
}

// ForceSync force-syncs with a specific peer
func (s *SyncService) ForceSync(hostname string) (SyncResult, error) {
	err := syncer.Run(syncer.Options{
		TargetHost: hostname,
		Force:      true,
		Quiet:      true,
	})
	if err != nil {
		return SyncResult{Success: false, Message: err.Error(), Source: hostname}, nil
	}
	return SyncResult{Success: true, Message: "Force synced with " + hostname, Source: hostname}, nil
}
