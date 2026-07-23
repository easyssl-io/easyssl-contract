package wire

// Client registration and check-in wire types (S4 US1). These cross the
// public Client↔Server boundary, so keep them dependency-free and additive
// (Principle VI): new fields are optional, existing ones never change meaning.

// Client-facing route paths. The Client posts registration and check-in here;
// neither carries the session cookie (registration is gated by the shared
// organization registration key, check-in by the per-Client bearer token).
const (
	// RouteRegistrations is the check-in that creates a pending Client and, on
	// success, returns the one-time bearer token.
	RouteRegistrations = "/api/v1/client/registrations"
	// RouteCheckin is the authenticated poll: while pending it reports status,
	// and once active it also refreshes the Client's last-seen.
	RouteCheckin = "/api/v1/client/checkin"
)

// Error codes returned in the ErrorResponse.Code field so the Client can react
// programmatically rather than parsing prose.
const (
	// CodeInvalidRegistrationKey: the presented registration key matched no
	// organization (registration rejected, 401).
	CodeInvalidRegistrationKey = "invalid_registration_key"
	// CodeInstanceIDTaken: the instance ID is already registered; the installer
	// regenerates it and retries (409).
	CodeInstanceIDTaken = "instance_id_taken"
	// CodeUnknownToken: the bearer token is not recognized — revoked, or the
	// registration was rejected (a Client cannot distinguish the two, by
	// design) (401).
	CodeUnknownToken = "unknown_token"
)

// Client lifecycle states as reported to the Client over the wire.
const (
	StatusPending = "pending"
	StatusActive  = "active"
)

// Diagnostics is the host information the Client captures at install and
// reports at registration (FR-004). Install-log reporting (FR-004a) is deferred
// with US4 and is intentionally absent here.
type Diagnostics struct {
	MachineID   string `json:"machineId"`
	Hostname    string `json:"hostname"`
	OS          string `json:"os"`
	HostType    string `json:"hostType"`
	BinaryPath  string `json:"binaryPath"`
	ConfigPath  string `json:"configPath"`
	ServiceName string `json:"serviceName"`
}

// RegistrationRequest is the Client's first contact: it proposes a name and
// identity and presents the organization's registration key.
type RegistrationRequest struct {
	RegistrationKey string      `json:"registrationKey"`
	InstanceID      string      `json:"instanceId"`
	ProposedName    string      `json:"proposedName"`
	AgentVersion    string      `json:"agentVersion"`
	ProtocolVersion string      `json:"protocolVersion"`
	Diagnostics     Diagnostics `json:"diagnostics"`
}

// RegistrationResponse carries the one-time bearer token (transmitted only
// here) plus the initial status and the poll cadence the Client should use.
type RegistrationResponse struct {
	ClientToken     string `json:"clientToken"`
	Status          string `json:"status"`
	NextPollSeconds int    `json:"nextPollSeconds"`
}

// CheckinRequest is the authenticated poll body; the bearer token travels in
// the Authorization header, not here.
type CheckinRequest struct {
	AgentVersion string `json:"agentVersion"`
}

// CheckinResponse reports the current status, the display name the console
// shows, and the next poll cadence (short while pending, longer once active).
type CheckinResponse struct {
	Status          string `json:"status"`
	DisplayName     string `json:"displayName"`
	NextPollSeconds int    `json:"nextPollSeconds"`
}

// ErrorResponse is the JSON body returned with a non-2xx status; Code is one of
// the constants above.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
}
