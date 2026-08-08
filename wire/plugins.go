package wire

// The Client reporting what it has installed. A Client holds many Plugins but
// exactly one version of each, so a report for a Plugin already present is an
// in-place replacement, never a second row.

import "github.com/easyssl-io/easyssl-contract/plugincontract"

// RouteClientPlugin is RouteClientPlugins + "/{pluginID}": PUT reports an
// install or in-place upgrade of that Plugin. It carries the per-Client bearer
// token.
const RouteClientPlugin = "/api/v1/client/plugins/{pluginID}"

// SettingDeclaration is the declaration the Plugin produced and the Client
// passes on verbatim. One definition on purpose: the report must be exactly
// what `describe` answered, and a second struct would let the two drift.
type SettingDeclaration = plugincontract.SettingDeclaration

// PluginReport is what the Client sends after installing or upgrading a
// Plugin: the version and artifact digest identifying the build, and that
// build's declared settings. The Server stores it and answers every later
// question from the store — nothing ever queries the host on demand, so the
// web console works whether or not the host is reachable.
type PluginReport struct {
	Version string `json:"version"`
	// Digest is the SHA-256 of the installed artifact, as recorded from the
	// catalog at install time. Identity metadata: it says which bytes this
	// host runs.
	Digest       string               `json:"digest"`
	Declarations []SettingDeclaration `json:"declarations"`
}
