# AGENTS.md

Guidance for coding agents in `easyssl-contract`. `README.md` says what this
module is; this file is what a change to it must respect.

`go build ./... && go test ./... && go vet ./...` is the whole build. No
database, no network, no build tags: a change that needs any of those belongs in
the Client or the Server instead.

## Code Review Rules

This module is a published interface between two independently deployed
programs. A Client already installed on a customer host cannot be force-upgraded,
so one compiled against an older version of this module MUST keep working
against a newer Server. Flag any of the following.

- **Breaking a deployed Client without a MAJOR bump**: removing a field, renaming
  a JSON tag, changing a type, narrowing the values a field may hold, or making
  an optional field required. Reusing an existing field for a new meaning breaks
  compatibility while looking additive; add a new field instead.
- **A `wire/` change that leaves `ProtocolVersion` alone.** MAJOR for breaking,
  MINOR for additive, BUILD for a fix, moved in the same change. A MAJOR bump
  additionally needs a deprecation path for Clients already installed.
- **Treating the Go identifier as the contract.** The JSON tag is. Renaming the
  identifier is safe; renaming the tag is not, however cosmetic it looks.
- **Any new `require` in `go.mod`.** This module has zero dependencies and is
  compiled into a binary whose value is that it can be audited quickly.
- **Implementation details leaking in**: database, HTTP framework, ACME library
  or filesystem types, or any type only ever used on one side.

## Conventions

American English. "Client" and "Server" are capitalized when they mean the
easySSL components; never call the Client an "agent". This repository is public
and self-contained: never cite a document, identifier, or issue that lives in a
private repository. State the property directly instead.
