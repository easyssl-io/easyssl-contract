package wire_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/easyssl-io/easyssl-contract/contracttest"
	"github.com/easyssl-io/easyssl-contract/l10n"
	"github.com/easyssl-io/easyssl-contract/wire"
)

// Every message of the protocol, one example each, as a reader on the other
// side receives it.
func TestMessages(t *testing.T) {
	diagnostics := wire.Diagnostics{
		Hostname: "web-01", OS: "linux/amd64", HostType: "vm",
		BinaryPath: "/usr/local/bin/easyssl-agent", ConfigPath: "/etc/easyssl/config.json", ServiceName: "easyssl-agent.service",
	}
	contracttest.Messages(t, "testdata", map[string]any{
		"registration-request": wire.RegistrationRequest{RegistrationKey: "rk_example", ProposedName: "web-01", Diagnostics: diagnostics},
		"registration-response": wire.RegistrationResponse{
			AgentToken: "at_example", AgentID: "agnt-x2nfigyipdlpv5", Status: wire.AgentStatusPending, DisplayName: "web-01",
			RequiredPlugins: []string{"nginx"}, NextPollSeconds: 30, MarketplaceURL: "/api/v1/marketplace/catalog",
			OrganizationName: "Northwind", OrganizationID: "org-northwindaaaaaa",
		},
		"registration-status": wire.RegistrationStatus{Status: wire.AgentStatusActive, DisplayName: "web-01", AgentID: "agnt-x2nfigyipdlpv5"},
		"checkin-request":     wire.CheckinRequest{InventoryFingerprint: "3b1f0c9a"},
		"checkin-response": wire.CheckinResponse{
			Status: wire.AgentStatusActive, DisplayName: "web-01", NextPollSeconds: 30, ReportInventory: true, PendingCommands: 1,
		},
		"error-response": wire.ErrorResponse{
			Code: wire.CodeContractUnsupported, Params: map[string]string{"version": "0.0.1"},
			Message: "this Server does not speak contract version 0.0.1",
		},
		"commands-response": wire.CommandsResponse{Commands: []wire.Command{{ID: "cmd-aaaaaaaaaaaaaa", Type: wire.CommandRequestCSR, RefID: "crt-aaaaaaaaaaaaaa"}}},
		"command":           wire.Command{ID: "cmd-bbbbbbbbbbbbbb", Type: wire.CommandDeploy, RefID: "dep-aaaaaaaaaaaaaa"},
		"ack-request":       wire.AckRequest{Status: wire.CommandStatusAcked},
		"csr-work":          wire.CSRWork{Domains: []string{"example.com", "*.example.com"}, KeyType: wire.KeyTypeECP256},
		"csr-submission":    wire.CSRSubmission{CSR: "-----BEGIN CERTIFICATE REQUEST-----\nMIIB...\n-----END CERTIFICATE REQUEST-----\n"},
		"deployment-work": wire.DeploymentWork{
			Status: wire.DeploymentStatusPending, CertificateID: "crt-aaaaaaaaaaaaaa", KeyFingerprint: "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
			PluginID: "nginx", CertificatePem: "-----BEGIN CERTIFICATE-----\nMIIB...\n-----END CERTIFICATE-----\n",
			ChainPem: "-----BEGIN CERTIFICATE-----\nMIIC...\n-----END CERTIFICATE-----\n",
			Settings: map[string]string{"certificatePath": "/etc/nginx/tls/example.com.crt", "keyPath": "/etc/nginx/tls/example.com.key"},
		},
		"deployment-result": wire.DeploymentResult{
			OK: true, PluginVersion: "0.0.1",
			PluginRun: &wire.PluginRun{
				ExitStatus: 0, DurationMs: 412, StdoutBytes: 31, StderrBytes: 0,
				Output: []wire.OutputChunk{{Stream: wire.StreamStdout, Text: "nginx reloaded; serving the new certificate\n"}},
			},
		},
		"deployment-result-failed": wire.DeploymentResult{
			OK: false, PluginVersion: "0.0.1",
			AgentMessage: l10n.Text{"en": "The Plugin was not installed.", "es": "El Plugin no está instalado."},
		},
		"inventory": wire.Inventory{
			Fingerprint: "3b1f0c9a", Diagnostics: diagnostics,
			Plugins: []wire.PluginInventory{{
				ID: "nginx", SHA256: "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
				Describe: json.RawMessage(`{"contractVersion":"0.0.1","version":"0.0.1","declarations":[]}`),
				Discover: json.RawMessage(`{"contractVersion":"0.0.1","discoveredTargets":[],"warnings":[]}`),
				Warnings: []l10n.Text{},
			}},
			Keys: []wire.KeyInfo{{
				Fingerprint: "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08", Algorithm: wire.KeyTypeECP256,
				CreatedAt: time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC),
			}},
		},
	})
}
