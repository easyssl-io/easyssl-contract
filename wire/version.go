package wire

// ProtocolVersion is the easySSL Agent-to-Server wire-protocol version,
// formatted MAJOR.MINOR.BUILD. MAJOR is bumped only for breaking contract
// changes, MINOR for additive backward-compatible features, BUILD for
// backward-compatible fixes.
//
// 3.0.0 drops the Plugin's response document and reports the run as the
// process gave it. A Plugin now answers OpInstall with an exit status and its
// two streams, so DeploymentResult.OK is exitStatus == 0 rather than a field a
// Plugin set, and PluginRun carries Output — both streams interleaved in the
// order they arrived, each chunk naming its stream — in place of two separate
// strings that could not say which came first. Breaking twice over, so MAJOR.
//
// 2.0.0 replaces DeploymentResult's Output and Message with AgentMessage and a
// PluginRun holding the exit status, the duration and both streams with their
// produced sizes. Breaking, so MAJOR. Output and Message could not say which
// voice a reader was hearing: Message was the Agent's sentence and also the
// container the streams were formatted into, differently for each way a run
// could fail, so the exit status was recoverable only by parsing English and
// nothing could be rendered as what it was. Nothing is published, so no Agent
// or Server outside this repository speaks the old shape.
//
// 1.0.0 empties the command event: it carried an id no Agent ever read, and an
// Agent that acted on that id instead of reconciling would lose work whenever
// an event went missing. Breaking, so MAJOR: a Server on 1.0.0 sends `data: {}`
// where a 0.x Agent expects an id, and that Agent skips the event entirely,
// falling back to its check-in. Nothing is published yet, so no such Agent
// exists outside this repository.
//
// 0.2.0 adds PendingCommands to CheckinResponse, letting an Agent skip its
// reconcile fetch when the Server has nothing outstanding for it. Additive, so
// MINOR again: the field is a pointer and its absence means "this Server does
// not report it", which an Agent reads as "fetch anyway".
//
// 0.1.0 added ProtocolVersion to CheckinRequest, which is additive and so bumps
// MINOR: a Server that knows the field reads it from Agents that send it and
// keeps what it stored for Agents that do not, and an Agent that does not know
// it is unaffected. This is the first bump off the unpublished 0.0.1 baseline,
// and it is what makes the value worth reporting on a check-in at all: until
// there were two of them, a stale record and a current one could not differ.
const ProtocolVersion = "3.0.0"
