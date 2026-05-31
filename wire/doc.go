// Package wire defines the versioned EasySSL Client-to-Server contract.
//
// It is the single, public source of truth for the backward-compatible wire
// protocol that both the open-source Client and the Server depend on. Keep this
// package dependency-light so the Client's compile-time surface stays auditable.
package wire
