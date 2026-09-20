package wire

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"time"
)

// RouteAgentInventory is where an Agent PUTs its whole host inventory,
// replacing what the Server holds for it. Sent only when the check-in answers
// that the Server wants it (CheckinResponse.ReportInventory), so the check-in
// stays constant-size and this route carries the payload however large a host's
// state grows.
const RouteAgentInventory = "/api/v1/agent/inventory"

// Inventory is everything an Agent observes about its own host, recomputed on
// every check-in and never persisted on the host: the diagnostics that used to
// be reported at registration alone, and the keystore's key list. Nothing
// Server-owned is in it, and no private key: KeyInfo carries a fingerprint.
//
// Replacing semantics are the point. An entry absent from the document is one
// the host no longer has, which is what lets a deleted key stop being a holder
// without a route to say so; it is also why an Agent that failed to read any
// contributor must not send the document at all (S2 FR-026).
type Inventory struct {
	Diagnostics Diagnostics `json:"diagnostics"`
	Keys        []KeyInfo   `json:"keys"`
}

// KeyInfo is one private key in the Agent's keystore, by its public half only.
type KeyInfo struct {
	// Fingerprint is the SHA-256 of the key's SubjectPublicKeyInfo, lowercase
	// hex, which is also the key file's name on the host and what the Server
	// computes from a CSR or certificate without seeing the key.
	Fingerprint string `json:"fingerprint"`
	// Algorithm as the Server names key types: "ec_p256" today.
	Algorithm string `json:"algorithm"`
	// CreatedAt is when the key file was created, as the host reports it.
	CreatedAt time.Time `json:"createdAt"`
}

// InventoryFingerprint is the one hash both sides compute over an Inventory: the
// Agent to put on its check-in, the Server to store beside the document and
// compare against the next check-in. Both call this function, from this
// module, because if the two canonicalized differently every check-in would
// disagree, every check-in would ask for a report, and the mechanism would
// invert into a permanent full-inventory push that nothing reports as broken.
//
// Canonical form: keys sorted by fingerprint, then the JSON encoding with
// object members in sorted order, which encoding/json gives for a map. Times
// are compared as RFC 3339 in UTC so a document that round-tripped through the
// Server's clock-free storage hashes the same.
func InventoryFingerprint(inv Inventory) (string, error) {
	keys := make([]KeyInfo, len(inv.Keys))
	copy(keys, inv.Keys)
	sort.Slice(keys, func(i, j int) bool { return keys[i].Fingerprint < keys[j].Fingerprint })
	for i := range keys {
		keys[i].CreatedAt = keys[i].CreatedAt.UTC().Truncate(time.Second)
	}
	canon := Inventory{Diagnostics: inv.Diagnostics, Keys: keys}
	if canon.Keys == nil {
		canon.Keys = []KeyInfo{}
	}
	// Through a generic value, so member order is the sorted order a map gives
	// and not the struct's declaration order, which a field added later would
	// otherwise change.
	raw, err := json.Marshal(canon)
	if err != nil {
		return "", err
	}
	var generic any
	if err := json.Unmarshal(raw, &generic); err != nil {
		return "", err
	}
	canonical, err := json.Marshal(generic)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(canonical)
	return hex.EncodeToString(sum[:]), nil
}
