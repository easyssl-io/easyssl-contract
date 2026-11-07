package plugincontract_test

import (
	"testing"

	"github.com/easyssl-io/easyssl-contract/contracttest"
	"github.com/easyssl-io/easyssl-contract/l10n"
	"github.com/easyssl-io/easyssl-contract/plugincontract"
)

// Every document a Plugin reads or writes, one example each.
func TestMessages(t *testing.T) {
	contracttest.Messages(t, "testdata", map[string]any{
		"describe-response": plugincontract.DescribeResponse{
			ContractVersion: plugincontract.ContractVersion, Version: "0.0.1",
			Description: l10n.Text{"en": "Runs your `command` after writing the files.", "es": "Ejecuta tu `command` tras escribir los archivos."},
			Declarations: []plugincontract.SettingDeclaration{{
				Key: "command", Label: l10n.Text{"en": "Command", "es": "Comando"},
				Description: l10n.Text{"en": "Runs as root on the host.", "es": "Se ejecuta como root en el host."},
				Type:        plugincontract.SettingTypeCommand, Required: true, Summary: true, Identity: true,
			}},
		},
		"discover-response": plugincontract.DiscoverResponse{
			ContractVersion: plugincontract.ContractVersion,
			DiscoveredTargets: []plugincontract.DiscoveredTarget{{
				Settings:  map[string]string{"certificatePath": "/etc/nginx/tls/example.com.crt"},
				ExtraData: []plugincontract.ExtraDatum{{Label: l10n.Text{"en": "Server name", "es": "Nombre de servidor"}, Value: "example.com"}},
				Certificate: &plugincontract.DiscoveredCertificate{
					Domains: []string{"example.com"}, KeyType: "ec_p256",
					SHA256: "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
				},
			}},
			Warnings: []l10n.Text{{"en": "`/etc/nginx/old.conf` could not be read.", "es": "No se pudo leer `/etc/nginx/old.conf`."}},
		},
		"deploy-request": plugincontract.DeployRequest{
			CertificatePem: "-----BEGIN CERTIFICATE-----\nMIIB...\n-----END CERTIFICATE-----\n",
			ChainPem:       "-----BEGIN CERTIFICATE-----\nMIIC...\n-----END CERTIFICATE-----\n",
			PrivateKeyPem:  "-----BEGIN PRIVATE KEY-----\nMIGH...\n-----END PRIVATE KEY-----\n",
			WorkDir:        "/var/lib/easyssl/work/dep-aaaaaaaaaaaaaa",
			Settings:       map[string]string{"command": "systemctl reload nginx"},
		},
	})
}
