package catalog_test

import (
	"testing"

	"github.com/easyssl-io/easyssl-contract/catalog"
	"github.com/easyssl-io/easyssl-contract/contracttest"
	"github.com/easyssl-io/easyssl-contract/l10n"
	"github.com/easyssl-io/easyssl-contract/plugincontract"
)

// The catalog as a Server publishes it and an Agent reads it.
func TestMessages(t *testing.T) {
	contracttest.Messages(t, "testdata", map[string]any{
		"response": catalog.Response{Plugins: []catalog.Plugin{{
			ID: "generic-cli", Version: "0.0.1",
			Description:  l10n.Text{"en": "Runs your `command` after writing the files.", "es": "Ejecuta tu `command` tras escribir los archivos."},
			Declarations: []plugincontract.SettingDeclaration{},
			Party:        catalog.PartyFirst, Publisher: "easySSL", SourceURL: "https://github.com/easyssl-io/easyssl-plugins/tree/main/generic-cli",
			Artifacts: []catalog.Artifact{{
				OS: "linux", Arch: "amd64", SHA256: "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
				URL: catalog.PluginArtifactPath("generic-cli", "linux", "amd64"),
			}},
		}}},
	})
}
