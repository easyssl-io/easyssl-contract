package wire

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
)

// The request_csr resource: the certificate-signing work behind a
// CommandRequestCSR. The command carries only the Certificate id; these routes
// are where the Agent learns what to sign and hands back the result.
//
// Both routes carry the per-Agent bearer token and answer only the Agent named
// on the Certificate. To any other Agent, its own Organization's included, the
// Certificate is indistinguishable from one that does not exist.

// Certificate key types: every type Let's Encrypt accepts for a subscriber
// certificate. The Server names one on each request; the Agent generates it.
// ECDSA P-256 is the default, being the smallest and the fastest; RSA exists
// for the clients and configurations that need it, notably a dual-certificate
// server.
const (
	KeyTypeECP256  = "ec_p256"
	KeyTypeECP384  = "ec_p384"
	KeyTypeRSA2048 = "rsa_2048"
	KeyTypeRSA3072 = "rsa_3072"
	KeyTypeRSA4096 = "rsa_4096"
)

// KeyTypes lists every key type in the order a picker offers them. The list
// grows: an Agent asked for a key type it does not know acks the command
// unsupported.
var KeyTypes = []string{KeyTypeECP256, KeyTypeECP384, KeyTypeRSA2048, KeyTypeRSA3072, KeyTypeRSA4096}

// KeyTypeOf names a public key's type, or "" for a key no easySSL Certificate
// can use. The Agent labels its keystore with it and the Server checks each CSR
// against the requested type with it, so the mapping has one definition.
func KeyTypeOf(pub crypto.PublicKey) string {
	switch k := pub.(type) {
	case *ecdsa.PublicKey:
		switch k.Curve {
		case elliptic.P256():
			return KeyTypeECP256
		case elliptic.P384():
			return KeyTypeECP384
		}
	case *rsa.PublicKey:
		switch k.N.BitLen() {
		case 2048:
			return KeyTypeRSA2048
		case 3072:
			return KeyTypeRSA3072
		case 4096:
			return KeyTypeRSA4096
		}
	}
	return ""
}

func ValidKeyType(t string) bool {
	for _, k := range KeyTypes {
		if k == t {
			return true
		}
	}
	return false
}

// CSRWork is what the named Agent is asked to sign: the domain set fixed when
// the Certificate was requested, which the Agent signs exactly and never
// chooses. The CSR carries these as SANs only, no Subject Common Name, because
// an ACME client orders the Common Name too and an added name there would be
// one nobody requested.
//
// KeyFingerprint names the key to sign with, as the keystore names it: the
// lowercase hex SHA-256 of its SubjectPublicKeyInfo. Empty asks for a fresh
// key. It is set when the request reused an existing key, and once a CSR has
// been stored, so a re-run after a crash loads the key it already used. An
// Agent that does not hold the named key builds no CSR and never substitutes
// one: the Server records the fingerprint of every CSR it accepts and rejects a
// CSR from a different key.
type CSRWork struct {
	Domains        []string `json:"domains"`
	KeyType        string   `json:"keyType"`
	KeyFingerprint string   `json:"keyFingerprint,omitempty"`
}

// CSRSubmission carries the PEM CSR back: one CERTIFICATE REQUEST block,
// bounded small (a CSR is a few kilobytes). The Server verifies the SAN set,
// the absence of a CN, the key type and the self-signature before storing,
// since the CSR and not the Certificate record decides what gets issued.
type CSRSubmission struct {
	CSR string `json:"csr"`
}

// Error codes for the CSR submission. All three are terminal for the Agent:
// none can be fixed by retrying the same submission, so the Agent records the
// refusal and acks the command.
const (
	// CodeCSRAlreadySubmitted (409): a CSR is already stored. Treated as success,
	// because an Agent re-run after a crash cannot recognize its own submission:
	// an ECDSA re-signature never reproduces the same bytes, and an RSA one
	// depends on the key being the same. The stored CSR is from the key the re-run
	// was told to use, since CSRWork names the stored CSR's key once one exists.
	CodeCSRAlreadySubmitted = "csr_already_submitted"
	// CodeNotAwaitingCSR (409): the certificate is failed or issued; the
	// instruction is stale and is discarded.
	CodeNotAwaitingCSR = "not_awaiting_csr"
	// CodeCSRMismatch (422): verification failed and nothing was stored. The
	// Server fails the Certificate at the rejection, naming this Agent and what
	// failed, because the Agent acks a rejection as handled and an acked command
	// never expires. A retry issues a fresh command.
	CodeCSRMismatch = "csr_mismatch"
)

// CodeCertificateNotFound (404): no Certificate with that id is waiting on this
// Agent, whether it never existed, was deleted, or names another Agent. The
// Agent discards the command: no retry changes the answer. Any other 404, such
// as a proxy's or an older Server's for a route it lacks, carries no such code
// and is retried.
const CodeCertificateNotFound = "certificate_not_found"
