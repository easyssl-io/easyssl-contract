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
// succeeded, which Plugin build ran, and what the run produced. Sent after the
// outcome is final and before the command is acknowledged, never the other way
// around.
//
// Two voices, and which is which is carried by the field names rather than by a
// convention a reader has to remember. Everything under PluginRun is the
// Plugin's; AgentMessage is the Agent's. Nothing merges them, so no field is
// ever part prose and part captured bytes.
type DeploymentResult struct {
	OK            bool   `json:"ok"`
	PluginVersion string `json:"pluginVersion"`
	// AgentMessage is what the Agent has to say that the Plugin could not: it
	// was not installed, its artifact could not be run, this host holds no key
	// for the Certificate, the run was killed at the bound. One sentence, and
	// never a container for captured output — the streams have their own
	// fields below. Absent on a clean run, so its presence is itself the
	// signal that something happened outside the Plugin.
	AgentMessage string `json:"agentMessage,omitempty"`
	// PluginRun is what the process did, and is absent when no process ran.
	// Its presence is the fact that one did.
	PluginRun *PluginRun `json:"pluginRun,omitempty"`
}

// PluginRun is one execution of a Plugin, as facts rather than as prose: what
// it exited with, how long it took, and what it wrote. Nothing here is
// interpreted, and nothing is the Agent's own words.
//
// The request is deliberately absent and must stay absent. It carries the
// Certificate's private key (plugincontract.InstallRequest.PrivateKeyPem),
// which never crosses to the Server in either direction; a stdin field added
// here for symmetry would undo the one property the whole key design rests on.
//
// Both streams are bounded by the Agent and are reported after its private-key
// redaction, because a Plugin that echoes its request or dumps its environment
// on failure is how a key reaches its own output.
type PluginRun struct {
	ExitStatus int   `json:"exitStatus"`
	DurationMs int64 `json:"durationMs"`
	// Stdout is the Plugin's answer: normally one JSON document, which is why
	// nothing here interprets it. Omitted when the Plugin printed nothing.
	Stdout string `json:"stdout,omitempty"`
	// Stderr is the Plugin's own diagnostics, which is where a broken Plugin
	// explains itself. Omitted when it wrote none.
	Stderr string `json:"stderr,omitempty"`
	// StdoutBytes and StderrBytes are what the Plugin produced, which is not
	// always what the fields above carry: a stream past the Agent's bound is
	// cut. Counts rather than a `truncated` flag so there is no boolean that
	// can disagree with the text beside it — a string shorter than its count
	// is a stream that was clipped.
	StdoutBytes int `json:"stdoutBytes"`
	StderrBytes int `json:"stderrBytes"`
}
