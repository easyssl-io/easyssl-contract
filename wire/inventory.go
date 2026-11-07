package wire

import (
	"encoding/json"
	"time"

	"github.com/easyssl-io/easyssl-contract/l10n"
)

// Inventory is everything an Agent observes about its own host, recomputed on
// every check-in and never persisted on the host. Nothing Server-owned is in
// it, and no private key: KeyInfo carries the SHA-256 of its public half.
//
// The document replaces: an entry absent from it is one the host no longer has,
// which is what lets a deleted key stop being a holder. An Agent that failed to
// read any contributor must therefore not send the document at all.
type Inventory struct {
	// Fingerprint is the Agent's own hash of the rest of the document, 64
	// lowercase hex characters, the same value its check-in carries in
	// CheckinRequest.InventoryFingerprint. The Server stores it as sent and
	// never computes one, so how the Agent hashes is the Agent's business and
	// no field added to this document can make the two sides disagree. It
	// only decides whether this Agent sends its own document again.
	Fingerprint string      `json:"fingerprint"`
	Diagnostics Diagnostics `json:"diagnostics"`
	// Plugins is every Plugin executable on the host with what it answered
	// this run. A Plugin absent here is one the host no longer has, and the
	// Server removes it with everything it reported.
	Plugins []PluginInventory `json:"plugins"`
	Keys    []KeyInfo         `json:"keys"`
}

// PluginInventory is one installed Plugin as the host sees it now: which bytes
// it runs, and what those bytes answered.
type PluginInventory struct {
	// ID is the Plugin's name, which is also its executable's name in the
	// Plugin directory.
	ID string `json:"id"`
	// SHA256 is the SHA-256 of the executable as it is on disk at this run,
	// lowercase hex, not the catalog's figure recorded at install: the Server
	// compares the two, and a mismatch is a binary someone replaced by hand.
	// Every hash field of the contract is named after its algorithm.
	SHA256 string `json:"sha256"`
	// Describe and Discover are the Plugin's answers to OpDescribe and
	// OpDiscover of this run, byte for byte as it printed them
	// (plugincontract.DescribeResponse and DiscoverResponse). The Agent checks
	// a decoded copy and forwards the bytes, so a field a newer Plugin adds
	// reaches the Server through an Agent that does not know it. Both are
	// absent when Unobserved.
	Describe json.RawMessage `json:"describe,omitempty"`
	Discover json.RawMessage `json:"discover,omitempty"`
	// Unobserved says describe or discover produced nothing usable this run, a
	// non-zero exit, the timeout or an answer that failed the contract's
	// checks, or the executable changed between the two. The Server keeps the
	// last answers it received and shows Warnings; the Agent keeps nothing.
	Unobserved bool `json:"unobserved,omitempty"`
	// Warnings is the Agent's own account of why the Plugin is Unobserved, in
	// every language. A Plugin's own warnings are inside Discover.
	Warnings []l10n.Text `json:"warnings"`
}

// KeyInfo is one private key in the Agent's keystore, by its public half only.
type KeyInfo struct {
	// SPKISHA256 is the SHA-256 of the key's SubjectPublicKeyInfo, lowercase
	// hex, which is also the key file's name on the host and what the Server
	// computes from a CSR or certificate without seeing the key.
	SPKISHA256 string `json:"spkiSha256"`
	// KeyType is the key type, as KeyTypeOf names it, the same value a
	// CSRWork's KeyType asks for.
	KeyType string `json:"keyType"`
	// CreatedAt is when the key file was created, as the host reports it.
	CreatedAt time.Time `json:"createdAt"`
}
