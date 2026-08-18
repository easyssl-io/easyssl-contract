package wire

// The deployment resource behind a CommandDeploy: how the Agent learns what
// to install and reports how it went. The command carries only the deployment
// id; these routes are where the work lives, the same split the request_csr
// resource uses.
//
// The payload never carries a private key, in either direction. The key was
// generated on the Agent and lives in its keystore; the Agent reads it there
// by the certificate id this resource names and hands it to the Plugin on the
// host. A deployment can therefore only ever run on the Agent that holds the
// key.

// Deployment route paths. Both carry the per-Agent bearer token and answer
// only the Agent the deployment names; to any other Agent the deployment is
// indistinguishable from one that does not exist.
const (
	// RouteAgentDeployment is where the Agent reads a deployment's work:
	// the material, and the resolved value of every setting the Plugin will be
	// called with. A deployment already finished answers with its state and no
	// material, so a re-delivered command is acknowledged without re-running.
	RouteAgentDeployment = "/api/v1/agent/deployments/{id}"
	// RouteAgentDeploymentResult is RouteAgentDeployment + "/result": the
	// Agent's terminal report, write-once.
	RouteAgentDeploymentResult = "/api/v1/agent/deployments/{id}/result"
)

// Deployment states. A deployment records one push and is never rewritten
// after it ends.
const (
	DeploymentStatePending   = "pending"
	DeploymentStateSucceeded = "succeeded"
	DeploymentStateFailed    = "failed"
)

// CodeDeploymentFinished is returned when a result arrives for a deployment
// that is already terminal (409). The record is write-once, so the Agent
// treats this as done: an Agent that died between acting and reporting re-runs
// safely and its second report lands here.
const CodeDeploymentFinished = "deployment_finished"

// DeploymentWork is what RouteAgentDeployment answers while the deployment is
// pending. CertificateID names the certificate whose key the Agent reads from
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

// DeploymentResult is the Agent's terminal report: whether the install
// succeeded, which Plugin build actually ran, and the one output text the
// Plugin returned (or the Agent's own line when the Plugin could not answer).
// Sent after the outcome is final and before the command is acknowledged,
// never the other way around.
type DeploymentResult struct {
	OK            bool   `json:"ok"`
	PluginVersion string `json:"pluginVersion"`
	// Output is what the Plugin itself printed, and nothing else: empty
	// whenever no Plugin ran, so a reader is never shown the Agent's words
	// under the Plugin's name.
	Output string `json:"output"`
	// Message is the Agent's own account of the attempt, for what the Plugin
	// could not say: it was not installed, its artifact could not be run, this
	// host holds no key for the Certificate, the run was killed at the bound.
	Message string `json:"message,omitempty"`
}
