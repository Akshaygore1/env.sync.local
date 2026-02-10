package discovery

import "testing"

func TestParseDnssdPeersAcceptsTrailingDotService(t *testing.T) {
	sample := `
DATE: ---Tue 10 Feb 2026---
21:12:46.949  ...STARTING...
Timestamp     A/R    Flags  if Domain               Service Type         Instance Name
21:12:46.949  Add        3  14 local.               _envsync._tcp.       beelink
21:12:46.949  Add        3   1 local.               _envsync._tcp.       mbp16
21:12:46.949  Add        2  14 local.               _envsync._tcp.       mbp16
`

	peers := parseDnssdPeers(sample)
	expected := []string{"beelink.local", "mbp16.local"}

	if len(peers) != len(expected) {
		t.Fatalf("expected %d peers, got %d (%v)", len(expected), len(peers), peers)
	}

	for i, peer := range expected {
		if peers[i] != peer {
			t.Fatalf("peer %d mismatch: expected %s, got %s", i, peer, peers[i])
		}
	}
}
