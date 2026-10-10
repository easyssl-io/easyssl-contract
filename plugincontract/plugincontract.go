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

import (
	"errors"
	"fmt"
	"strings"

	"github.com/easyssl-io/easyssl-contract/l10n"
)

// ContractVersion is this contract's version, MAJOR.MINOR.BUILD. Additive
// fields bump MINOR, and a Plugin seeing a newer MINOR than it knows ignores
// the fields it does not recognize. Before the first external release it
// remains at the unpublished 0.0.1 baseline; those bump rules apply after that
// release.
const ContractVersion = "0.0.1"

// Operation names, passed as the Plugin's single argv argument. A Plugin
// receiving an operation it does not implement exits non-zero with a line on
// stderr, and the Agent treats the operation as unsupported.
const (
	// OpDiscover asks the Plugin what it is and what it sees on the host: its
	// identity and declared settings, and the installation points it found by
	// parsing the host's configuration, as discovered Targets, with warnings for
	// what it could not turn into one. It takes no request body. The Agent runs
	// it on every check-in, never once at install, so a binary swapped by hand
	// is read from what would run, and the answer rides the host inventory to
	// the Server. Every Plugin implements it; one that inspects nothing answers
	// with both lists empty.
	//
	// Failing to inspect the host is not a failure of the operation: a
	// configuration the Plugin cannot read is a warning beside the
	// declarations. A non-zero exit or a timeout is the operation producing
	// nothing this run, which the Agent reports as the Plugin being present
	// but unobserved.
	OpDiscover = "discover"
	// OpDeploy carries out one Deployment: it puts a Certificate onto whatever
	// the Plugin installs onto, and reloads whatever has to be reloaded.
	// Everything that happens on the host happens inside this one call.
	//
	// It is named for the Deployment record it carries out, not for what it
	// does to the Certificate, and that is deliberate. "install" named four
	// different acts in this system — installing the Agent, installing a
	// Plugin, installing the systemd unit, installing a Certificate — and two
	// of them met inside one file, where the Agent installs the Plugin
	// executable ten lines above building this request. A reader tracing a
	// `dep-` record from the web console through the Server to the host now
	// finds one word the whole way.
	OpDeploy = "deploy"
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
// A later build of a Plugin MUST accept every setting an earlier build
// declared, because Targets configured under the earlier build keep their
// values and resolve against whatever build the host holds now, and a
// Deployment queued before an upgrade runs on the new build with the settings
// resolved before it:
//
//   - a Key keeps its meaning and is never renamed or reused;
//   - a setting a later build adds is optional, or has a Default that
//     reproduces the earlier behavior, and deploy treats it as absent when a
//     request does not carry it;
//   - a setting a later build drops is ignored when a request still carries it.
//
// A build that cannot honor this is a new Plugin, under a new name. Adding a
// Required setting with no Default is allowed but makes every existing Target
// of the Plugin undeployable until someone sets it.
//
// Label and Description, here and on DiscoverResponse, are text per language
// (l10n.Text): a Plugin is the author of its own translations, English always
// among them, and whoever shows the text chooses its reader's language. A
// Plugin cannot be asked for one language, because what it answers is stored
// and read later by people who do not share one.
//
// Description is written, in each language, in a small Markdown subset, which
// is all the web console renders:
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
	Key         string    `json:"key"`
	Label       l10n.Text `json:"label"`
	Description l10n.Text `json:"description"`
	Type        string    `json:"type"`
	Default     string    `json:"default"`
	Required    bool      `json:"required"`
	// Summary marks the one setting whose value an unlabeled installation
	// point of this Plugin is known by on screen. At most one per Plugin;
	// false everywhere means none. ValidateSettings enforces the "at most
	// one", and both the Agent and the Server call it, so a Plugin that
	// declares two is refused rather than having one of them chosen for it.
	Summary bool `json:"summary"`
	// Identity marks a setting whose value identifies an installation point:
	// the Server composes a discovered Target's discovery key from the values
	// of every Identity setting, in declaration order, and matches discovered
	// Targets to Targets on it. Distinct from Summary, which is what a reader sees a Target
	// called; the two happen to coincide for nginx. A Plugin that reports
	// discovered Targets flags at least one (ValidateDiscover).
	Identity bool `json:"identity"`
}

// DiscoverResponse is OpDiscover's answer: who this build is, what it
// declares, and what it found on the host.
type DiscoverResponse struct {
	ContractVersion string `json:"contractVersion"`
	Name            string `json:"name"`
	Version         string `json:"version"`
	// Description is the build's own account of what it installs onto, in the
	// Markdown subset the web console renders, per language; optional.
	Description l10n.Text            `json:"description,omitempty"`
	Settings    []SettingDeclaration `json:"settings"`
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
	// PluginSettings is a value per declared setting key, opaque to the Server
	// and resolved by the Plugin: a path arrives as the file it names, never
	// as the spelling in the configuration. Promotion copies these into the
	// Target.
	PluginSettings map[string]string `json:"pluginSettings"`
	// PluginExtraData is shown and never read: stored opaquely, never copied
	// into a Target, never compared. Values are in the Markdown subset.
	PluginExtraData []ExtraDatum `json:"pluginExtraData"`
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
	// algorithm. A Certificate requested for the Target starts with it.
	KeyType string `json:"keyType,omitempty"`
	// SHA256 is the lowercase hex SHA-256 of the leaf certificate's DER
	// encoding: which exact certificate is installed there. It is the one
	// field that asserts a certificate was read, the others being reportable
	// from elsewhere, so a Plugin sends it whenever it read one. The Server
	// tells from it whether the host still serves what it last deployed, and
	// that a certificate is installed where it has deployed none. Optional
	// only because a Plugin may be unable to read the file at all.
	SHA256 string `json:"sha256,omitempty"`
}

