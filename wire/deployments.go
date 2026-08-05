package wire

// The deployment resource behind a CommandDeploy: how the Client learns what
// to install and reports how it went. The command carries only the deployment
// id; these routes are where the work lives, the same split the request_csr
// resource uses.
//
// The payload never carries a private key, in either direction. The key was
// generated on the Client and lives in its keystore; the Client reads it there
// by the certificate id this resource names and hands it to the Plugin on the
// host. A deployment can therefore only ever run on the Client that holds the
// key.

// Deployment route paths. Both carry the per-Client bearer token and answer
// only the Client the deployment names; to any other Client the deployment is
// indistinguishable from one that does not exist.
const (
	// RouteClientDeployment is where the Client reads a deployment's work:
	// the material, and the resolved value of every setting the Plugin will be
	// called with. A deployment already finished answers with its state and no
	// material, so a re-delivered command is acknowledged without re-running.
	RouteClientDeployment = "/api/v1/client/deployments/{id}"
	// RouteClientDeploymentResult is RouteClientDeployment + "/result": the
	// Client's terminal report, write-once.
	RouteClientDeploymentResult = "/api/v1/client/deployments/{id}/result"
)

// Deployment states. A deployment records one push and is never rewritten
// after it ends.
const (
	DeploymentStatePending   = "pending"
	DeploymentStateSucceeded = "succeeded"
	DeploymentStateFailed    = "failed"
)

// CodeDeploymentFinished is returned when a result arrives for a deployment
// that is already terminal (409). The record is write-once, so the Client
// treats this as done: a Client that died between acting and reporting re-runs
// safely and its second report lands here.
const CodeDeploymentFinished = "deployment_finished"

// DeploymentWork is what RouteClientDeployment answers while the deployment is
// pending. CertificateID names the certificate whose key the Client reads from
// its own keystore; Settings is the Server-resolved value of every setting the
// owning Plugin declares, snapshotted when the deployment was queued, which is
// also what gets recorded — so what ran never depends on an edit made after
// the user confirmed.
type DeploymentWork struct {
	ID             string            `json:"id"`
	State          string            `json:"state"`
	CertificateID  string            `json:"certificateId"`
	PluginID       string            `json:"pluginId"`
	CertificatePem string            `json:"certificatePem,omitempty"`
	ChainPem       string            `json:"chainPem,omitempty"`
	Settings       map[string]string `json:"settings,omitempty"`
}

// DeploymentResult is the Client's terminal report: whether the install
// succeeded, which Plugin build actually ran, and the one output text the
// Plugin returned (or the Client's own line when the Plugin could not answer).
// Sent after the outcome is final and before the command is acknowledged,
// never the other way around.
type DeploymentResult struct {
	OK            bool   `json:"ok"`
	PluginVersion string `json:"pluginVersion"`
	Output        string `json:"output"`
}
