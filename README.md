# easyssl-contract

Shared wire-protocol contract for EasySSL: the versioned types and IDL that both
the EasySSL Client and the EasySSL Server compile against.

This module is intentionally dependency-light and is the public, auditable
definition of the Client-to-Server contract (EasySSL Constitution Principle V,
Auditable Client, and Principle VI, Backward-Compatible Wire Protocol). It is
consumed as a Git submodule by `easyssl-parent` during development and is
imported as a normal Go module by the Client and the Server.

Status: scaffolding. Concrete wire types land here as the transport decision
(REST/SSE vs gRPC) is finalized.
