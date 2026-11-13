# Changelog

Every release of `easyssl-contract`, newest first. What a version number
promises is in [README.md](README.md#versioning); `scripts/release.sh` tags
no version this file has no entry for.

## Unreleased

- `catalog`: the Agent build a Server hands out (`Response.Agent`: its version
  and one artifact per platform), which an Agent from this version on
  (`AgentUpdateSince`) reads to update itself. Its shape is frozen across
  breaking versions.
- `catalog`: each Plugin's `contractVersion`, the contract its build was made
  against, so an Agent updating itself can tell whether it would still run the
  Plugin builds it is about to install.
- `catalog`: `Release`, one finished release of a Plugin or of the Agent as its
  `releases.json` records it, listed as each Plugin's `releases` and the
  response's `agentReleases`.

## 0.2.0

Breaking, for moving certificates out of another ACME client's directory:
the Agent, no longer each Plugin, refuses to deploy into a directory another
ACME client manages. A 0.1 Agent would run a 0.2 Plugin with nothing refusing
those paths, so this release moves the breaking version and such an Agent
refuses such a Plugin; a 0.2 Agent still runs 0.1 Plugins, which keep their
own check, and a Server still answers 0.1 Agents (`PreviousBreakingVersion`).

- `plugincontract`: a discovered Target's `references` (where the
  configuration names its files) and `warnings` (why it is not suggested;
  any warning means not suggested), and the `relocate` operation, which
  rewrites those references and is called only on a Plugin reporting 0.2.0
  or later (`AtLeast`, `RelocateSince`).

## 0.1.0

The first release. It freezes four contracts at their first version; from
here on they change only as Versioning allows.

- `wire`: the protocol between the Agent and the Server. Registration and
  reactivation, check-in, commands and their event stream, the signing work
  and the CSR the Agent returns, the Deployment work and its result, the
  host's inventory, and error responses.
- `plugincontract`: the protocol between the Agent and a Plugin. The
  `describe`, `discover` and `deploy` operations and the setting types a
  Plugin declares.
- `catalog`: the Plugin catalog a Server publishes for its Marketplace.
- `l10n`: text carried in every language, English and Spanish so far.

Each message's example for this version is kept under `testdata/v0.1.0/` in
its package, and every later version must still decode it.
