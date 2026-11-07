// Package plugincontract is the contract between the easySSL Agent and every
// Plugin executable. It shares the module's one version with `wire`, which
// says which Plugins an Agent runs and which Agents a Server answers.
//
// Every operation is one execution of the Plugin binary: the operation name is
// the single argv argument, the JSON request rides stdin, the JSON response
// rides stdout, and stderr carries the Plugin's own diagnostics, which the
// Agent logs but never parses. Nothing else rides argv because
// `/proc/<pid>/cmdline` is world-readable on Linux, so anything sensitive, the
// private key above all, must ride the pipe.
//
// A Plugin runs as the Agent's user, root, with the Agent's environment, and
// whatever it starts (a customer's command, say) inherits both. Environment
// variables named `EASYSSL_*` are reserved for this contract and the
// Agent: a Plugin sets only the ones it documents for the commands it runs.
// What a Plugin starts lives in the Agent's process group and service, so a
// process left running ends when the Agent restarts or is updated; a Plugin
// that starts services hands them to the host's own service manager.
//
// An exit status is zero or not: a Plugin does not give other meanings to
// particular non-zero statuses, so the Agent never reads one.
package plugincontract

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/easyssl-io/easyssl-contract/l10n"
)

// ContractVersion is the version of this module, the one release tag `wire`,
// `plugincontract` and `l10n` share; README.md says what moves it. It stays at
// 0.0.1 until the first release.
const ContractVersion = "0.0.1"

// PreviousBreakingVersion is the breaking version before ContractVersion's,
// which a build still speaks to: a Server still answers an Agent built against
// it, and an Agent still runs a Plugin built against it. Empty until a release
// moves the breaking version for the first time; that release sets it.
const PreviousBreakingVersion = ""

// BreakingVersion is the part of a semantic version that a breaking change
// moves: `0.<minor>` while the major number is 0, the major number from 1.0.0
// on. ok is false for anything that is not MAJOR.MINOR.PATCH in digits.
func BreakingVersion(v string) (breaking string, ok bool) {
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return "", false
	}
	for _, p := range parts {
		if p == "" || strings.Trim(p, "0123456789") != "" || (len(p) > 1 && p[0] == '0') {
			return "", false
		}
	}
	if parts[0] == "0" {
		return "0." + parts[1], true
	}
	return parts[0], true
}

// Compatible reports whether a build of this contract speaks to a peer built
// against version v: the same breaking version as ContractVersion, or
// PreviousBreakingVersion. A Server refuses any other Agent with
// `contract_unsupported`, and an Agent refuses any other Plugin, so a mismatch
// is named rather than failing on a field one side does not know.
func Compatible(v string) bool {
	got, ok := BreakingVersion(v)
	if !ok {
		return false
	}
	own, _ := BreakingVersion(ContractVersion)
	return got == own || (PreviousBreakingVersion != "" && got == PreviousBreakingVersion)
}

// AnswerTimeout bounds each describe and discover execution. The Agent runs
// both before every check-in, each Plugin at once, and kills a Plugin's whole
// process group when it is late, reporting that Plugin unobserved for the run.
const AnswerTimeout = 10 * time.Second

