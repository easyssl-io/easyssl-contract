package wire

// What the installer the Server serves and the Agent it installs must agree on
// about the host, kept here so neither spells them on its own.
const (
	// AgentBinaryPath is where the installer places the Agent binary and its
	// service runs it from.
	AgentBinaryPath = "/usr/local/bin/easyssl-agent"
	// The environment variables `easyssl-agent register` reads, which the
	// installer passes on and the web console's install command sets. Each has
	// the flag of the same name, lower case and without the prefix
	// (`--registration-key`).
	EnvRegistrationKey   = "EASYSSL_REGISTRATION_KEY"
	EnvReactivationToken = "EASYSSL_REACTIVATION_TOKEN"
	EnvAgentName         = "EASYSSL_AGENT_NAME"
)
