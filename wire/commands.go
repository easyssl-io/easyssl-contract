package wire

// The Server→Agent push channel. The Server cannot open a connection to a
// Agent, so the Agent holds one long-lived stream open and the Server writes
// to it. What travels over that stream is deliberately almost nothing: a
// command id. The Agent then reads the command, and the resource it refers to,
// over ordinary request/response routes.
//
// The stream is an optimization, never the delivery mechanism. A Agent that
// never manages to hold it open still performs every command, just later, by
// listing pending commands on its regular poll. That is what makes a lost,
// duplicated or out-of-order event harmless, and it is why nothing here
// promises delivery.

import "time"

// Push-channel route paths. All three carry the per-Agent bearer token and all
// three require an active Agent: a Agent still awaiting authorization, or one
// whose registration was removed, is refused exactly as it is on check-in.
const (
	// RouteEvents is the long-lived stream (SSE). One connection per Agent
	// carries every command type, rather than one connection per capability,
	// so a Agent behind a restrictive network has a single thing to keep open.
	RouteEvents = "/api/v1/agent/events"
	// RouteCommands lists the commands a Agent has not finished yet. The
	// Agent reads this on every (re)connect and on every poll, which is what
	// actually guarantees a command runs.
	RouteCommands = "/api/v1/agent/commands"
	// RouteCommandAck is RouteCommands + "/{id}/ack": the Agent reporting that
	// it has finished with a command, one way or another.
	RouteCommandAck = "/api/v1/agent/commands/{id}/ack"
)

// Command types. A Agent that does not recognize a type reports it
// unsupported rather than failing, so adding one here stays backward
// compatible with Agents already in the field.
const (
	// CommandRequestCSR asks the Agent to generate a key and return a CSR for
	// a certificate the Server is issuing. RefID is the certificate.
	//
	// The names are already decided when this is sent: the Agent reads the
	// certificate's domains and signs those, rather than choosing what to ask
	// for. The Server checks the returned CSR against them before using it.
	CommandRequestCSR = "request_csr"
	// CommandDeploy asks the Agent to place a certificate at a Target and run
	// its reload command. RefID is the deployment.
	CommandDeploy = "deploy"
)

// Terminal states a Agent can report for a command. There is no "failed":
// whether the work succeeded belongs to the resource the command refers to, so
// a deployment that fails is recorded on the deployment. These say only whether
// the Agent is done with the command.
const (
	// CommandStatusAcked: the Agent's attempt terminated. It may have failed;
	// that outcome lives on the referenced resource.
	CommandStatusAcked = "acked"
	// CommandStatusUnsupported: this Agent does not know this command type,
	// which usually means it is older than the Server.
	CommandStatusUnsupported = "unsupported"
)

// CodeCommandExpired is returned when a Agent acks a command whose expiry has
// passed (409). The Agent must not have acted on it: the instruction is stale
// and is discarded rather than run late.
const CodeCommandExpired = "command_expired"

// EventCommand is the SSE event name carrying a command id. It is the only
// event type on the stream today; an unknown event name is ignored, the same
// way an unknown command type is.
const EventCommand = "command"

// CommandEvent is the body of an EventCommand. It carries an id and nothing
// else on purpose: putting the work in the event would create a second copy of
// it that could disagree with the stored one, and would mean an older Agent
// receiving a payload it cannot interpret.
type CommandEvent struct {
	CommandID string `json:"commandId"`
}

// Command is one instruction for a Agent, as returned by RouteCommands. RefID
// names the resource holding the actual work; the Agent fetches that resource
// by its own route. Nothing here describes the work itself.
type Command struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	RefID     string    `json:"refId"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// CommandsResponse is the reconcile fetch, oldest command first so a Agent
// works through a backlog in the order the Server created it.
type CommandsResponse struct {
	Commands []Command `json:"commands"`
}

// AckRequest reports that the Agent has finished with a command. Status is
// CommandStatusAcked or CommandStatusUnsupported.
//
// It is sent after the attempt terminates and after any outcome has been
// recorded on the referenced resource, never before. A Agent that dies in
// between leaves the command outstanding, so it is handed back on the next
// connect and re-run against an outcome already recorded, which is why the
// Agent must be able to receive the same command twice and act once.
type AckRequest struct {
	Status string `json:"status"`
}
