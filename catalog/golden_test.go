package catalog_test

import (
	"testing"

	"github.com/easyssl-io/easyssl-contract/catalog"
	"github.com/easyssl-io/easyssl-contract/contracttest"
	"github.com/easyssl-io/easyssl-contract/l10n"
	"github.com/easyssl-io/easyssl-contract/plugincontract"
	"github.com/easyssl-io/easyssl-contract/wire"
)

// The catalog as a Server publishes it and an Agent reads it.
func TestMessages(t *testing.T) {
	contracttest.Messages(t, "testdata", map[string]any{
		"response": catalog.Response{Plugins: []catalog.Plugin{{
			ID: "generic-cli", Version: "0.0.1",
			Description: l10n.Text{"en": "Installs a Certificate on a host by running a command you write.", "es": "Instala un Certificado en un host ejecutando un comando que tú escribes."},
			Declarations: []plugincontract.SettingDeclaration{{
				Key: "install", Label: l10n.Text{"en": "Install command", "es": "Comando de instalación"},
				Description: l10n.Text{"en": "The command that installs a new Certificate on this Target and reloads the service that uses it.", "es": "El comando que instala un nuevo Certificado en este Destino y recarga el servicio que lo usa."},
				Type:        plugincontract.SettingTypeCommand, Required: true,
			}},
			Party: catalog.PartyFirst, Publisher: "easySSL", SourceURL: "https://github.com/easyssl-io/easyssl-plugins/tree/main/generic-cli",
			Artifacts: []catalog.Artifact{{
				OS: "linux", Arch: "amd64", SHA256: "1ee54b71705e472b4d06ba5c341e0f4acc0a0735be8b1c34f1ade4c6a4e77823",
				URL: wire.APIPrefix + catalog.PluginArtifactPath("generic-cli", "linux", "amd64"),
			}},
		}},
			Agent: &catalog.Agent{Version: "0.2.2", Artifacts: []catalog.Artifact{{
				OS: "linux", Arch: "amd64", SHA256: "5a713b52153da8911d0abdb7f32894138ebb58e533bb257021dcc912b98e02cb",
				URL: wire.APIPrefix + catalog.AgentArtifactPath("linux", "amd64"),
			}}},
		},
	})
}