// Operation names, passed as the Plugin's single argv argument. Every Plugin
// implements all three. An operation added later comes with a new contract
// version, and an Agent calls it only on a Plugin whose describe answer reports
// that version or a later one, so no Plugin is ever asked for an operation it
// does not know. A Plugin's own internal modes never take an operation name:
// they are flags, so this list stays the contract's alone.
const (
	// OpDescribe asks the Plugin what it is: its identity, its description and
	// the settings it declares. It takes no request body and reads nothing from
	// the host, so it needs no privilege and answers the same on any machine,
	// which is what lets a build publish it beside the artifact. Every Plugin
	// implements it. The Agent runs it on every check-in, right before discover,
	// so a binary swapped by hand is read from what would run.
	OpDescribe = "describe"
	// OpDiscover asks the Plugin what it sees on the host: the installation
	// points it found as discovered Targets, with warnings for what it could not
	// turn into one. It takes no request body. Every Plugin implements it; one
	// that inspects nothing answers with both lists empty.
	//
	// A configuration the Plugin cannot read is a warning, not a failure of the
	// operation. A non-zero exit or a timeout is the operation producing nothing
	// this run, which the Agent reports as the Plugin being present but
	// unobserved.
	OpDiscover = "discover"
	// OpDeploy carries out one Deployment: it puts a Certificate onto whatever the
	// Plugin installs onto, and reloads whatever has to be reloaded. Everything
	// that happens on the host happens inside this one call. It is named for the
	// Deployment record and not "install", which already names installing the
	// Agent, a Plugin and the systemd unit.
	OpDeploy = "deploy"
)

// Setting types. The type tells a consumer how to present and check a value; a
// command is a shell script that may span lines and is run whole via
// `/bin/sh -c`. A command runs as root and a path is written as root, so
// whoever can set either, the Server and its Organization's administrators,
// holds root on the host, by design. A reader that does not know a type shows and edits the value as
// text, and never refuses the declaration: a type is added only with a new
// contract version, and an older reader still works with the value.
const (
	SettingTypeCommand = "command"
	SettingTypePath    = "path"
	SettingTypeText    = "text"
)

// SettingDeclaration is one setting a Plugin declares about itself. Values are
// stored against Key, so it must never be renamed between versions; Label and
// Description may change freely. Required with an empty Default means someone
// must supply a value before a Deployment can run.
//
// A later build of a Plugin MUST accept every setting an earlier build
// declared, because Targets configured under the earlier build keep their
// values, and a Deployment queued before an upgrade runs on the new build with
// the settings resolved before it:
//
//   - a Key keeps its meaning and is never renamed or reused;
//   - a setting a later build adds is optional, or has a Default that
//     reproduces the earlier behavior, and deploy treats it as absent when a
//     request does not carry it;
//   - a setting a later build drops is ignored when a request still carries it;
//   - whether a setting is Identity, and an Identity setting's Default, never
//     change, since either would give every discovered Target of the Plugin a
//     new discovery key, and the Targets made from them would no longer match.
//
// A build that cannot honor this is a new Plugin, under a new name. Adding a
// Required setting with no Default is allowed but makes every existing Target
// of the Plugin undeployable until someone sets it.
//
// Label and Description, here and on DescribeResponse, are text per language
// (l10n.Text), English always among them. A Plugin cannot be asked for one
// language, because what it answers is stored and read later by people who do
// not share one.
//
// Description is written, in each language, in a small Markdown subset, which
// is all the web console renders:
//
//   - `backticks` set a span as code: anything a reader would type, copy or
//     look for on a host.
//   - **double asterisks** set a span bold. A code span may sit inside a bold
//     span; bold does not nest.
//   - a blank line separates paragraphs; a single line break is kept as one.
//
// Nothing else is interpreted, and HTML is shown as the characters typed, never
// rendered: a subset this small has nothing to sanitize. There are no escapes:
// a code span cannot contain a backtick, so text that holds one, a host path
// or a program's output say, is written outside a code span (Code does this).
type SettingDeclaration struct {
	Key         string    `json:"key"`
	Label       l10n.Text `json:"label"`
	Description l10n.Text `json:"description"`
	Type        string    `json:"type"`
	Default     string    `json:"default"`
	Required    bool      `json:"required"`
	// Summary marks the one setting whose value an unlabeled installation point of
	// this Plugin is known by on screen. At most one per Plugin, which
	// ValidateSettings enforces; false everywhere means none.
	Summary bool `json:"summary"`
	// Identity marks a setting whose value identifies an installation point: the
	// Server composes a discovered Target's discovery key from the values of every
	// Identity setting, ordered by Key so that reordering the declarations changes
	// nothing, and matches discovered Targets to Targets on it. Distinct from Summary, which is what a reader sees a Target
	// called. A Plugin that reports discovered Targets flags at least one
	// (ValidateDiscover).
	Identity bool `json:"identity"`
}

