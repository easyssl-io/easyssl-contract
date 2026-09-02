package wire

// ProtocolVersion is the easySSL Agent-to-Server wire-protocol version,
// formatted MAJOR.MINOR.BUILD. MAJOR is bumped only for breaking contract
// changes, MINOR for additive backward-compatible features, BUILD for
// backward-compatible fixes.
//
// 0.1.0 adds ProtocolVersion to CheckinRequest, which is additive and so bumps
// MINOR: a Server that knows the field reads it from Agents that send it and
// keeps what it stored for Agents that do not, and an Agent that does not know
// it is unaffected. This is the first bump off the unpublished 0.0.1 baseline,
// and it is what makes the value worth reporting on a check-in at all: until
// there were two of them, a stale record and a current one could not differ.
const ProtocolVersion = "0.1.0"