// ExtraDatum is one labeled item of a discovered Target's extra data. The
// Label is text per language; the Value is what was found on the host, a
// literal, and is the same in every language.
type ExtraDatum struct {
	Label l10n.Text `json:"label"`
	Value string    `json:"value"`
}

// DeployRequest is OpDeploy's request. The material arrives as PEM text, not
// as paths: a Plugin that installs through an API needs no file, and one that
// does need files creates them itself, inside WorkDir.
type DeployRequest struct {
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

// OpDeploy answers with the process itself: exit 0 means the Certificate is
// installed, anything else means it is not, stdout is what happened for a
// person to read, and stderr is diagnostics. There is no response document.
//
// There was one, carrying `ok` and `output`, and it earned its place while the
// Deployment record needed structure the process could not give. The record
// now keeps the exit status, the duration and both streams itself, so the
// document repeated what the Agent already had — in a form free to contradict
// it. A Plugin could exit 0 and say it failed, or exit 3 and say it worked, and
// something had to decide which to believe.
//
// What that field could say and an exit code cannot is "the Plugin worked and
// the install did not". It is not lost, it moved: a Plugin says so in its own
// output, in its own words, which is where a reader looks anyway. What is gone
// is a Plugin's ability to claim one thing while its exit status says another.
//
// The cost, stated because it is the one that matters later: there is no
// envelope to add a field to. A Plugin that must one day report structured
// facts about what it changed needs a mechanism that does not exist here, and
// once this contract is published, reintroducing one is a breaking change for
// every Plugin. Nothing wants it today.
//
// `discover` keeps its JSON, because it genuinely returns structure. The
// asymmetry is the two operations being different in kind: one answers a
// question, the other does work.

// SummarySetting returns the key of the setting an unlabeled installation point
// is named by, or "" when the Plugin declares none. Meaningful only on a
// declaration set that passed ValidateSettings, which is what makes "the
// summary setting" a single thing rather than "the first of them".
func SummarySetting(decls []SettingDeclaration) string {
	for _, d := range decls {
		if d.Summary {
			return d.Key
		}
	}
	return ""
}

// ValidateSettings checks what a declaration set must satisfy whoever is
// reading it: at most one setting may be Summary, and every Label and
// Description is a text someone could show (l10n.Text.Validate).
//
// It is here rather than in either caller because both the Agent, at install
// time, and the Server, when the report arrives, have to reach the same
// verdict. A rule enforced in one of them is a rule the other disagrees with
// the moment somebody reaches it another way.
//
// Two summary settings is not a shape problem — it parses, and every field is
// valid — which is exactly why it needs saying. Without this the Server names
// an unlabeled Target after whichever the Plugin happened to declare first,
// which is a choice nobody made and which changes if the Plugin's author
// reorders its list.
func ValidateSettings(decls []SettingDeclaration) error {
	var summaries []string
	for _, d := range decls {
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

// TooManySummaryError names the settings that claimed it, because the Plugin's
// author is the one who has to pick, and the keys are how they find them.
// ValidateDiscover checks a whole discover answer, ValidateSettings included:
// every text is one someone could show, a Plugin that reports discovered Targets must flag at least one identity
// setting, each discovered Target must give every identity setting a
// non-empty value, and must name only declared settings, since a key the
// declarations do not know could be neither shown nor promoted.
func ValidateDiscover(d DiscoverResponse) error {
	if err := ValidateSettings(d.Settings); err != nil {
		return err
	}
	if err := d.Description.Validate(); err != nil {
		return fmt.Errorf("plugincontract: the description: %w", err)
	}
	for i, w := range d.Warnings {
		if err := w.Validate(); err != nil {
			return fmt.Errorf("plugincontract: warning %d: %w", i, err)
		}
	}
	for i, t := range d.DiscoveredTargets {
		for _, x := range t.PluginExtraData {
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
	for _, s := range d.Settings {
		declared[s.Key] = true
		if s.Identity {
			identity = append(identity, s.Key)
		}
	}
	if len(identity) == 0 {
		return errors.New("plugincontract: a Plugin that reports discovered Targets must mark at least one setting identity")
	}
	for i, t := range d.DiscoveredTargets {
		for k := range t.PluginSettings {
			if !declared[k] {
				return fmt.Errorf("plugincontract: discovered Target %d sets %q, which is not a declared setting", i, k)
			}
		}
		for _, k := range identity {
			if t.PluginSettings[k] == "" {
				return fmt.Errorf("plugincontract: discovered Target %d gives identity setting %q no value", i, k)
			}
		}
	}
	return nil
}

type TooManySummaryError struct{ Keys []string }

func (e *TooManySummaryError) Error() string {
	return "plugincontract: at most one setting may be marked summary, but " +
		strings.Join(e.Keys, ", ") + " all are"
}
