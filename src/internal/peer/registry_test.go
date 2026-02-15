package peer

import (
	"testing"
)

func TestRegistryAddAndGetPeer(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	reg := &Registry{}
	p := Peer{
		ID:                   "host-a",
		Hostname:             "hosta.local",
		TransportFingerprint: "aa:bb:cc",
		AGEPubkey:            "age1hosta",
	}

	if err := reg.AddPeer(p); err != nil {
		t.Fatalf("AddPeer() error = %v", err)
	}

	got, err := reg.GetPeer("host-a")
	if err != nil {
		t.Fatalf("GetPeer() error = %v", err)
	}
	if got.Hostname != "hosta.local" {
		t.Fatalf("Hostname = %q, want %q", got.Hostname, "hosta.local")
	}
	if got.State != StatePending {
		t.Fatalf("State = %q, want %q", got.State, StatePending)
	}
}

func TestRegistryAddDuplicate(t *testing.T) {
	reg := &Registry{}
	p := Peer{ID: "host-a", Hostname: "hosta.local"}
	_ = reg.AddPeer(p)

	err := reg.AddPeer(p)
	if err == nil {
		t.Fatal("Expected error for duplicate peer")
	}
}

func TestRegistryApproveAndRevoke(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	reg := &Registry{}
	_ = reg.AddPeer(Peer{ID: "host-a", Hostname: "hosta.local", AGEPubkey: "age1hosta"})

	if err := reg.ApprovePeer("host-a"); err != nil {
		t.Fatalf("ApprovePeer() error = %v", err)
	}
	got, _ := reg.GetPeer("host-a")
	if got.State != StateApproved {
		t.Fatalf("State = %q, want %q", got.State, StateApproved)
	}
	if !got.HasCapability(CapRead) {
		t.Fatal("Expected CapRead capability after approval")
	}

	if err := reg.RevokePeer("host-a"); err != nil {
		t.Fatalf("RevokePeer() error = %v", err)
	}
	got, _ = reg.GetPeer("host-a")
	if got.State != StateRevoked {
		t.Fatalf("State = %q, want %q", got.State, StateRevoked)
	}
	if len(got.Capabilities) != 0 {
		t.Fatal("Expected empty capabilities after revocation")
	}
}

func TestRegistryListPeers(t *testing.T) {
	reg := &Registry{}
	_ = reg.AddPeer(Peer{ID: "host-a", Hostname: "hosta.local"})
	_ = reg.AddPeer(Peer{ID: "host-b", Hostname: "hostb.local"})
	_ = reg.AddPeer(Peer{ID: "host-c", Hostname: "hostc.local"})

	_ = reg.ApprovePeer("host-a")
	_ = reg.ApprovePeer("host-b")

	approved := reg.ListApprovedPeers()
	if len(approved) != 2 {
		t.Fatalf("ListApprovedPeers() count = %d, want 2", len(approved))
	}

	pending := reg.ListPendingPeers()
	if len(pending) != 1 {
		t.Fatalf("ListPendingPeers() count = %d, want 1", len(pending))
	}
}

func TestRegistryIsAuthorized(t *testing.T) {
	reg := &Registry{}
	_ = reg.AddPeer(Peer{ID: "host-a", Hostname: "hosta.local"})

	// Pending peer should not be authorized
	if reg.IsAuthorized("host-a", CapRead) {
		t.Fatal("Pending peer should not be authorized")
	}

	_ = reg.ApprovePeer("host-a")
	if !reg.IsAuthorized("host-a", CapRead) {
		t.Fatal("Approved peer should be authorized for read")
	}

	// Non-existent peer
	if reg.IsAuthorized("nonexistent", CapRead) {
		t.Fatal("Non-existent peer should not be authorized")
	}
}

func TestRegistrySaveAndLoad(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	reg := &Registry{}
	_ = reg.AddPeer(Peer{ID: "host-a", Hostname: "hosta.local", AGEPubkey: "age1hosta"})
	_ = reg.ApprovePeer("host-a")

	if err := SaveRegistry(reg); err != nil {
		t.Fatalf("SaveRegistry() error = %v", err)
	}

	loaded, err := LoadRegistry()
	if err != nil {
		t.Fatalf("LoadRegistry() error = %v", err)
	}

	if len(loaded.Peers) != 1 {
		t.Fatalf("Loaded peers count = %d, want 1", len(loaded.Peers))
	}
	if loaded.Peers[0].State != StateApproved {
		t.Fatalf("Loaded peer state = %q, want %q", loaded.Peers[0].State, StateApproved)
	}
}

func TestRegistryRemovePeer(t *testing.T) {
	reg := &Registry{}
	_ = reg.AddPeer(Peer{ID: "host-a", Hostname: "hosta.local"})

	if err := reg.RemovePeer("host-a"); err != nil {
		t.Fatalf("RemovePeer() error = %v", err)
	}

	if len(reg.Peers) != 0 {
		t.Fatalf("Peers count = %d, want 0", len(reg.Peers))
	}
}

func TestRegistryGetApprovedAGEPubkeys(t *testing.T) {
	reg := &Registry{}
	_ = reg.AddPeer(Peer{ID: "host-a", Hostname: "hosta.local", AGEPubkey: "age1hosta"})
	_ = reg.AddPeer(Peer{ID: "host-b", Hostname: "hostb.local", AGEPubkey: "age1hostb"})
	_ = reg.AddPeer(Peer{ID: "host-c", Hostname: "hostc.local", AGEPubkey: "age1hostc"})

	_ = reg.ApprovePeer("host-a")
	_ = reg.ApprovePeer("host-b")
	// host-c stays pending

	pubkeys := reg.GetApprovedAGEPubkeys()
	if len(pubkeys) != 2 {
		t.Fatalf("GetApprovedAGEPubkeys() count = %d, want 2", len(pubkeys))
	}
}

func TestRegistryGetPeerByHostname(t *testing.T) {
	reg := &Registry{}
	_ = reg.AddPeer(Peer{ID: "host-a", Hostname: "hosta.local"})

	got, err := reg.GetPeerByHostname("hosta.local")
	if err != nil {
		t.Fatalf("GetPeerByHostname() error = %v", err)
	}
	if got.ID != "host-a" {
		t.Fatalf("ID = %q, want %q", got.ID, "host-a")
	}
}

func TestRegistryGetPeerByFingerprint(t *testing.T) {
	reg := &Registry{}
	_ = reg.AddPeer(Peer{ID: "host-a", Hostname: "hosta.local", TransportFingerprint: "aa:bb:cc"})

	got, err := reg.GetPeerByFingerprint("aa:bb:cc")
	if err != nil {
		t.Fatalf("GetPeerByFingerprint() error = %v", err)
	}
	if got.ID != "host-a" {
		t.Fatalf("ID = %q, want %q", got.ID, "host-a")
	}
}
