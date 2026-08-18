package wire

// The Agent reporting what it has installed. A Agent holds many Plugins but
// exactly one version of each, so a report for a Plugin already present is an
// in-place replacement, never a second row.

import "github.com/easyssl-io/easyssl-contract/plugincontract"

// RouteAgentPlugin is RouteAgentPlugins + "/{pluginID}": PUT reports an
// install or in-place upgrade of that Plugin. It carries the per-Agent bearer
// token.
const RouteAgentPlugin = "/api/v1/agent/plugins/{pluginID}"

// SettingDeclaration is the declaration the Plugin produced and the Agent
// passes on verbatim. One definition on purpose: the report must be exactly
// what `describe` answered, and a second struct would let the two drift.
type SettingDeclaration = plugincontract.SettingDeclaration

// PluginReport is what the Agent sends after installing or upgrading a
// Plugin: the version and artifact digest identifying the build, and that
// build's declared settings. The Server stores it and answers every later
// question from the store — nothing ever queries the host on demand, so the
// web console works whether or not the host is reachable.
type PluginReport struct {
	Version string `json:"version"`
	// Description is the build's own account of what it installs onto, passed
	// on verbatim from its describe answer like the declarations below.
	Description string `json:"description,omitempty"`
	// Digest is the SHA-256 of the installed artifact, as recorded from the
	// catalog at install time. Identity metadata: it says which bytes this
	// host runs.
	Digest       string               `json:"digest"`
	Declarations []SettingDeclaration `json:"declarations"`
}
