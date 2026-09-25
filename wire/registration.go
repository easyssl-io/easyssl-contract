package wire

// Agent registration, check-in, and unregistration wire types. These cross the
// public Agent↔Server boundary, so keep them dependency-free and additive:
// Agents cannot be force-upgraded, so new fields are optional and existing
// ones never change meaning.

// Agent-facing route paths. None carries the web console session cookie:
// registration is gated by the shared Organization registration key, and the
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
	// deactivates it, which is how an uninstalling Agent tells the Server its
	// token is no longer in use. Deactivation invalidates the token and
	// nothing else: the Agent's Targets and record stay, and a reactivation
	// token minted in the web console brings the same Agent back on the same
	// or another host. Removing the Agent for good is the operator's call in
	// the web console, never something a host can do to its own record.
	RouteRegistration = "/api/v1/agent/registration"
)

// Error codes returned in the ErrorResponse.Code field so the Agent can react
// programmatically rather than parsing prose.
const (
	// CodeInvalidRegistrationKey: the presented registration key matched no
	// Organization (registration rejected, 401).
	CodeInvalidRegistrationKey = "invalid_registration_key"
	// CodeInvalidReactivationToken: the presented reactivation token matched
	// no deactivated Agent, has expired, or was already used. A reactivation
	// token is single-use and short-lived, so the operator mints another in
	// the web console rather than the host retrying (401).
	CodeInvalidReactivationToken = "invalid_reactivation_token"
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

// RegistrationRequest is the Agent's first contact. It carries exactly one
// credential. With the Organization's registration key it proposes a name and
// the Server mints a new pending Agent, with a public identifier (`agnt-` plus
// 14 base32 characters) it returns; that identifier is the same one the web
// console shows, and an Agent has no separate machine-facing identity. With a
// reactivation token, minted in the web console on a deactivated Agent, it
// resumes that Agent instead: the identity, name, Targets and everything the
// Server recorded are the deactivated Agent's, ProposedName is ignored, and the
// Agent lands active at once, because minting the token was the operator's
// review. A host therefore never presents an Agent ID of its own: only the
// Server names Agents, and a host that wants to be one it was before proves it
// with a token the operator handed it.
type RegistrationRequest struct {
	RegistrationKey   string      `json:"registrationKey,omitempty"`
	ReactivationToken string      `json:"reactivationToken,omitempty"`
	ProposedName      string      `json:"proposedName"`
	AgentVersion      string      `json:"agentVersion"`
	ProtocolVersion   string      `json:"protocolVersion"`
	Diagnostics       Diagnostics `json:"diagnostics"`
}

// RegistrationResponse carries the one-time bearer token (transmitted only
// here), the Agent's public ID, the status the Agent starts in (pending after a
// registration key, active after a reactivation token), the poll cadence the
// Agent should use, and the web console Marketplace URL to show before the
// operator may interrupt the authorization wait. MarketplaceURL is optional
// for compatibility with older Servers.
type RegistrationResponse struct {
	AgentToken string `json:"agentToken"`
	AgentID    string `json:"agentId"`
	Status     string `json:"status"`
	// DisplayName is what the web console calls this Agent. After a
	// reactivation it is the deactivated Agent's name, which the host did not
	// choose and has no other way to learn; after a fresh registration it is
	// the name the host proposed.
	DisplayName string `json:"displayName,omitempty"`
	// RequiredPlugins names the Plugins this Agent's Targets are configured
	// with, so a reactivated host can offer to install what its Deployments
	// will need before the first push arrives and fails for want of one. Empty
	// after a fresh registration, which has no Targets yet. Each entry is a
	// Plugin id as the Marketplace catalog names it.
	RequiredPlugins []string `json:"requiredPlugins,omitempty"`
	NextPollSeconds int      `json:"nextPollSeconds"`
	MarketplaceURL  string   `json:"marketplaceUrl,omitempty"`
	// OrganizationName is what the web console calls the Organization this host now
	// belongs to. Sent so the host can say where it is registered without a
	// second call, in the installer's own output, in `easyssl-agent status` and
	// in the uninstall prompt, which otherwise says "the Organization" to
	// somebody who may administer several.
	OrganizationName string `json:"organizationName,omitempty"`
	// OrganizationID is the web console's public id for it (`org-…`). Sent beside
	// the name because the name is not unique: every account starts with one
	// called "Default organization", so the name alone cannot tell an operator
	// which one a host joined.
	OrganizationID string `json:"organizationId,omitempty"`
}

// CheckinRequest is the authenticated poll body; the bearer token travels in
// the Authorization header, not here.
//
// Both fields are what this host is running now, and both are reported on every
// poll rather than at registration alone. A host is upgraded in place by
// re-running the installer, which keeps its token and therefore never
// re-registers, so the check-in is the only route by which a new version can
// reach the Server. Without it the web console shows the version the host had
// when it first registered, under a footnote saying the Agent reported it.
//
// An Agent older than this field sends no ProtocolVersion at all. The Server
// therefore keeps what it stored rather than overwriting it with the empty
// string, which is the whole of what makes the addition backward compatible.
type CheckinRequest struct {
	AgentVersion    string `json:"agentVersion"`
	ProtocolVersion string `json:"protocolVersion,omitempty"`
	// InventoryFingerprint is InventoryFingerprint() over the host inventory
	// the Agent computed for this check-in. The Server compares it with the
	// one it stored and answers ReportInventory when they differ, so the
	// check-in carries the signal and RouteAgentInventory the payload. Empty
	// from an Agent older than the field, or from the installer's own
	// check-in, which computes no inventory; the Server then asks the service
	// for one on its next check-in and nothing is lost.
	InventoryFingerprint string `json:"inventoryFingerprint,omitempty"`
}

// RegistrationStatus is the read-only view of an Agent's own registration,
// returned by GET RouteRegistration. Reading is deliberately free of side
// effects: unlike a check-in it does NOT refresh last-seen, so an Agent (or an
// operator running a status command) can ask "what am I?" without making a
// stopped Agent look alive in the web console.
type RegistrationStatus struct {
	Status      string `json:"status"`
	DisplayName string `json:"displayName"`
	// AgentID is the Agent's public id, which the host keeps nowhere and
	// learns here, so `easyssl-agent status` can name what the web console
	// calls this host.
	AgentID string `json:"agentId,omitempty"`
	// OrganizationName as above: an Agent registered before this field existed
	// learns it here, on the first read after the Server is updated.
	OrganizationName string `json:"organizationName,omitempty"`
	OrganizationID   string `json:"organizationId,omitempty"`
}

// CheckinResponse reports the current status, the display name the web console
// shows, and the next poll cadence (short while pending, longer once active).
type CheckinResponse struct {
	Status          string `json:"status"`
	DisplayName     string `json:"displayName"`
	NextPollSeconds int    `json:"nextPollSeconds"`
	// OrganizationName as above, so a rename in the web console reaches the host on
	// its next check-in rather than staying wrong until it re-registers. The id
	// never changes, and rides along so a host registered before these fields
	// existed learns both at once.
	OrganizationName string `json:"organizationName,omitempty"`
	OrganizationID   string `json:"organizationId,omitempty"`
	// ReportInventory asks the Agent to PUT its host inventory now: the
	// fingerprint it sent differs from the one stored, or the Server holds no
	// inventory for it at all, which is every Agent's first check-in after a
	// Server that lost its data. Absent (false) otherwise, and from a Server
	// older than the field, so an Agent never sends a document nobody asked
	// for.
	ReportInventory bool `json:"reportInventory,omitempty"`
	// PendingCommands is how many commands the Server is still waiting for this
	// Agent to finish, counted at the moment of the check-in. It exists so an
	// idle Agent can skip the reconcile fetch that follows a check-in, which on
	// a fleet that is mostly idle is one request per Agent per cadence
	// answering with an empty list.
	//
	// It is a count and deliberately not the commands themselves. The list is
	// fetched from RouteCommands and nowhere else, so there is exactly one
	// authoritative read of what is outstanding; a copy carried here could
	// disagree with it, which is the same reason the event stream carries an id
	// rather than the work.
	//
	// A pointer, because nil and zero mean opposite things. nil is a Server
	// that does not report this, and the Agent must fetch as it always did; 0
	// is a Server saying there is nothing, and the Agent may skip. Reading a
	// missing field as zero would make a newer Agent stop fetching from an
	// older Server, which is exactly the break this protocol's additive rule
	// exists to prevent.
	//
	// It is advisory in one direction only: it may be stale by the time the
	// Agent reads it, because a command can be created immediately after the
	// count. That costs at most one cadence, since the event stream nudges and
	// the next check-in counts again. It is never used to conclude that a
	// command exists, only that none did.
	//
	// Absent while the Agent is pending: every command route refuses a pending
	// Agent, so it has nothing to fetch and nothing to skip.
	PendingCommands *int `json:"pendingCommands,omitempty"`
}

// ErrorResponse is the JSON body returned with a non-2xx status; Code is one of
// the constants above.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
}
