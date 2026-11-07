package wire_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
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
		// Resuming a deactivated Agent: the token replaces the key, and no name is
		// proposed, since the Agent keeps the one it had.
		"registration-request-reactivation": wire.RegistrationRequest{ReactivationToken: "rt_example", Diagnostics: diagnostics},
		"registration-response": wire.RegistrationResponse{
			AgentToken: "at_example", AgentID: "agnt-x2nfigyipdlpv5", Status: wire.AgentStatusPending, DisplayName: "web-01",
			RequiredPlugins: []string{"nginx"}, NextPollSeconds: 30, MarketplaceURL: "https://console.example.com/plugins/marketplace",
			OrganizationName: "Northwind", OrganizationID: "org-northwindaaaaaa",
		},
		"registration-status": wire.RegistrationStatus{
			Status: wire.AgentStatusActive, DisplayName: "web-01", AgentID: "agnt-x2nfigyipdlpv5",
			OrganizationName: "Northwind", OrganizationID: "org-northwindaaaaaa",
		},
		"checkin-request": wire.CheckinRequest{InventoryFingerprint: "5994471abb01112afcc18159f6cc74b4f511b99806da59b3caf5a9c173cacfc5"},
		"checkin-response": wire.CheckinResponse{
			Status: wire.AgentStatusActive, DisplayName: "web-01", NextPollSeconds: 30, ReportInventory: true, PendingCommands: 1,
			OrganizationName: "Northwind", OrganizationID: "org-northwindaaaaaa",
		},
		"error-response": wire.ErrorResponse{
			Code: wire.CodeContractUnsupported, Params: map[string]string{"version": "0.0.1"},
			Message: "this Server does not speak contract version 0.0.1",
		},
		"commands-response": wire.CommandsResponse{Commands: []wire.Command{{ID: "cmd-aaaaaaaaaaaaaa", Type: wire.CommandRequestCSR, RefID: "crt-aaaaaaaaaaaaaa"}}},
		"command":           wire.Command{ID: "cmd-bbbbbbbbbbbbbb", Type: wire.CommandDeploy, RefID: "dep-aaaaaaaaaaaaaa"},
		"ack-request":       wire.AckRequest{Status: wire.CommandStatusAcked},
		"csr-work":          wire.CSRWork{Domains: []string{"example.com", "*.example.com"}, KeyType: wire.KeyTypeECP256},
		// Signing with a key the host already holds: a renewal that keeps its key,
		// or a re-run after a CSR was stored.
		"csr-work-existing-key": wire.CSRWork{
			Domains: []string{"example.com", "*.example.com"}, KeyType: wire.KeyTypeECP256,
			KeySPKISHA256: "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
		},
		"csr-submission": wire.CSRSubmission{CSR: "-----BEGIN CERTIFICATE REQUEST-----\nMIIB...\n-----END CERTIFICATE REQUEST-----\n"},
		"deployment-work": wire.DeploymentWork{
			Status: wire.DeploymentStatusPending, CertificateID: "crt-aaaaaaaaaaaaaa", KeySPKISHA256: "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
			PluginID: "nginx", CertificatePem: "-----BEGIN CERTIFICATE-----\nMIIB...\n-----END CERTIFICATE-----\n",
			ChainPem: "-----BEGIN CERTIFICATE-----\nMIIC...\n-----END CERTIFICATE-----\n",
			// The same nginx Target as plugincontract's examples.
			Settings: map[string]string{
				"ssl_certificate":     "/etc/nginx/tls/example.com.crt",
				"ssl_certificate_key": "/etc/nginx/tls/example.com.key",
				"test":                "nginx -t",
				"reload":              "systemctl reload nginx",
			},
		},
		"deployment-result": wire.DeploymentResult{
			OK: true, PluginVersion: "0.0.1",
			PluginRun: &wire.PluginRun{
				ExitStatus: 0, DurationMs: 412, StdoutBytes: 44, StderrBytes: 0,
				Output: []wire.OutputChunk{{Stream: wire.StreamStdout, Text: "nginx reloaded; serving the new certificate\n"}},
			},
		},
		"deployment-result-failed": wire.DeploymentResult{
			OK: false, PluginVersion: "0.0.1",
			AgentMessage: l10n.Text{"en": "The Plugin was not installed.", "es": "El Plugin no está instalado."},
		},
		"inventory": wire.Inventory{
			Fingerprint: "5994471abb01112afcc18159f6cc74b4f511b99806da59b3caf5a9c173cacfc5", Diagnostics: diagnostics,
			Plugins: []wire.PluginInventory{{
				ID: "nginx", SHA256: "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
				Describe: pluginAnswer(t, "describe-response"),
				Discover: pluginAnswer(t, "discover-response"),
				Warnings: []l10n.Text{},
			}, {
				// A Plugin whose answers were unusable this run: no answers, and the
				// Agent's own account of why.
				ID: "apache", SHA256: "fcde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9",
				Unobserved: true,
				Warnings: []l10n.Text{{
					"en": "`discover` produced nothing usable this run: exit status 1",
					"es": "`discover` no produjo nada utilizable en esta ejecución: exit status 1",
				}},
			}},
			Keys: []wire.KeyInfo{{
				SPKISHA256: "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08", KeyType: wire.KeyTypeECP256,
				CreatedAt: time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC),
			}},
		},
	})
}

// pluginAnswer is one of plugincontract's examples, which the inventory carries
// as the Plugin wrote it; reading the file keeps the two examples one document.
func pluginAnswer(t *testing.T, name string) json.RawMessage {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "plugincontract", "testdata", name+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, b); err != nil {
		t.Fatal(err)
	}
	return compact.Bytes()
}
