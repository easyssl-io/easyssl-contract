package wire

// ProtocolVersion is the easySSL Agent-to-Server wire-protocol version,
// formatted MAJOR.MINOR.BUILD. MAJOR is bumped only for breaking contract
// changes, MINOR for additive backward-compatible features, BUILD for
// backward-compatible fixes.
//
// Nothing is released yet, so it stays at 0.0.1 whatever changes, and those
// rules apply from the first release on: until then there is no deployed
// Agent or Server a version could be compatible with, and git history records
// every change.
const ProtocolVersion = "0.0.1"