// DescribeResponse is OpDescribe's answer: what this build is and what it
// declares, the same wherever it runs. It carries no id: a Plugin's id is the
// name of its executable, which is how the Agent runs it and the Marketplace
// offers it, so a second copy could only disagree.
type DescribeResponse struct {
	ContractVersion string `json:"contractVersion"`
	Version         string `json:"version"`
	// Description is the build's own account of what it installs onto, in the
	// Markdown subset the web console renders, per language; optional.
	Description  l10n.Text            `json:"description,omitempty"`
	Declarations []SettingDeclaration `json:"declarations"`
}

// DiscoverResponse is OpDiscover's answer: what this build found on the host.
type DiscoverResponse struct {
	ContractVersion string `json:"contractVersion"`
	// DiscoveredTargets are the installation points the Plugin found on the
	// host, reported as observations: what the Server makes of each, an offer
	// to add a Target or the host's account of a Target that exists, is the
	// Server's decision. Empty for a Plugin that inspects nothing.
	DiscoveredTargets []DiscoveredTarget `json:"discoveredTargets"`
	// Warnings are what discovery saw and could not report as a discovered
	// Target, each per language and in the Markdown subset, shown on the Agent
	// page under the Plugin.
	Warnings []l10n.Text `json:"warnings"`
}

// DiscoveredTarget is one installation point a Plugin found. It carries no
// key of its own: the Server composes one from the Identity settings.
type DiscoveredTarget struct {
	// Settings is a value per declared setting key, opaque to the Server and
	// resolved by the Plugin: a path arrives as the file it names, never as
	// the spelling in the configuration. Promotion copies these into the
	// Target.
	Settings map[string]string `json:"settings"`
	// ExtraData is shown and never read: stored opaquely, never copied into a
	// Target, never compared.
	ExtraData []ExtraDatum `json:"extraData"`
	// Certificate describes the certificate installed at that point, when the
	// Plugin read one; absent when there is none yet or it could not be read.
	// It is a statement about a concrete certificate, so a Plugin that knows
	// names only from configuration, a virtual host's say, leaves it out.
	Certificate *DiscoveredCertificate `json:"certificate,omitempty"`
}

// DiscoveredCertificate is what a Plugin read from the certificate installed
// at a discovered Target. Typed rather than extra data because the Server
// acts on it.
type DiscoveredCertificate struct {
	// Domains are the names the certificate holds, in the order its subject
	// alternative name extension lists them, addresses and wildcards
	// included; what is requestable is the Server's rule.
	Domains []string `json:"domains"`
	// KeyType names the algorithm of its key when the Plugin can tell: one of
	// ec_p256, ec_p384, rsa_2048, rsa_3072 or rsa_4096, the key types a
	// Certificate can be requested with. Empty when unknown or another
	// algorithm. A Certificate requested for the Target starts with it. A
	// reader that does not know a value reads it as empty.
	KeyType string `json:"keyType,omitempty"`
	// SHA256 is the lowercase hex SHA-256 of the leaf certificate's DER encoding.
	// It is the one field that asserts a certificate was read, so a Plugin sends
	// it whenever it read one; optional only because a Plugin may be unable to
	// read the file at all. The Server tells from it whether the host still serves
	// what it last deployed, and that a certificate is installed where it has
	// deployed none.
	SHA256 string `json:"sha256,omitempty"`
}

// ExtraDatum is one labeled item of a discovered Target's extra data. The
// Label is plain text per language, a name and not prose; the Value is what
// was found on the host, a literal shown as the characters it holds, never as
// Markdown, and is the same in every language.
type ExtraDatum struct {
	Label l10n.Text `json:"label"`
	Value string    `json:"value"`
}

