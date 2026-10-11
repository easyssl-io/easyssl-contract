package wire

// ProtocolVersion is the easySSL Agent-to-Server wire-protocol version,
// formatted MAJOR.MINOR.BUILD. MAJOR is bumped only for breaking contract
// changes, MINOR for additive backward-compatible features, BUILD for
// backward-compatible fixes. It stays at 0.0.1 until the first release, and
// those rules apply from then on.
const ProtocolVersion = "0.0.1"
