package wire

// Agent registration, check-in, and unregistration wire types. These cross the
// public Agent↔Server boundary, so keep them dependency-free and additive:
// Agents cannot be force-upgraded, so new fields are optional and existing
// ones never change meaning.

// Agent-facing route paths. None carries the console session cookie:
// registration is gated by the shared organization registration key, and the
// other routes by the per-Agent bearer token.
const (
	// RouteRegistrations is the check-in that creates a pending Agent and, on
	// success, returns the one-time bearer token.
	RouteRegistrations = "/api/v1/agent/registrations"
	// RouteCheckin is the authenticated poll: while pending it reports status,
	// and once active it also refreshes the Agent's last-seen.
	RouteCheckin = "/api/v1/agent/checkin"
	// RouteRegistration is the Agent's own registration. GET reads it, so an
	// Agent can ask what the Server currently considers it to be; DELETE
	// removes it, which is how an Agent unregisters itself when uninstalling,
	// dropping the Agent and its token so the host disappears from the
	// organization without the operator removing it from the web console.
	RouteRegistration = "/api/v1/agent/registration"
)

// Error codes returned in the ErrorResponse.Code field so the Agent can react
// programmatically rather than parsing prose.
const (
	// CodeInvalidRegistrationKey: the presented registration key matched no
	// organization (registration rejected, 401).
	CodeInvalidRegistrationKey = "invalid_registration_key"
	// CodeAgentIDTaken: the presented Agent ID names an Agent that cannot
	// be re-registered with this key — a live row, or a removed one claimed
	// with another organization's key (409).
	CodeAgentIDTaken = "agent_id_taken"
	// CodeUnknownAgentID: the presented Agent ID names no Agent. Only the
	// Server mints identifiers, so an unknown one is stale local state (a
	// reset Server, a hand-edited config), never something to register under;
	// the host re-registers fresh after clearing it (404).
	CodeUnknownAgentID = "unknown_agent_id"
	// CodeUnknownToken: the bearer token is not recognized — revoked, or the
	// registration was rejected (an Agent cannot distinguish the two, by
	// design) (401).
	CodeUnknownToken = "unknown_token"
)

// Agent lifecycle states as reported to the Agent over the wire.
const (
	StatusPending = "pending"
	StatusActive  = "active"
)

// Diagnostics is the host information the Agent captures at install and
// reports at registration. Reporting the installation log is a later addition
// and is intentionally absent here.
type Diagnostics struct {
	Hostname    string `json:"hostname"`
	OS          string `json:"os"`
	HostType    string `json:"hostType"`
	BinaryPath  string `json:"binaryPath"`
	ConfigPath  string `json:"configPath"`
	ServiceName string `json:"serviceName"`
}

// RegistrationRequest is the Agent's first contact: it proposes a name and
// presents the organization's registration key. AgentID is empty on a fresh
// registration — the Server mints the Agent's public identifier (`agnt-` plus
// 14 base32 characters) and returns it — and carries the persisted identifier
// on a reinstall, so the Server updates that Agent instead of creating a
// second one. It is the same identifier the console shows; an Agent has no
// separate machine-facing identity.
type RegistrationRequest struct {
	RegistrationKey string      `json:"registrationKey"`
	AgentID         string      `json:"agentId,omitempty"`
	ProposedName    string      `json:"proposedName"`
	AgentVersion    string      `json:"agentVersion"`
	ProtocolVersion string      `json:"protocolVersion"`
	Diagnostics     Diagnostics `json:"diagnostics"`
}

// RegistrationResponse carries the one-time bearer token (transmitted only
// here), the Server-minted Agent ID the Agent persists, plus the initial
// status, the poll cadence the Agent should use, and the web-console
// Marketplace URL to show before the operator may interrupt the authorization
// wait. MarketplaceURL is optional for compatibility with older Servers.
type RegistrationResponse struct {
	AgentToken      string `json:"agentToken"`
	AgentID         string `json:"agentId"`
	Status          string `json:"status"`
	NextPollSeconds int    `json:"nextPollSeconds"`
	MarketplaceURL  string `json:"marketplaceUrl,omitempty"`
	// OrganizationName is what the console calls the organization this host now
	// belongs to. Sent so the host can say where it is registered without a
	// second call, in the installer's own output, in `easyssl-agent status` and
	// in the uninstall prompt, which otherwise says "the organization" to
	// somebody who may administer several.
	OrganizationName string `json:"organizationName,omitempty"`
}

// CheckinRequest is the authenticated poll body; the bearer token travels in
// the Authorization header, not here.
type CheckinRequest struct {
	AgentVersion string `json:"agentVersion"`
}

// RegistrationStatus is the read-only view of an Agent's own registration,
// returned by GET RouteRegistration. Reading is deliberately free of side
// effects: unlike a check-in it does NOT refresh last-seen, so an Agent (or an
// operator running a status command) can ask "what am I?" without making a
// stopped Agent look alive in the web console.
type RegistrationStatus struct {
	Status      string `json:"status"`
	DisplayName string `json:"displayName"`
	// OrganizationName as above: an Agent registered before this field existed
	// learns it here, on the first read after the Server is updated.
	OrganizationName string `json:"organizationName,omitempty"`
}

// CheckinResponse reports the current status, the display name the console
// shows, and the next poll cadence (short while pending, longer once active).
type CheckinResponse struct {
	Status          string `json:"status"`
	DisplayName     string `json:"displayName"`
	NextPollSeconds int    `json:"nextPollSeconds"`
	// OrganizationName as above, so a rename in the console reaches the host on
	// its next check-in rather than staying wrong until it re-registers.
	OrganizationName string `json:"organizationName,omitempty"`
}

// ErrorResponse is the JSON body returned with a non-2xx status; Code is one of
// the constants above.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
}
