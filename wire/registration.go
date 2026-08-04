package wire

// Client registration, check-in, and unregistration wire types. These cross the
// public Client↔Server boundary, so keep them dependency-free and additive:
// Clients cannot be force-upgraded, so new fields are optional and existing
// ones never change meaning.

// Client-facing route paths. None carries the console session cookie:
// registration is gated by the shared organization registration key, and the
// other routes by the per-Client bearer token.
const (
	// RouteRegistrations is the check-in that creates a pending Client and, on
	// success, returns the one-time bearer token.
	RouteRegistrations = "/api/v1/client/registrations"
	// RouteCheckin is the authenticated poll: while pending it reports status,
	// and once active it also refreshes the Client's last-seen.
	RouteCheckin = "/api/v1/client/checkin"
	// RouteRegistration is the Client's own registration. GET reads it, so a
	// Client can ask what the Server currently considers it to be; DELETE
	// removes it, which is how a Client unregisters itself when uninstalling,
	// dropping the Client and its token so the host disappears from the
	// organization without the operator removing it from the web console.
	RouteRegistration = "/api/v1/client/registration"
)

// Error codes returned in the ErrorResponse.Code field so the Client can react
// programmatically rather than parsing prose.
const (
	// CodeInvalidRegistrationKey: the presented registration key matched no
	// organization (registration rejected, 401).
	CodeInvalidRegistrationKey = "invalid_registration_key"
	// CodeClientIDTaken: the presented Client ID names a Client that cannot
	// be re-registered with this key — a live row, or a removed one claimed
	// with another organization's key (409).
	CodeClientIDTaken = "client_id_taken"
	// CodeUnknownClientID: the presented Client ID names no Client. Only the
	// Server mints identifiers, so an unknown one is stale local state (a
	// reset Server, a hand-edited config), never something to register under;
	// the host re-registers fresh after clearing it (404).
	CodeUnknownClientID = "unknown_client_id"
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

// RegistrationRequest is the Client's first contact: it proposes a name and
// presents the organization's registration key. ClientID is empty on a fresh
// registration — the Server mints the Client's public identifier (`clnt-` plus
// 14 base32 characters) and returns it — and carries the persisted identifier
// on a reinstall, so the Server updates that Client instead of creating a
// second one. It is the same identifier the console shows; a Client has no
// separate machine-facing identity.
type RegistrationRequest struct {
	RegistrationKey string      `json:"registrationKey"`
	ClientID        string      `json:"clientId,omitempty"`
	ProposedName    string      `json:"proposedName"`
	ClientVersion   string      `json:"clientVersion"`
	ProtocolVersion string      `json:"protocolVersion"`
	Diagnostics     Diagnostics `json:"diagnostics"`
}

// RegistrationResponse carries the one-time bearer token (transmitted only
// here), the Server-minted Client ID the Client persists, plus the initial
// status and the poll cadence the Client should use.
type RegistrationResponse struct {
	ClientToken     string `json:"clientToken"`
	ClientID        string `json:"clientId"`
	Status          string `json:"status"`
	NextPollSeconds int    `json:"nextPollSeconds"`
}

// CheckinRequest is the authenticated poll body; the bearer token travels in
// the Authorization header, not here.
type CheckinRequest struct {
	ClientVersion string `json:"clientVersion"`
}

// RegistrationStatus is the read-only view of a Client's own registration,
// returned by GET RouteRegistration. Reading is deliberately free of side
// effects: unlike a check-in it does NOT refresh last-seen, so a Client (or an
// operator running a status command) can ask "what am I?" without making a
// stopped Client look alive in the web console.
type RegistrationStatus struct {
	Status      string `json:"status"`
	DisplayName string `json:"displayName"`
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
