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

// Discover is the nginx Plugin's discover answer, which the inventory also carries:
// the Target it suggests, one it found but does not suggest, and a warning about
// what it could not read at all.
func Discover() plugincontract.DiscoverResponse {
	return plugincontract.DiscoverResponse{
		ContractVersion: plugincontract.ContractVersion,
		DiscoveredTargets: []plugincontract.DiscoveredTarget{{
			Settings:  NginxSettings(),
			ExtraData: []plugincontract.ExtraDatum{},
			Certificate: &plugincontract.DiscoveredCertificate{
				Domains: []string{"example.com"}, KeyType: "ec_p256",
				SHA256:         "14a5786863c316cc76d39e84511315bd110308108db2bb5bcce7be45f70f7601",
				CertificatePem: "-----BEGIN CERTIFICATE-----\nMIIB...\n-----END CERTIFICATE-----\n",
				ChainPem:       "-----BEGIN CERTIFICATE-----\nMIIC...\n-----END CERTIFICATE-----\n",
			},
			References: []plugincontract.Reference{
				{Setting: "ssl_certificate", File: "/etc/nginx/sites-enabled/example.com", Line: 14, Value: "/etc/nginx/tls/example.com.crt"},
				{Setting: "ssl_certificate_key", File: "/etc/nginx/sites-enabled/example.com", Line: 15, Value: "/etc/nginx/tls/example.com.key"},
			},
			Warnings: []l10n.Text{},
		}, {
			Settings: map[string]string{
				"ssl_certificate":     "/etc/letsencrypt/archive/example.org/fullchain2.pem",
				"ssl_certificate_key": "/etc/letsencrypt/archive/example.org/privkey2.pem",
				"test":                "nginx -t",
				"reload":              "systemctl reload nginx",
			},
			ExtraData: []plugincontract.ExtraDatum{},
			Certificate: &plugincontract.DiscoveredCertificate{
				Domains: []string{"example.org", "www.example.org"}, KeyType: "ec_p256",
				SHA256: "5e0f7a3c1b9d24e86f70a2c4d8b6e1f3a5c7e9b1d3f5a7c9e1b3d5f7a9c1e3b5",
			},
			References: []plugincontract.Reference{
				{Setting: "ssl_certificate", File: "/etc/nginx/sites-enabled/example.org", Line: 9, Value: "/etc/letsencrypt/live/example.org/fullchain.pem"},
				{Setting: "ssl_certificate_key", File: "/etc/nginx/sites-enabled/example.org", Line: 10, Value: "/etc/letsencrypt/live/example.org/privkey.pem"},
			},
			Warnings: []l10n.Text{{
				"en": "`/etc/letsencrypt/archive/example.org/fullchain2.pem` is in certbot's directory `/etc/letsencrypt`, so easySSL does not deploy there. Run `sudo easyssl-agent migrate` to move it out.",
				"es": "`/etc/letsencrypt/archive/example.org/fullchain2.pem` está en el directorio de certbot `/etc/letsencrypt`, así que easySSL no despliega ahí. Ejecuta `sudo easyssl-agent migrate` para sacarlo de ahí.",
			}},
		}},
		Warnings: []l10n.Text{{
			"en": "pid 977 (`/usr/sbin/nginx`): its configuration could not be read: `nginx -t` failed",
			"es": "pid 977 (`/usr/sbin/nginx`): no se pudo leer su configuración: `nginx -t` falló",
		}},
	}
}

// Relocate is the request that moves the second discovered Target of Discover
// out of certbot's directory.
func Relocate() plugincontract.RelocateRequest {
	return plugincontract.RelocateRequest{
		Moves: []plugincontract.Move{
			{From: "/etc/letsencrypt/live/example.org/fullchain.pem", To: "/etc/ssl/easyssl/example.org/certificate.pem"},
			{From: "/etc/letsencrypt/live/example.org/privkey.pem", To: "/etc/ssl/easyssl/example.org/key.pem"},
		},
		References: Discover().DiscoveredTargets[1].References,
	}
}
