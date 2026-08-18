// Package wire defines the versioned easySSL Agent-to-Server contract.
//
// It is the single, public source of truth for the backward-compatible wire
// protocol that both the open-source Agent and the Server depend on. Keep this
// package dependency-light so the Agent's compile-time surface stays auditable.
package wire
