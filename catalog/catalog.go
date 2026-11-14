// Package catalog is the Marketplace: the one shape of the catalog a Server
// publishes, which an Agent reads before installing a Plugin and a browser
// reads to show what can be installed. It is part of the contract because
// both an Agent and a Server depend on it, and the paths it names are routes
// of the Server like the Agent's own.
package catalog

import (
	"net/url"
	"strings"

	"github.com/easyssl-io/easyssl-contract/l10n"
	"github.com/easyssl-io/easyssl-contract/plugincontract"
)

// The Marketplace routes, relative to wire.APIPrefix like every other route
// of the Server. None needs a credential: a host reads them before it has
// one. Only the installer itself, /install, stays outside the prefix, being
// the URL customers paste.
const (
	// RouteCatalog answers Response.
	RouteCatalog = "/marketplace/catalog"
	// RoutePluginArtifact serves one Plugin's executable for one platform.
	RoutePluginArtifact = "/download/plugins/{id}/{os}/{arch}"
	// RouteAgentArtifact serves the Agent's executable for one platform.
	RouteAgentArtifact = "/download/easyssl-agent/{os}/{arch}"
)

// PartyFirst is the party of a Plugin easySSL publishes itself, the only
// party an Agent installs.
const PartyFirst = "first"

// Response is what RouteCatalog answers: every Plugin the Server offers, and
// the Agent it hands out.
type Response struct {
	Plugins []Plugin `json:"plugins"`
	// Agent is the Agent build the Server hands out, which an Agent reads to
	// update itself; absent when the Server holds none.
	//
	// Its shape is frozen across breaking versions: the catalog and the
	// downloads are not behind the contract gate, so this entry is how an Agent
	// a Server no longer speaks to reaches one it does, and a later release
	// changing it would take that away from the Agents that need it most.
	Agent *Agent `json:"agent,omitempty"`
	// AgentReleases are the Agent's finished releases, newest first, which the
	// Marketplace lists beside the build handed out now. Outside Agent, whose
	// shape is frozen.
	AgentReleases []Release `json:"agentReleases,omitempty"`
}

// AgentUpdateSince is the first contract version whose Agents update
// themselves from Response.Agent (`easyssl-agent update`); an Agent built
// against an older one is updated by running the installer again.
const AgentUpdateSince = "0.2.1"

// Agent is the Agent build a Server hands out: its version, and one artifact
// per platform it holds a build for, each served at RouteAgentArtifact.
type Agent struct {
	Version   string     `json:"version"`
	Artifacts []Artifact `json:"artifacts"`
}

// Plugin is one Plugin the Server offers. Its version, description and
// declarations are its own `describe` answer; the rest is about distributing
// it.
type Plugin struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	// ContractVersion is the contract version the build was made against, from
	// its `describe` answer, so an Agent about to replace itself can tell
	// whether the new build would still run this Plugin's. Empty from a Server
	// older than the field.
	ContractVersion string    `json:"contractVersion,omitempty"`
	Description     l10n.Text `json:"description"`
	// Declarations are the Plugin's own, so a reader can see what a Target of
	// it is configured with before anything is installed.
	Declarations []plugincontract.SettingDeclaration `json:"declarations"`
	// Party is who publishes the Plugin; an Agent installs only PartyFirst.
	Party string `json:"party"`
	// Publisher and SourceURL answer who released this build and where its
	// code is, which is what an install decision is made on.
	Publisher string `json:"publisher"`
	SourceURL string `json:"sourceUrl"`
	// Artifacts are the builds the Server holds, one per platform it has one
	// for; a platform with none is absent rather than offered and broken.
	Artifacts []Artifact `json:"artifacts"`
	// Releases are the Plugin's finished releases, newest first, which the
	// Marketplace lists beside the build handed out now. The build handed out
	// is one of them when its Version is; a development build is none.
	Releases []Release `json:"releases,omitempty"`
}

// Release is one finished release of a Plugin or of the Agent, as the release
// recorded it in the `releases.json` kept with its code: a JSON array of these,
// newest first, which a Server carries and lists in the catalog. Nothing in it
// is downloadable; only the build a Server holds is.
type Release struct {
	Version string `json:"version"`
	// Date is the day it was released, YYYY-MM-DD in UTC.
	Date            string `json:"date"`
	ContractVersion string `json:"contractVersion"`
	// Description and Declarations are a Plugin's `describe` answer at this
	// release, so a reader sees what each version declared; absent for the
	// Agent.
	Description  l10n.Text                           `json:"description,omitempty"`
	Declarations []plugincontract.SettingDeclaration `json:"declarations,omitempty"`
	// Builds are the release's published builds, one per platform, by the
	// SHA-256 of their bytes: a host's build with one of these is that release.
	// Absent for a release recorded after the fact, whose bytes were never
	// published.
	Builds []Build `json:"builds,omitempty"`
}

// Build is one platform's build of a Release, named by its digest.
type Build struct {
	OS     string `json:"os"`
	Arch   string `json:"arch"`
	SHA256 string `json:"sha256"`
}

// Artifact is one build of a Plugin, or of the Agent, for one platform.
type Artifact struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
	// SHA256 is the SHA-256 of the artifact's bytes, lowercase hex, which the
	// Agent checks before it installs anything.
	SHA256 string `json:"sha256"`
	// URL is where the artifact is served, a path on the same Server (under
	// wire.APIPrefix): an Agent resolves it against the Server it already
	// talks to and follows nothing else.
	URL string `json:"url"`
}

// PluginArtifactPath is RoutePluginArtifact with its placeholders filled,
// path-escaped.
func PluginArtifactPath(id, os, arch string) string {
	return fill(RoutePluginArtifact, "{id}", id, "{os}", os, "{arch}", arch)
}

// AgentArtifactPath is RouteAgentArtifact with its placeholders filled,
// path-escaped.
func AgentArtifactPath(os, arch string) string {
	return fill(RouteAgentArtifact, "{os}", os, "{arch}", arch)
}

func fill(route string, pairs ...string) string {
	for i := 0; i+1 < len(pairs); i += 2 {
		route = strings.Replace(route, pairs[i], url.PathEscape(pairs[i+1]), 1)
	}
	return route
}
