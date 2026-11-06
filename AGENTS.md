# AGENTS.md

Guidance for AI coding agents working in `easyssl-contract`. The file is named
`AGENTS.md` because that is the name the tools read; "agent" in this file means
an AI coding agent, while **Agent** capitalized means easySSL's own
host-installed component. `README.md` says what this module is; this file is
what a change to it must respect.

`go build ./... && go test ./... && go vet ./...` is the whole build. No
database, no network, no build tags: a change that needs any of those belongs in
the Agent or the Server instead.

## Code Review Rules

This module is the public interface between separately installed programs: the
Server, the Agent and each Plugin. An Agent already installed on a customer host
cannot be force-upgraded, so one compiled against an older version of this
module MUST keep working against a newer Server, and the same holds between the
Agent and a Plugin, in both directions. The rules are in `README.md`
(Versioning) and apply from the first release; until then everything stays at
`0.0.1`. Flag any of the following.

- **A breaking change made in one step**: removing a field, renaming a JSON tag,
  changing a type, narrowing the values a field may hold, or making an optional
  field required, without first adding the new way beside the old one and
  marking the old one deprecated. Reusing an existing field for a new meaning
  breaks compatibility while looking additive; add a new field instead.
- **A reader that refuses what it does not know**: an unknown field, or an enum
  value with no stated handling.
- **Treating the Go identifier as the contract.** The JSON tag is. Renaming the
  identifier is safe; renaming the tag is not, however cosmetic it looks.
- **Any new `require` in `go.mod`.** This module has zero dependencies and is
  compiled into a binary whose value is that it can be audited quickly.
- **Implementation details leaking in**: database, HTTP framework, ACME library
  or filesystem types, or any type only ever used on one side.

## Conventions

American English. Capitalize Organization, Agent, Server, Plugin, Certificate,
Target and Deployment whenever they name easySSL's own records or components;
write "web console" in full for easySSL's browser UI. Keep a genuinely generic
technical role lowercase, such as an ACME client or a CNAME target. This
repository is public and self-contained: never cite a document, identifier, or
issue that lives in a private repository. State the property directly instead.

A `Description`, on a `SettingDeclaration` or a `DescribeResponse`, is written
in the small Markdown subset stated on `SettingDeclaration.Description`:
backtick spans for anything a reader would type or look for, `**bold**`, a blank
line between paragraphs, a single line break kept as one, and nothing else. HTML
is shown as the characters typed, never rendered. A Plugin's text is somebody
else's text once Plugins are published and the web console must not carry its
markup. Markdown rather than HTML for the same reason: a subset this small has
nothing to sanitize.
