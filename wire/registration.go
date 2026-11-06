package wire

// Agent registration, check-in, and unregistration wire types. These cross the
// public Agent↔Server boundary, so keep them dependency-free and additive:
// Agents cannot be force-upgraded, so new fields are optional and existing
// ones never change meaning.

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
	// CodeContractUnsupported: the request's HeaderContractVersion is missing
	// or not plugincontract.Compatible with the Server's, on any route (400).
	// The Agent says so and that it needs updating, rather than retrying as if
	// the Server were down.
	CodeContractUnsupported = "contract_unsupported"
	// CodeInternalError: the Server failed at something that is not the Agent's
	// doing. Never a verdict on the Agent's credential: the Agent retries
	// later (500).
	CodeInternalError = "internal_error"
)

// Agent lifecycle states as reported to the Agent over the wire. An Agent that
// does not know a status a newer Server sends treats it as AgentStatusPending: it
// keeps checking in and does no work until it reads one it knows.
const (
	AgentStatusPending = "pending"
	AgentStatusActive  = "active"
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
	Diagnostics       Diagnostics `json:"diagnostics"`
}

// RegistrationResponse carries the one-time bearer token (transmitted only
// here), the Agent's public ID, the status the Agent starts in (pending after a
// registration key, active after a reactivation token), the poll cadence the
// Agent should use, and the web console Marketplace URL to show before the
// operator may interrupt the authorization wait.
type RegistrationResponse struct {
	AgentToken string `json:"agentToken"`
	AgentID    string `json:"agentId"`
	Status     string `json:"status"`
	// DisplayName is what the web console calls this Agent. After a
	// reactivation it is the deactivated Agent's name, which the host did not
	// choose and has no other way to learn; after a fresh registration it is
	// the name the host proposed.
	DisplayName string `json:"displayName"`
	// RequiredPlugins names the Plugins this Agent's Targets are configured
	// with, so a reactivated host can offer to install what its Deployments
	// will need before the first push arrives and fails for want of one. Empty
	// after a fresh registration, which has no Targets yet. Each entry is a
	// Plugin id as the Marketplace catalog names it.
	RequiredPlugins []string `json:"requiredPlugins"`
	NextPollSeconds int      `json:"nextPollSeconds"`
	MarketplaceURL  string   `json:"marketplaceUrl"`
	// OrganizationName is what the web console calls the Organization this host
	// now belongs to. Sent so the host can say where it is registered without a
	// second call: in the installer's output, in `easyssl-agent status` and in the
	// uninstall prompt. Omitted, like OrganizationID, when the Server could not
	// read it: a message on the host never fails the exchange that carries it.
	OrganizationName string `json:"organizationName,omitempty"`
	// OrganizationID is the web console's public id for it (`org-…`). Sent beside
	// the name because the name is not unique: two Organizations may carry the
	// same one, so the name alone cannot tell an operator which one a host
	// joined.
	OrganizationID string `json:"organizationId,omitempty"`
}

// CheckinRequest is the authenticated poll body; the bearer token travels in
// the Authorization header, not here. The Agent's versions travel in headers
// on every request (HeaderContractVersion), and the Server records them from
// the check-in: a host upgraded in place by re-running the installer keeps its
// token and never re-registers.
type CheckinRequest struct {
	// InventoryFingerprint is the fingerprint of the host inventory the Agent
	// computed for this check-in, as it will send it in Inventory.Fingerprint.
	// The Server compares it with the one it stored and answers ReportInventory
	// when they differ. Empty from the installer's own check-in, which computes
	// no inventory; the Server then asks the service for one on its next
	// check-in.
	InventoryFingerprint string `json:"inventoryFingerprint,omitempty"`
}

// RegistrationStatus is the read-only view of an Agent's own registration,
// returned by GET RouteAgentMe. Reading is deliberately free of side
// effects: unlike a check-in it does NOT refresh last-seen, so an Agent (or an
// operator running a status command) can ask "what am I?" without making a
// stopped Agent look alive in the web console.
type RegistrationStatus struct {
	Status      string `json:"status"`
	DisplayName string `json:"displayName"`
	// AgentID is the Agent's public id, which the host keeps nowhere and
	// learns here, so `easyssl-agent status` can name what the web console
	// calls this host.
	AgentID          string `json:"agentId"`
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
	// its next check-in rather than staying wrong until it re-registers.
	OrganizationName string `json:"organizationName,omitempty"`
	OrganizationID   string `json:"organizationId,omitempty"`
	// ReportInventory asks the Agent to PUT its host inventory now: the
	// fingerprint it sent differs from the one stored, or the Server holds no
	// inventory for it at all, which is every Agent's first check-in after a
	// Server that lost its data. Absent (false) otherwise, so an Agent never
	// sends a document nobody asked for.
	ReportInventory bool `json:"reportInventory,omitempty"`
	// PendingCommands is how many commands the Server is still waiting for this
	// Agent to finish, counted at the moment of the check-in, so an idle Agent can
	// skip the reconcile fetch that follows. It is a count and not the commands:
	// the list is fetched from RouteAgentCommands and nowhere else, so no copy can
	// disagree with it.
	//
	// It may be stale by the time the Agent reads it, because a command can be
	// created immediately after the count. That costs at most one cadence, since
	// the event stream nudges and the next check-in counts again. It is never used
	// to conclude that a command exists, only that none did.
	//
	// Zero while the Agent is pending: every command route refuses a pending
	// Agent.
	PendingCommands int `json:"pendingCommands"`
}

// ErrorResponse is the JSON body returned with a non-2xx status; Code is one of
// the constants above.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
}
