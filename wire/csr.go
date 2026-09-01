package wire

// The request_csr resource: the certificate-signing work behind a
// CommandRequestCSR, fetched and answered over ordinary request/response
// routes. The command carries only the certificate id; these routes are where
// the Agent learns what to sign and hands back the result.
//
// Both routes carry the per-Agent bearer token and answer only the Agent
// named on the certificate. To any other Agent, its own Organization's
// included, the certificate is indistinguishable from one that does not exist,
// the same no-probing posture the check-in takes with unknown tokens.

// CSR route paths. "{id}" is the certificate id from Command.RefID.
const (
	// RouteAgentCertificate is where the named Agent reads its signing work.
	// Answered whatever the certificate's state, so a re-run can re-read it.
	RouteAgentCertificate = "/api/v1/agent/certificates/{id}"
	// RouteAgentCertificateCSR is where the named Agent submits the CSR it
	// built. Stored write-once; the ACME order opens from here.
	RouteAgentCertificateCSR = "/api/v1/agent/certificates/{id}/csr"
)

// KeyTypeECP256 is the one certificate key type the protocol names today. The
// field exists so adding another is a MINOR bump rather than a new route; an
// Agent seeing a type it cannot generate leaves the work unanswered and the
// command's expiry reports it (the Agent is older than the Server).
const KeyTypeECP256 = "ec_p256"

// CSRWork is what the named Agent is asked to sign: the domain set fixed when
// the certificate was requested, which the Agent signs exactly and never
// chooses. The CSR it builds carries these as SANs only, no Subject Common
// Name, because an ACME client orders the Common Name too and an added name
// there would be one nobody requested.
type CSRWork struct {
	CertificateID string   `json:"certificateId"`
	Domains       []string `json:"domains"`
	KeyType       string   `json:"keyType"`
}

// CSRSubmission carries the PEM CSR back. One CERTIFICATE REQUEST block,
// bounded small (a CSR is a few kilobytes); the Server parses defensively and
// verifies the content — SAN set, no CN, key type, self-signature — before
// storing, since the CSR and not the certificate record is what decides what
// gets issued.
type CSRSubmission struct {
	CSR string `json:"csr"`
}

// Error codes for the CSR submission. All three are terminal for the Agent:
// none can be fixed by retrying the same submission, so the Agent records the
// refusal and acks the command.
const (
	// CodeCSRAlreadySubmitted (409): a CSR is already stored. Treated as
	// success — an ECDSA re-signature never reproduces the same bytes, so an
	// Agent re-run after a crash cannot recognize its own submission, and the
	// stored CSR is from the same key anyway because key generation is
	// idempotent per certificate.
	CodeCSRAlreadySubmitted = "csr_already_submitted"
	// CodeNotAwaitingCSR (409): the certificate is failed or issued; the
	// instruction is stale and is discarded.
	CodeNotAwaitingCSR = "not_awaiting_csr"
	// CodeCSRMismatch (422): verification failed and nothing was stored. The
	// Server fails the certificate at the rejection, naming this Agent and
	// what failed — it cannot rely on the command expiring, because the Agent
	// acks a rejection as handled and an acked command never expires. A retry
	// issues a fresh command.
	CodeCSRMismatch = "csr_mismatch"
)