// DeployRequest is OpDeploy's request. The material arrives as PEM text, not
// as paths: a Plugin that installs through an API needs no file, and one that
// does need files creates them itself, inside WorkDir.
type DeployRequest struct {
	CertificatePem string `json:"certificatePem"`
	ChainPem       string `json:"chainPem"`
	PrivateKeyPem  string `json:"privateKeyPem"`
	// WorkDir is a directory the Agent created for this run and removes after
	// it, however it ends. A Plugin that materializes files puts them there, so
	// a run the Agent had to kill cannot leave key material behind; a Plugin
	// that writes nothing ignores it.
	WorkDir string `json:"workDir"`
	// Settings is the resolved value of every setting this Plugin declares,
	// decided Server-side; the Plugin trusts it and resolves nothing itself.
	Settings map[string]string `json:"settings"`
}

// Code sets s as a code span in the Markdown subset, or leaves it as it is
// when it holds a backtick, which no code span can contain.
func Code(s string) string {
	if s == "" || strings.Contains(s, "`") {
		return s
	}
	return "`" + s + "`"
}

// OpDeploy answers with the process itself: exit 0 means the Certificate is
// installed, anything else means it is not, stdout is what happened for a
// person to read, and stderr is diagnostics. There is no response document,
// because one could contradict the exit status. A Plugin that worked while the
// install did not says so in its own output.
//
// A Plugin that must one day report structured facts about what it changed
// gets a channel for them with a new contract version, which the Agent uses
// only with Plugins that report it, as it does a new operation.
//
// `describe` and `discover` keep their JSON because they return structure.

// SummarySetting returns the key of the setting an unlabeled installation point
// is named by, or "" when the Plugin declares none. Meaningful only on a
// declaration set that passed ValidateSettings.
func SummarySetting(decls []SettingDeclaration) string {
	for _, d := range decls {
		if d.Summary {
			return d.Key
		}
	}
	return ""
}

// ValidateSettings checks what a declaration set must satisfy whoever is
// reading it: every Key is a setting key (SettingKeyPattern) and appears once,
// every setting has a Label with English and a Type, at most one is Summary,
// and every Label and Description is a text someone could show
// (l10n.Text.Validate). It lives here so the Agent, at install time, and the
// Server, when the report arrives, reach the same verdict. A Type this build
// does not know is accepted: a reader shows it as text.
//
// Two summary settings parse and every field is valid, so without this an
// unlabeled Target is named after whichever the Plugin declared first, which
// changes if its author reorders the list.
func ValidateSettings(decls []SettingDeclaration) error {
	var summaries []string
	seen := map[string]bool{}
	for _, d := range decls {
		if !settingKey.MatchString(d.Key) {
			return fmt.Errorf("plugincontract: %q is not a setting key: a letter, then up to 63 letters, digits or underscores", d.Key)
		}
		if seen[d.Key] {
			return fmt.Errorf("plugincontract: setting %q is declared twice", d.Key)
		}
		seen[d.Key] = true
		if d.Type == "" {
			return fmt.Errorf("plugincontract: setting %q has no type", d.Key)
		}
		if d.Label[l10n.English] == "" {
			return fmt.Errorf("plugincontract: setting %q has no label in English", d.Key)
		}
		if d.Summary {
			summaries = append(summaries, d.Key)
		}
		if err := d.Label.Validate(); err != nil {
			return fmt.Errorf("plugincontract: the label of setting %q: %w", d.Key, err)
		}
		if err := d.Description.Validate(); err != nil {
			return fmt.Errorf("plugincontract: the description of setting %q: %w", d.Key, err)
		}
	}
	if len(summaries) > 1 {
		return &TooManySummaryError{Keys: summaries}
	}
	return nil
}

