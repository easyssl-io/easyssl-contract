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
	// parsing the host's configuration, as suggested Targets, with warnings for
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
// Description, here and on DiscoverResponse, is written in a small Markdown
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
	// false everywhere means none. ValidateSettings enforces the "at most
	// one", and both the Agent and the Server call it, so a Plugin that
	// declares two is refused rather than having one of them chosen for it.
	Summary bool `json:"summary"`
	// Identity marks a setting whose value identifies an installation point:
	// the Server composes a suggested Target's discovery key from the values of
	// every Identity setting, in declaration order, and matches suggestions to
	// Targets on it. Distinct from Summary, which is what a reader sees a Target
	// called; the two happen to coincide for nginx. A Plugin that reports
	// suggestions flags at least one (ValidateDiscover).
	Identity bool `json:"identity"`
}

// DiscoverResponse is OpDiscover's answer: who this build is, what it
// declares, and what it found on the host.
type DiscoverResponse struct {
	ContractVersion string `json:"contractVersion"`
	Name            string `json:"name"`
	Version         string `json:"version"`
	// Description is the build's own account of what it installs onto, in the
	// Markdown subset the web console renders; optional.
	Description string               `json:"description,omitempty"`
	Settings    []SettingDeclaration `json:"settings"`
	// SuggestedTargets are the installation points the Plugin parsed, each an
	// offer to add a Target from the web console and never a Target itself.
	// Empty for a Plugin that inspects nothing.
	SuggestedTargets []SuggestedTarget `json:"suggestedTargets"`
	// Warnings are what discovery saw and could not turn into a suggestion, in
	// the Markdown subset, shown on the Agent page under the Plugin.
	Warnings []string `json:"warnings"`
}

// SuggestedTarget is one installation point a Plugin found. It carries no
// key of its own: the Server composes one from the Identity settings.
type SuggestedTarget struct {
	// PluginSettings is a value per declared setting key, opaque to the Server
	// and resolved by the Plugin: a path arrives as the file it names, never
	// as the spelling in the configuration. Promotion copies these into the
	// Target.
	PluginSettings map[string]string `json:"pluginSettings"`
	// PluginExtraData is shown and never read: stored opaquely, never copied
	// into a Target, never compared. Values are in the Markdown subset.
	PluginExtraData []ExtraDatum `json:"pluginExtraData"`
	// SuggestedDomains are the names the certificate at that point holds,
	// addresses and wildcards included; what is requestable is the Server's
	// rule.
	SuggestedDomains []string `json:"suggestedDomains"`
}

// ExtraDatum is one labeled item of a suggestion's extra data.
type ExtraDatum struct {
	Label string `json:"label"`
	Value string `json:"value"`
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
// reading it. One rule today: at most one setting may be Summary.
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
	}
	if len(summaries) > 1 {
		return &TooManySummaryError{Keys: summaries}
	}
	return nil
}

// TooManySummaryError names the settings that claimed it, because the Plugin's
// author is the one who has to pick, and the keys are how they find them.
// ValidateDiscover checks a whole discover answer, ValidateSettings included:
// a Plugin that reports suggestions must flag at least one identity setting,
// each suggestion must give every identity setting a non-empty value, and
// must name only declared settings, since a key the declarations do not know
// could be neither shown nor promoted.
func ValidateDiscover(d DiscoverResponse) error {
	if err := ValidateSettings(d.Settings); err != nil {
		return err
	}
	if len(d.SuggestedTargets) == 0 {
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
		return errors.New("plugincontract: a Plugin that reports suggested Targets must mark at least one setting identity")
	}
	for i, t := range d.SuggestedTargets {
		for k := range t.PluginSettings {
			if !declared[k] {
				return fmt.Errorf("plugincontract: suggestion %d sets %q, which is not a declared setting", i, k)
			}
		}
		for _, k := range identity {
			if t.PluginSettings[k] == "" {
				return fmt.Errorf("plugincontract: suggestion %d gives identity setting %q no value", i, k)
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
