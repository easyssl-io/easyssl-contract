package wire

import (
	"testing"
	"time"
)

func TestInventoryFingerprintIsOrderAndClockIndependent(t *testing.T) {
	at := time.Date(2026, 9, 5, 23, 45, 49, 0, time.UTC)
	a := Inventory{
		Diagnostics: Diagnostics{Hostname: "h", OS: "Debian 13", HostType: "vm"},
		Keys: []KeyInfo{
			{Fingerprint: "bb", Algorithm: "ec_p256", CreatedAt: at},
			{Fingerprint: "aa", Algorithm: "ec_p256", CreatedAt: at.In(time.FixedZone("x", 3600)).Add(300 * time.Millisecond)},
		},
	}
	b := Inventory{
		Diagnostics: a.Diagnostics,
		Keys: []KeyInfo{
			{Fingerprint: "aa", Algorithm: "ec_p256", CreatedAt: at},
			{Fingerprint: "bb", Algorithm: "ec_p256", CreatedAt: at},
		},
	}
	fa, err := InventoryFingerprint(a)
	if err != nil {
		t.Fatal(err)
	}
	fb, _ := InventoryFingerprint(b)
	if fa != fb {
		t.Fatalf("key order and sub-second clock changed the fingerprint: %s vs %s", fa, fb)
	}
	if len(fa) != 64 {
		t.Fatalf("fingerprint = %q, want 64 hex characters", fa)
	}

	c := b
	c.Diagnostics.Hostname = "renamed"
	if fc, _ := InventoryFingerprint(c); fc == fb {
		t.Fatal("a changed hostname did not change the fingerprint")
	}
	empty, _ := InventoryFingerprint(Inventory{Diagnostics: a.Diagnostics})
	emptySlice, _ := InventoryFingerprint(Inventory{Diagnostics: a.Diagnostics, Keys: []KeyInfo{}})
	if empty != emptySlice {
		t.Fatal("nil and empty key lists hash differently")
	}
}
