// Package wire defines the versioned EasySSL Client-to-Server contract.
//
// It is the single, public source of truth for the wire protocol that both the
// open-source Client and the Server depend on (EasySSL Constitution Principle VI,
// Backward-Compatible Wire Protocol). Keep this package dependency-light so the
// Client's compile-time surface stays auditable (Principle V).
package wire