// SettingKeyPattern is what a setting key looks like: a letter, then up to 63
// letters, digits or underscores. A key is stored against values forever, so
// it is held to characters every reader can carry.
const SettingKeyPattern = `^[A-Za-z][A-Za-z0-9_]{0,63}$`

var settingKey = regexp.MustCompile(SettingKeyPattern)

// The bounds on a Plugin's answers. They are part of the contract, checked by
// the Agent before it reports a Plugin and by the Server when the report
// arrives, so a Plugin past one is refused on the host, where it shows as
// unobserved, rather than on the Server.
const (
	// MaxVersionBytes bounds a describe answer's version.
	MaxVersionBytes = 64
	// MaxSettings bounds the settings one describe answer declares.
	MaxSettings = 64
	// MaxDiscoveredTargets bounds the Targets one discover answer reports.
	MaxDiscoveredTargets = 200
)

// ValidateDescribe checks a whole describe answer, ValidateSettings included:
// the version is present and bounded, the settings are bounded, and every text
// is one someone could show.
func ValidateDescribe(d DescribeResponse) error {
	if d.Version == "" || len(d.Version) > MaxVersionBytes {
		return fmt.Errorf("plugincontract: the version must be 1 to %d bytes, not %d", MaxVersionBytes, len(d.Version))
	}
	if len(d.Declarations) > MaxSettings {
		return fmt.Errorf("plugincontract: %d settings are declared, more than %d", len(d.Declarations), MaxSettings)
	}
	if err := ValidateSettings(d.Declarations); err != nil {
		return err
	}
	if err := d.Description.Validate(); err != nil {
		return fmt.Errorf("plugincontract: the description: %w", err)
	}
	return nil
}

// ValidateDiscover checks a discover answer against the settings the same
// build declares (DescribeResponse.Declarations): every text is one someone could
// show, a Plugin that reports discovered Targets must flag at least one
// identity setting, and each discovered Target must give every identity
// setting a non-empty value and name only declared settings, since a key the
// declarations do not know could be neither shown nor promoted.
func ValidateDiscover(settings []SettingDeclaration, d DiscoverResponse) error {
	if len(d.DiscoveredTargets) > MaxDiscoveredTargets {
		return fmt.Errorf("plugincontract: %d discovered Targets are reported, more than %d", len(d.DiscoveredTargets), MaxDiscoveredTargets)
	}
	for i, w := range d.Warnings {
		if err := w.Validate(); err != nil {
			return fmt.Errorf("plugincontract: warning %d: %w", i, err)
		}
	}
	for i, t := range d.DiscoveredTargets {
		for _, x := range t.ExtraData {
			if err := x.Label.Validate(); err != nil {
				return fmt.Errorf("plugincontract: an extra-data label of discovered Target %d: %w", i, err)
			}
		}
	}
	if len(d.DiscoveredTargets) == 0 {
		return nil
	}
	declared := map[string]bool{}
	var identity []string
	for _, s := range settings {
		declared[s.Key] = true
		if s.Identity {
			identity = append(identity, s.Key)
		}
	}
	if len(identity) == 0 {
		return errors.New("plugincontract: a Plugin that reports discovered Targets must mark at least one setting identity")
	}
	for i, t := range d.DiscoveredTargets {
		for k := range t.Settings {
			if !declared[k] {
				return fmt.Errorf("plugincontract: discovered Target %d sets %q, which is not a declared setting", i, k)
			}
		}
		for _, k := range identity {
			if t.Settings[k] == "" {
				return fmt.Errorf("plugincontract: discovered Target %d gives identity setting %q no value", i, k)
			}
		}
	}
	return nil
}

// TooManySummaryError names the settings that claimed it, because the Plugin's
// author is the one who has to pick, and the keys are how they find them.
type TooManySummaryError struct{ Keys []string }

func (e *TooManySummaryError) Error() string {
	return "plugincontract: at most one setting may be marked summary, but " +
		strings.Join(e.Keys, ", ") + " all are"
}
