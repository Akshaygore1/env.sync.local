package peer

import (
	"testing"
	"time"
)

func TestCreateAndVerifyEvent(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	key, err := GenerateSigningKey()
	if err != nil {
		t.Fatalf("GenerateSigningKey() error = %v", err)
	}

	log := &MembershipLog{}
	p := Peer{
		ID:                   "host-b",
		Hostname:             "hostb.local",
		TransportFingerprint: "bb:cc:dd",
		AGEPubkey:            "age1hostb",
	}

	event, err := CreateApproveEvent(log, p, "host-a", key, 24*time.Hour)
	if err != nil {
		t.Fatalf("CreateApproveEvent() error = %v", err)
	}

	if event.EventID != 1 {
		t.Fatalf("EventID = %d, want 1", event.EventID)
	}
	if event.Action != "approve" {
		t.Fatalf("Action = %q, want %q", event.Action, "approve")
	}

	// Verify signature
	if !VerifyEvent(event, &key.PublicKey) {
		t.Fatal("Event signature verification failed")
	}
}

func TestVerifyEventWrongKey(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	key1, _ := GenerateSigningKey()
	key2, _ := GenerateSigningKey()

	log := &MembershipLog{}
	p := Peer{ID: "host-b", Hostname: "hostb.local"}

	event, err := CreateApproveEvent(log, p, "host-a", key1, 24*time.Hour)
	if err != nil {
		t.Fatalf("CreateApproveEvent() error = %v", err)
	}

	// Verify with wrong key should fail
	if VerifyEvent(event, &key2.PublicKey) {
		t.Fatal("Expected verification to fail with wrong key")
	}
}

func TestAppendEventStaleRejection(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	key, _ := GenerateSigningKey()
	log := &MembershipLog{LastEvent: 5}

	p := Peer{ID: "host-b", Hostname: "hostb.local"}
	event, _ := CreateApproveEvent(log, p, "host-a", key, 24*time.Hour)

	// Manually set event ID to something stale
	event.EventID = 3

	err := AppendEvent(log, event)
	if err == nil {
		t.Fatal("Expected error for stale event")
	}
}

func TestApplyEventsApprovesPeer(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	key, _ := GenerateSigningKey()

	// Create events log with an approval
	log := &MembershipLog{}
	p := Peer{
		ID:                   "host-b",
		Hostname:             "hostb.local",
		TransportFingerprint: "bb:cc:dd",
		AGEPubkey:            "age1hostb",
	}
	event, _ := CreateApproveEvent(log, p, "host-a", key, 24*time.Hour)
	_ = AppendEvent(log, event)

	// Apply to empty registry
	reg := &Registry{}
	lastApplied, err := ApplyEvents(reg, log, 0)
	if err != nil {
		t.Fatalf("ApplyEvents() error = %v", err)
	}
	if lastApplied != 1 {
		t.Fatalf("lastApplied = %d, want 1", lastApplied)
	}

	got, err := reg.GetPeer("host-b")
	if err != nil {
		t.Fatalf("GetPeer() error = %v", err)
	}
	if got.State != StateApproved {
		t.Fatalf("State = %q, want %q", got.State, StateApproved)
	}
}

func TestApplyEventsRevokesPeer(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	key, _ := GenerateSigningKey()

	log := &MembershipLog{}
	p := Peer{ID: "host-b", Hostname: "hostb.local", AGEPubkey: "age1hostb"}

	// Approve first
	approveEvent, _ := CreateApproveEvent(log, p, "host-a", key, 24*time.Hour)
	_ = AppendEvent(log, approveEvent)

	// Then revoke
	revokeEvent, _ := CreateRevokeEvent(log, p, "host-a", key, 24*time.Hour)
	_ = AppendEvent(log, revokeEvent)

	reg := &Registry{}
	lastApplied, _ := ApplyEvents(reg, log, 0)
	if lastApplied != 2 {
		t.Fatalf("lastApplied = %d, want 2", lastApplied)
	}

	got, _ := reg.GetPeer("host-b")
	if got.State != StateRevoked {
		t.Fatalf("State = %q, want %q", got.State, StateRevoked)
	}
}

func TestGetEventsSince(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	key, _ := GenerateSigningKey()

	log := &MembershipLog{}
	for i := 0; i < 5; i++ {
		p := Peer{ID: "host-" + string(rune('a'+i)), Hostname: "host.local"}
		event, _ := CreateApproveEvent(log, p, "sponsor", key, 24*time.Hour)
		_ = AppendEvent(log, event)
	}

	events := GetEventsSince(log, 3)
	if len(events) != 2 {
		t.Fatalf("GetEventsSince(3) count = %d, want 2", len(events))
	}
	if events[0].EventID != 4 {
		t.Fatalf("First event ID = %d, want 4", events[0].EventID)
	}
}

func TestMembershipLogSaveAndLoad(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	key, _ := GenerateSigningKey()
	log := &MembershipLog{}

	p := Peer{ID: "host-b", Hostname: "hostb.local", AGEPubkey: "age1hostb"}
	event, _ := CreateApproveEvent(log, p, "host-a", key, 24*time.Hour)
	_ = AppendEvent(log, event)

	if err := SaveMembershipLog(log); err != nil {
		t.Fatalf("SaveMembershipLog() error = %v", err)
	}

	loaded, err := LoadMembershipLog()
	if err != nil {
		t.Fatalf("LoadMembershipLog() error = %v", err)
	}
	if loaded.LastEvent != 1 {
		t.Fatalf("LastEvent = %d, want 1", loaded.LastEvent)
	}
	if len(loaded.Events) != 1 {
		t.Fatalf("Events count = %d, want 1", len(loaded.Events))
	}
}
