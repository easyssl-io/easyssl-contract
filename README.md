# easyssl-contract

The contracts easySSL's separately installed programs share, as Go types with
their JSON tags:

- `wire`: the Server and the Agent, over HTTPS.
- `plugincontract`: the Agent and each Plugin, one execution per operation.
- `l10n`: text carried in every supported language, used by both.

The JSON tag is the contract, not the Go identifier. This module has no
dependencies, so it can be audited quickly.

## Versioning

An Agent installed on a host cannot be upgraded by force, and a Plugin is
released apart from the Agent that runs it, so these rules hold from the first
release on.

1. **One version for the whole module**, its release tag, a semantic version
   that starts at `v0.1.0`. `wire`, `plugincontract` and `l10n` move together,
   since `wire` carries types of the other two.
2. **What a version change means.** While at `0.x`, a breaking change moves the
   minor number and anything else the patch number. From `1.0.0` on, a
   breaking change moves the major number, an addition the minor number and a
   fix the patch number.
3. **Changes are additive.** A new field is optional and omitted when empty; a
   new endpoint, command type, operation or enum value may be added. A reader
   ignores fields it does not know and refuses only JSON it cannot parse. Each
   enum says what a reader does with a value it does not know, and the older
   side never refuses one.
4. **Breaking is two steps.** The new way is added beside the old one and the
   old one is marked deprecated in a release; it is removed only in a later
   breaking release.
5. **The Server supports the current breaking version and the one before.** An
   Agent older than that is refused with a code that says so, rather than
   failing at random; an Agent on the previous one is shown as needing an
   update before the next breaking release.
6. **Each side says its version.** The Agent sends the contract version with
   every request to the Server, and a Plugin reports the one it was built
   against in its `describe` answer, which the Agent checks before running it.
   A new operation is called only on a Plugin built against a version that has
   it.
7. **A released message stays readable.** Every message has an example kept per
   released version, which every later build must still decode.
8. **Versions come from release tags**, never from a constant edited by hand.
9. **One encoding.** A time is RFC 3339 in UTC and its name ends in `At`
   (`notBefore` and `notAfter` keep their X.509 names). A list is always
   present, `[]` when empty, so a reader never tells "none" from "not sent". A
   scalar that is unknown or does not apply is omitted, never sent as `""` or
   a zero time standing for something. `TestEncodingRules` checks the names
   and the lists.

Until the first release everything stays at `0.0.1`.
