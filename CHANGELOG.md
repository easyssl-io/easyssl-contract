# Changelog

Every release of `easyssl-contract`, newest first. What a version number
promises is in [README.md](README.md#versioning); `scripts/release.sh` tags
no version this file has no entry for.

## 0.1.1

Additions for moving certificates out of another ACME client's directory.

- `plugincontract`: a discovered Target's `references` (where the
  configuration names its files) and `warnings` (why it is not suggested;
  any warning means not suggested), and the `relocate` operation, which
  rewrites those references and is called only on a Plugin reporting 0.1.1
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
