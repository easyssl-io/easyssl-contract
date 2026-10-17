package wire

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"time"

	"github.com/easyssl-io/easyssl-contract/l10n"
	"github.com/easyssl-io/easyssl-contract/plugincontract"
)

// RouteAgentInventory is where an Agent PUTs its whole host inventory,
// replacing what the Server holds for it. Sent only when the check-in answers
// that the Server wants it (CheckinResponse.ReportInventory), so the check-in
// stays constant-size however large a host's state grows.
const RouteAgentInventory = "/api/v1/agent/inventory"

// SettingDeclaration is the declaration a Plugin produced and the inventory
// passes on verbatim. One definition on purpose: a second struct would let the
// report drift from what the Plugin answered.
type SettingDeclaration = plugincontract.SettingDeclaration

// Inventory is everything an Agent observes about its own host, recomputed on
// every check-in and never persisted on the host. Nothing Server-owned is in
// it, and no private key: KeyInfo carries a fingerprint.
//
// The document replaces: an entry absent from it is one the host no longer has,
// which is what lets a deleted key stop being a holder. An Agent that failed to
// read any contributor must therefore not send the document at all.
type Inventory struct {
	Diagnostics Diagnostics `json:"diagnostics"`
	// Plugins is every Plugin executable on the host with what it answered to
	// discover this run. A Plugin absent here is one the host no longer has,
	// and the Server removes it with everything it reported.
	Plugins []PluginInventory `json:"plugins"`
	Keys    []KeyInfo         `json:"keys"`
}

// PluginInventory is one installed Plugin as the host sees it now: which bytes
// it runs, and what those bytes answered to discover.
type PluginInventory struct {
	// ID is the Plugin's name, which is also its executable's name in the
	// Plugin directory.
	ID string `json:"id"`
	// Digest is the SHA-256 of the executable as it is on disk at this run,
	// not the catalog's figure recorded at install: the Server compares the
	// two, and a mismatch is a binary someone replaced by hand.
	Digest string `json:"digest"`
	// Unobserved says describe or discover produced nothing usable this run, a
	// non-zero exit or the timeout, or the executable changed between the two,
	// so the fields below are empty and Warnings names the failure. The Server
	// keeps the last answer it received and shows the warning; the Agent keeps
	// nothing.
	Unobserved bool `json:"unobserved,omitempty"`
	// The describe answer (plugincontract.DescribeResponse) and the discover
	// answer (plugincontract.DiscoverResponse) of the same run, verbatim.
	Version           string                            `json:"version,omitempty"`
	Description       l10n.Text                         `json:"description,omitempty"`
	Declarations      []SettingDeclaration              `json:"declarations,omitempty"`
	DiscoveredTargets []plugincontract.DiscoveredTarget `json:"discoveredTargets,omitempty"`
	Warnings          []l10n.Text                       `json:"warnings,omitempty"`
}

// KeyInfo is one private key in the Agent's keystore, by its public half only.
type KeyInfo struct {
	// Fingerprint is the SHA-256 of the key's SubjectPublicKeyInfo, lowercase
	// hex, which is also the key file's name on the host and what the Server
	// computes from a CSR or certificate without seeing the key.
	Fingerprint string `json:"fingerprint"`
	// Algorithm is the key type, as KeyTypeOf names it.
	Algorithm string `json:"algorithm"`
	// CreatedAt is when the key file was created, as the host reports it.
	CreatedAt time.Time `json:"createdAt"`
}

// InventoryFingerprint is the one hash both sides compute over an Inventory:
// the Agent to put on its check-in, the Server to store beside the document and
// compare against the next check-in. Both call this function, because if the
// two canonicalized differently every check-in would ask for a report, a
// permanent full-inventory push that nothing reports as broken.
//
// Canonical form: keys sorted by fingerprint, Plugins by id, then the JSON
// encoding with object members in sorted order, which encoding/json gives for a
// map. Times are compared as RFC 3339 in UTC so a document that round-tripped
// through the Server's clock-free storage hashes the same.
func InventoryFingerprint(inv Inventory) (string, error) {
	keys := make([]KeyInfo, len(inv.Keys))
	copy(keys, inv.Keys)
	sort.Slice(keys, func(i, j int) bool { return keys[i].Fingerprint < keys[j].Fingerprint })
	for i := range keys {
		keys[i].CreatedAt = keys[i].CreatedAt.UTC().Truncate(time.Second)
	}
	plugins := make([]PluginInventory, len(inv.Plugins))
	copy(plugins, inv.Plugins)
	sort.Slice(plugins, func(i, j int) bool { return plugins[i].ID < plugins[j].ID })
	canon := Inventory{Diagnostics: inv.Diagnostics, Plugins: plugins, Keys: keys}
	if canon.Keys == nil {
		canon.Keys = []KeyInfo{}
	}
	if canon.Plugins == nil {
		canon.Plugins = []PluginInventory{}
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
