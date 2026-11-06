package wire

// The Server→Agent push channel. The Server cannot open a connection to an
// Agent, so the Agent holds one long-lived stream open and the Server writes to
// it. The stream carries no work: the Agent reads each command, and the
// resource it refers to, over ordinary request/response routes.
//
// The stream is an optimization, never the delivery mechanism. An Agent that
// never holds it open still performs every command, later, by listing pending
// commands on its regular poll, which is what makes a lost, duplicated or
// out-of-order event harmless.

import "time"

// Command types. An Agent that does not recognize a type acks it
// CommandStatusUnsupported rather than failing, so adding one here stays
// backward compatible with Agents already in the field.
const (
	// CommandRequestCSR asks the Agent to generate a key and return a CSR for a
	// Certificate the Server is issuing. RefID is the Certificate. The Agent signs
	// the Certificate's domains and never chooses them, and the Server checks the
	// returned CSR against them before using it.
	CommandRequestCSR = "request_csr"
	// CommandDeploy asks the Agent to place a Certificate at a Target and run
	// its reload command. RefID is the Deployment.
	CommandDeploy = "deploy"
)

// Terminal states an Agent can report for a command. There is no "failed":
// whether the work succeeded belongs to the resource the command refers to, so
// a deployment that fails is recorded on the deployment. These say only whether
// the Agent is done with the command. A Server that does not know a status a
// newer Agent sends reads it as CommandStatusAcked: the Agent is done either
// way, and refusing the ack would hand it the command again.
const (
	// CommandStatusAcked: the Agent's attempt terminated. It may have failed;
	// that outcome lives on the referenced resource.
	CommandStatusAcked = "acked"
	// CommandStatusUnsupported: this Agent cannot do what the command asks,
	// which usually means it is older than the Server: it does not know the
	// command type, or the work names something it does not know, such as a key
	// type. The Server ends the referenced work, naming this Agent.
	CommandStatusUnsupported = "unsupported"
)

// A command the Agent could not finish for a reason that may clear (the Server
// unreachable or failing, the host's keystore refusing) is left unacked: it is
// handed back on the next fetch and expires if the reason never clears, which
// fails the referenced work naming this Agent. Acking it would end it with
// nothing recorded.

// CodeCommandExpired is returned when an Agent acks a command whose expiry has
// passed (409). The Agent must not have acted on it: the instruction is stale
// and is discarded rather than run late.
const CodeCommandExpired = "command_expired"

// More error codes of the push channel.
const (
	// CodeUnknownCommand: the acknowledged command is not one this Agent was
	// given (404).
	CodeUnknownCommand = "unknown_command"
	// CodeStreamingUnsupported: the connection cannot carry an event stream,
	// usually a proxy in between; the Agent falls back to polling (500).
	CodeStreamingUnsupported = "streaming_unsupported"
)

// EventCommand is the SSE event name that tells an Agent something is waiting
// for it. It is the only event type on the stream; an unknown event name is
// ignored, the same way an unknown command type is.
//
// It carries no body: `data: {}` is sent only because the SSE grammar
// dispatches an event when a data field is present. The Agent answers by
// reading its outstanding commands over REST, the only path that also delivers
// what it missed while disconnected. An id in the event would invite acting on
// that one command and skipping the reconcile, which silently loses work the
// first time an event goes missing.
const EventCommand = "command"

// Command is one instruction for an Agent, as returned by RouteAgentCommands. RefID
// names the resource holding the actual work; the Agent fetches that resource
// by its own route. Nothing here describes the work itself.
type Command struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	RefID     string    `json:"refId"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// CommandsResponse is the reconcile fetch, oldest command first so an Agent
// works through a backlog in the order the Server created it.
type CommandsResponse struct {
	Commands []Command `json:"commands"`
}

// AckRequest reports that the Agent has finished with a command. Status is
// CommandStatusAcked or CommandStatusUnsupported.
//
// It is sent after the attempt terminates and after any outcome has been
// recorded on the referenced resource, never before. An Agent that dies in
// between is handed the command again on the next connect, so it must be able
// to receive the same command twice and act once.
type AckRequest struct {
	Status string `json:"status"`
}
