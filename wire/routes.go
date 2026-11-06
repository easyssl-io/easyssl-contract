package wire

import (
	"net/url"
	"strings"
)

// APIPrefix is where every route below is mounted on the Server. The routes
// are relative to it, so the Server registers them as they are on its API
// router and the Agent requests APIPrefix + Path(route, …).
const APIPrefix = "/api/v1"

// Agent routes. Each is named Route followed by its path after APIPrefix, in
// camel case and without placeholders, so no two names differ by a letter
// their paths do not. "{id}" is filled by Path. None carries the web console's
// session cookie: RouteAgentRegistrations is gated by the Organization's
// registration key or a reactivation token, every other route by the Agent's
// bearer token, and every route by the contract version (HeaderContractVersion).
const (
	// RouteAgentRegistrations creates a pending Agent, or brings a deactivated
	// one back with a reactivation token, and answers the one-time bearer
	// token.
	RouteAgentRegistrations = "/agent/registrations"
	// RouteAgentCheckin is the authenticated poll: while pending it reports
	// status, and once active it also refreshes the Agent's last-seen.
	RouteAgentCheckin = "/agent/checkin"
	// RouteAgentMe is the Agent's own registration. GET reads it without
	// refreshing last-seen; DELETE deactivates it, which is how an
	// uninstalling Agent tells the Server its token is no longer in use.
	// Deactivation invalidates the token and nothing else: the Agent's Targets
	// and record stay, and a reactivation token minted in the web console
	// brings the same Agent back. Removing the Agent for good is the
	// operator's call in the web console, never something a host can do to
	// its own record.
	RouteAgentMe = "/agent/me"
	// RouteAgentInventory is where an Agent PUTs its whole host inventory,
	// replacing what the Server holds for it. Sent only when the check-in
	// answers that the Server wants it (CheckinResponse.ReportInventory), so
	// the check-in stays constant-size however large a host's state grows.
	RouteAgentInventory = "/agent/inventory"
	// RouteAgentEvents is the long-lived stream (SSE). One connection per Agent
	// carries every command type, so an Agent behind a restrictive network has
	// a single thing to keep open.
	RouteAgentEvents = "/agent/events"
	// RouteAgentCommands lists the commands an Agent has not finished yet. The
	// Agent reads it on every (re)connect and on every poll, which is what
	// actually guarantees a command runs.
	RouteAgentCommands = "/agent/commands"
	// RouteAgentCommandAck is the Agent reporting that it has finished with a
	// command, one way or another.
	RouteAgentCommandAck = "/agent/commands/{id}/ack"
	// RouteAgentCertificate is where the Agent reads a Certificate's signing
	// work, "{id}" being Command.RefID. Answered whatever the Certificate's
	// state, so a re-run can re-read it.
	RouteAgentCertificate = "/agent/certificates/{id}"
	// RouteAgentCertificateCSR is where the Agent submits the CSR it built.
	// Stored write-once; the ACME order opens from here.
	RouteAgentCertificateCSR = "/agent/certificates/{id}/csr"
	// RouteAgentDeployment is where the Agent reads a Deployment's work: the
	// material, and the resolved value of every setting the Plugin will be
	// called with. A Deployment already finished answers with its state and no
	// material, so a re-delivered command is acknowledged without re-running.
	// Both Deployment routes answer only the Agent the Deployment names; to
	// any other it is indistinguishable from one that does not exist.
	RouteAgentDeployment = "/agent/deployments/{id}"
	// RouteAgentDeploymentResult is the Agent's terminal report for a
	// Deployment, write-once.
	RouteAgentDeploymentResult = "/agent/deployments/{id}/result"
)

// Path is route with its "{id}" placeholder replaced by id, path-escaped, so
// an id can never add a segment. Every Agent request is built through it.
func Path(route, id string) string {
	return strings.Replace(route, "{id}", url.PathEscape(id), 1)
}
