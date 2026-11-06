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
	// RouteRegistration is the Agent's own registration. GET reads it; DELETE
	// deactivates it, which is how an uninstalling Agent tells the Server its
	// token is no longer in use. Deactivation invalidates the token and nothing
	// else: the Agent's Targets and record stay, and a reactivation token minted
	// in the web console brings the same Agent back on the same or another host.
	// Removing the Agent for good is the operator's call in the web console, never
	// something a host can do to its own record.
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
	// CodeUnknownToken: the bearer token is not recognized, because it was revoked
	// or the registration was rejected. An Agent cannot distinguish the two, by
	// design (401).
	CodeUnknownToken = "unknown_token"
	// CodeAgentNotActive: the token is this Agent's, but the Agent is still
	// awaiting authorization, so no route other than check-in and the
	// registration read answers it yet (403). The check-in already tells a pending
	// Agent its status, so naming it here reveals nothing new.
	CodeAgentNotActive = "agent_not_active"
	// CodeInvalidRequest: the body is not JSON the route can read, or exceeds
	// the route's size limit (400).
	CodeInvalidRequest = "invalid_request"
	// CodeInternalError: the Server failed at something that is not the Agent's
	// doing. Never a verdict on the Agent's credential: the Agent retries
	// later (500).
	CodeInternalError = "internal_error"
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
// the Server mints a new pending Agent, whose public identifier (`agnt-` plus
// 14 base32 characters) is the one the web console shows; an Agent has no
// separate machine-facing identity. With a reactivation token, minted in the
// web console on a deactivated Agent, it resumes that Agent: the identity,
// name, Targets and everything the Server recorded are the deactivated Agent's,
// ProposedName is ignored, and the Agent lands active at once, because minting
// the token was the operator's review. A host never presents an Agent ID of its
// own.
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
	// OrganizationName is what the web console calls the Organization this host
	// now belongs to. Sent so the host can say where it is registered without a
	// second call: in the installer's output, in `easyssl-agent status` and in the
	// uninstall prompt.
	OrganizationName string `json:"organizationName,omitempty"`
	// OrganizationID is the web console's public id for it (`org-…`). Sent beside
	// the name because the name is not unique: two Organizations may carry the
	// same one, so the name alone cannot tell an operator which one a host
	// joined.
	OrganizationID string `json:"organizationId,omitempty"`
}

// CheckinRequest is the authenticated poll body; the bearer token travels in
// the Authorization header, not here.
//
// The two versions are what this host is running now, reported on every poll: a host
// upgraded in place by re-running the installer keeps its token and never
// re-registers, so the check-in is the only route by which a new version
// reaches the Server.
//
// An Agent older than ProtocolVersion sends none, and the Server then keeps
// what it stored rather than overwriting it with the empty string.
type CheckinRequest struct {
	AgentVersion    string `json:"agentVersion"`
	ProtocolVersion string `json:"protocolVersion,omitempty"`
	// InventoryFingerprint is InventoryFingerprint() over the host inventory the
	// Agent computed for this check-in. The Server compares it with the one it
	// stored and answers ReportInventory when they differ. Empty from an Agent
	// older than the field, or from the installer's own check-in, which computes
	// no inventory; the Server then asks the service for one on its next check-in.
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
	// Agent to finish, counted at the moment of the check-in, so an idle Agent can
	// skip the reconcile fetch that follows. It is a count and not the commands:
	// the list is fetched from RouteCommands and nowhere else, so no copy can
	// disagree with it.
	//
	// A pointer, because nil and zero mean opposite things. nil is a Server that
	// does not report this, and the Agent must fetch as it always did; 0 is a
	// Server saying there is nothing, and the Agent may skip. Reading a missing
	// field as zero would make a newer Agent stop fetching from an older Server.
	//
	// It may be stale by the time the Agent reads it, because a command can be
	// created immediately after the count. That costs at most one cadence, since
	// the event stream nudges and the next check-in counts again. It is never used
	// to conclude that a command exists, only that none did.
	//
	// Absent while the Agent is pending: every command route refuses a pending
	// Agent.
	PendingCommands *int `json:"pendingCommands,omitempty"`
}

// ErrorResponse is the JSON body returned with a non-2xx status; Code is one of
// the constants above.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
}
