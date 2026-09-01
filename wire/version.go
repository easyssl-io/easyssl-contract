package wire

// ProtocolVersion is the easySSL Agent-to-Server wire-protocol version,
// formatted MAJOR.MINOR.BUILD. MAJOR is bumped only for breaking contract
// changes, MINOR for additive backward-compatible features, BUILD for
// backward-compatible fixes. Before the first external release it remains at
// the unpublished 0.0.1 baseline; those bump rules apply after that release.
const ProtocolVersion = "0.0.1"
