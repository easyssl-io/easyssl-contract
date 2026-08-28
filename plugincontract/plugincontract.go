// Package plugincontract is the contract between the easySSL Agent and every
// Plugin executable. It is versioned independently of the Agent↔Server wire
// protocol, because Plugins and Agents release independently.
//
// Every operation is one execution of the Plugin binary: the operation name is
// the single argv argument, the JSON request rides stdin, the JSON response
// rides stdout, and stderr carries the Plugin's own diagnostics, which the
// Agent logs but never parses. The operation name is the only thing argv
// carries, deliberately: `/proc/<pid>/cmdline` is world-readable on Linux, so
// anything sensitive — the private key above all — must ride the pipe, where
// nothing remains to read afterwards.
package plugincontract

// ContractVersion is this contract's version, MAJOR.MINOR.BUILD. Additive
// fields bump MINOR, and a Plugin seeing a newer MINOR than it knows ignores
// the fields it does not recognize.
const ContractVersion = "0.0.1"

// Operation names, passed as the Plugin's single argv argument. A Plugin
// receiving an operation it does not implement exits non-zero with a line on
// stderr, and the Agent treats the operation as unsupported.
const (
	// OpDescribe asks the Plugin to identify itself and declare its settings.
	// It takes no request body. The Agent runs it once at install time and
	// reports the answer to the Server, which is how the web console can show
	// and configure a Plugin's settings without ever reaching the host.
	OpDescribe = "describe"
	// OpInstall performs one certificate installation. Everything that happens
	// on the host happens inside this one call; the Agent sees only the
	// response.
	OpInstall = "install"
)

// Setting types. The type tells a consumer how to present and check a value; a
// command is a shell script that may span lines and is run whole via
// `/bin/sh -c`.
const (
	SettingTypeCommand = "command"
	SettingTypePath    = "path"
	SettingTypeText    = "text"
	SettingTypeEnum    = "enum"
	SettingTypeBool    = "boolean"
)

// SettingDeclaration is one setting a Plugin declares about itself. The Key is
// identity — values are stored against it, so it must never be renamed between
// versions — while Label and Description may change freely, nothing being
// stored against them. Required with an empty Default means the Plugin cannot
// know the value and someone must supply one before a deployment can run.
//
// Description, here and on DescribeResponse, is written in a small Markdown
// subset, which is all the web console renders:
//
//   - `backticks` set a span as code. Use them for anything a reader would
//     type, copy or look for on a host: a command, a path, an environment
//     variable, a flag.
//   - **double asterisks** set a span bold. A code span may sit inside a bold
//     span, and keeps its backticks' meaning there; bold does not nest.
//   - a blank line separates paragraphs; a single line break is kept as one.
//
// Nothing else is interpreted. HTML in particular is shown as the characters
// typed, never rendered, because a Plugin's text is somebody else's text once
// Plugins are published and the web console must not carry its markup. Markdown
// rather than HTML for the same reason: a subset this small has nothing to
// sanitize.
type SettingDeclaration struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Default     string `json:"default"`
	Required    bool   `json:"required"`
	// Summary marks the one setting whose value an unlabeled installation
	// point of this Plugin is known by on screen. At most one per Plugin;
	// false everywhere means none.
	Summary bool `json:"summary"`
}

// DescribeResponse is OpDescribe's answer: who this build is and what it
// declares. Version is the Plugin's own; the Agent reports it together with
// the artifact digest, so what a host runs is always attributable.
type DescribeResponse struct {
	ContractVersion string `json:"contractVersion"`
	Name            string `json:"name"`
	Version         string `json:"version"`
	// Description says what this Plugin installs onto, in one or two sentences
	// a person choosing between Plugins can read. The settings each describe
	// one value; nothing until now described the Plugin, so a page about one
	// could only list what it could be configured with. Optional: a Plugin that
	// says nothing is shown without it rather than with an empty line. Same
	// Markdown subset as SettingDeclaration.Description.
	Description string               `json:"description,omitempty"`
	Settings    []SettingDeclaration `json:"settings"`
}

// InstallRequest is OpInstall's request. The material arrives as PEM text, not
// as paths: a Plugin that installs through an API needs no file, and one that
// does need files creates them itself, inside WorkDir.
type InstallRequest struct {
	ContractVersion string `json:"contractVersion"`
	CertificatePem  string `json:"certificatePem"`
	ChainPem        string `json:"chainPem"`
	PrivateKeyPem   string `json:"privateKeyPem"`
	// WorkDir is a directory the Agent created for this run and removes after
	// it, however it ends. A Plugin that materializes files puts them there, so
	// a run the Agent had to kill cannot leave key material behind; a Plugin
	// that writes nothing ignores it.
	WorkDir string `json:"workDir"`
	// Settings is the resolved value of every setting this Plugin declares,
	// decided Server-side; the Plugin trusts it and resolves nothing itself.
	Settings map[string]string `json:"settings"`
}

// InstallResponse is OpInstall's answer. OK is the deployment's outcome;
// Output is one text, whatever the Plugin chooses to say about what it did —
// exit statuses of inner commands live inside it, never as fields, because the
// commands ran inside the Plugin and only this answer crosses back.
//
// The process exit code is transport only: exit 0 with OK false is a
// deployment that failed, while a non-zero exit or unparseable stdout is a
// Plugin failure the Agent reports as the deployment's failure with whatever
// it captured.
type InstallResponse struct {
	ContractVersion string `json:"contractVersion"`
	OK              bool   `json:"ok"`
	Output          string `json:"output"`
}
