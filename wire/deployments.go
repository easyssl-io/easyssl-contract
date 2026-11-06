package wire

import "github.com/easyssl-io/easyssl-contract/l10n"

// The Deployment resource behind a CommandDeploy: how the Agent learns what to
// install and reports how it went. The command carries only the Deployment id.
//
// The payload never carries a private key, in either direction. The key was
// generated on the Agent and lives in its keystore; the Agent reads it there by
// the Certificate id this resource names and hands it to the Plugin on the
// host. A Deployment can therefore only run on the Agent that holds the key.

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

// CodeDeploymentNotFound (404): no Deployment with that id is this Agent's. The
// Agent discards the command, as for CodeCertificateNotFound.
const CodeDeploymentNotFound = "deployment_not_found"

// DeploymentWork is what RouteAgentDeployment answers while the Deployment is
// pending. KeyFingerprint names the Certificate's key the way the Agent's
// keystore names it (the lowercase hex SHA-256 of its SubjectPublicKeyInfo);
// the Agent reads that key from its own keystore and fails the Deployment if it
// is not there. Settings is the Server-resolved value of every setting the
// owning Plugin declares, snapshotted and recorded when the Deployment was
// queued, so what ran never depends on a later edit.
type DeploymentWork struct {
	ID             string            `json:"id"`
	State          string            `json:"state"`
	CertificateID  string            `json:"certificateId"`
	KeyFingerprint string            `json:"keyFingerprint,omitempty"`
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
// Everything under PluginRun is the Plugin's and AgentMessage is the Agent's.
// Nothing merges them, so no field is ever part prose and part captured bytes.
type DeploymentResult struct {
	OK            bool   `json:"ok"`
	PluginVersion string `json:"pluginVersion"`
	// AgentMessage is what the Agent has to say that the Plugin could not: it was
	// not installed, its artifact could not be run, this host holds no key for the
	// Certificate, the run was killed at the bound. One sentence, never captured
	// output. Absent on a clean run, so its presence is the signal that something
	// happened outside the Plugin. A text per language, English required, because
	// it is stored and read later by people who do not share the host's language.
	AgentMessage l10n.Text `json:"agentMessage,omitempty"`
	// PluginRun is what the process did, and is absent when no process ran.
	// Its presence is the fact that one did.
	PluginRun *PluginRun `json:"pluginRun,omitempty"`
}

// PluginRun is one execution of a Plugin as facts: what it exited with, how
// long it took, and what it wrote. Nothing here is interpreted, and nothing is
// the Agent's own words.
//
// The request must stay absent: it carries the Certificate's private key
// (plugincontract.DeployRequest.PrivateKeyPem), which never crosses to the
// Server.
//
// Output is redacted of private keys and bounded by the Agent, because a Plugin
// that echoes its request or dumps its environment on failure is how a key
// reaches its own output.
type PluginRun struct {
	ExitStatus int   `json:"exitStatus"`
	DurationMs int64 `json:"durationMs"`
	// Output is both streams in the order the Agent received them, each chunk
	// saying which one it came from, because two strings cannot say whether a
	// stderr message came before or after a line of stdout. Consecutive lines from
	// the same stream are one chunk, so a Plugin that writes only to stdout
	// produces exactly one.
	Output []OutputChunk `json:"output,omitempty"`
	// StdoutBytes and StderrBytes are what the Plugin produced, which is not
	// always what Output carries: a stream past the Agent's bound is cut.
	// Counts rather than a `truncated` flag so there is no boolean that can
	// disagree with the text beside it.
	StdoutBytes int `json:"stdoutBytes"`
	StderrBytes int `json:"stderrBytes"`
}

// Stream names which of a Plugin's two outputs a chunk came from.
const (
	StreamStdout = "stdout"
	StreamStderr = "stderr"
)

// OutputChunk is a run of consecutive lines from one stream.
type OutputChunk struct {
	Stream string `json:"stream"`
	Text   string `json:"text"`
}
