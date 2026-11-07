// Package examples holds the example values that more than one package's
// contract examples show, so the documents that repeat them are built from one
// value and cannot disagree. They follow one nginx Target from what the Plugin
// declares, through what it discovers, to the Deployment that carries it.
package examples

import (
	"github.com/easyssl-io/easyssl-contract/l10n"
	"github.com/easyssl-io/easyssl-contract/plugincontract"
)

// NginxSettings is the nginx Target's settings, as discovered and as deployed.
func NginxSettings() map[string]string {
	return map[string]string{
		"ssl_certificate":     "/etc/nginx/tls/example.com.crt",
		"ssl_certificate_key": "/etc/nginx/tls/example.com.key",
		"test":                "nginx -t",
		"reload":              "systemctl reload nginx",
	}
}

// Describe is the nginx Plugin's describe answer, which the inventory also carries.
func Describe() plugincontract.DescribeResponse {
	return plugincontract.DescribeResponse{
		ContractVersion: plugincontract.ContractVersion, Version: "0.0.1",
		Description: l10n.Text{"en": "Installs Certificates into nginx.", "es": "Instala Certificados en nginx."},
		Declarations: []plugincontract.SettingDeclaration{{
			Key: "ssl_certificate", Label: l10n.Text{"en": "ssl_certificate"},
			Description: l10n.Text{"en": "The file nginx reads the Certificate from, written as the Certificate followed by its chain.", "es": "El archivo del que nginx lee el Certificado, escrito como el Certificado seguido de su cadena."},
			Type:        plugincontract.SettingTypePath, Required: true, Summary: true, Identity: true,
		}, {
			Key: "ssl_certificate_key", Label: l10n.Text{"en": "ssl_certificate_key"},
			Description: l10n.Text{"en": "The file nginx reads the private key from.", "es": "El archivo del que nginx lee la clave privada."},
			Type:        plugincontract.SettingTypePath, Required: true,
		}, {
			Key: "test", Label: l10n.Text{"en": "Test command", "es": "Comando de prueba"},
			Description: l10n.Text{"en": "Checks the configuration after the files are written and before the reload.", "es": "Comprueba la configuración después de escribir los archivos y antes de la recarga."},
			Type:        plugincontract.SettingTypeCommand, Default: "nginx -t", Required: true,
		}, {
			Key: "reload", Label: l10n.Text{"en": "Reload command", "es": "Comando de recarga"},
			Description: l10n.Text{"en": "Makes nginx load the new files.", "es": "Hace que nginx cargue los archivos nuevos."},
			Type:        plugincontract.SettingTypeCommand, Default: "nginx -s reload", Required: true,
		}},
	}
}

// Discover is the nginx Plugin's discover answer, which the inventory also carries.
func Discover() plugincontract.DiscoverResponse {
	return plugincontract.DiscoverResponse{
		ContractVersion: plugincontract.ContractVersion,
		DiscoveredTargets: []plugincontract.DiscoveredTarget{{
			Settings:  NginxSettings(),
			ExtraData: []plugincontract.ExtraDatum{{Label: l10n.Text{"en": "Locations", "es": "Ubicaciones"}, Value: "pid 812:/etc/nginx/sites-enabled/example.com:14"}},
			Certificate: &plugincontract.DiscoveredCertificate{
				Domains: []string{"example.com"}, KeyType: "ec_p256",
				SHA256: "14a5786863c316cc76d39e84511315bd110308108db2bb5bcce7be45f70f7601",
			},
		}},
		Warnings: []l10n.Text{{
			"en": "`/etc/nginx/tls/old.crt`: its names could not be read, so none are reported: the file holds no PEM certificate",
			"es": "`/etc/nginx/tls/old.crt`: no se pudieron leer sus nombres, así que no se informa ninguno: el archivo no contiene ningún certificado PEM",
		}},
	}
}
