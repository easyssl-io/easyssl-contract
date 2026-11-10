package plugincontract_test

import (
	"testing"

	"github.com/easyssl-io/easyssl-contract/contracttest"
	"github.com/easyssl-io/easyssl-contract/internal/examples"
	"github.com/easyssl-io/easyssl-contract/plugincontract"
)

// Every document a Plugin reads or writes, one example each.
func TestMessages(t *testing.T) {
	contracttest.Messages(t, "testdata", map[string]any{
		"describe-response": examples.Describe(),
		"discover-response": examples.Discover(),
		"relocate-request":  examples.Relocate(),
		"deploy-request": plugincontract.DeployRequest{
			CertificatePem: "-----BEGIN CERTIFICATE-----\nMIIB...\n-----END CERTIFICATE-----\n",
			ChainPem:       "-----BEGIN CERTIFICATE-----\nMIIC...\n-----END CERTIFICATE-----\n",
			PrivateKeyPem:  "-----BEGIN PRIVATE KEY-----\nMIGH...\n-----END PRIVATE KEY-----\n",
			WorkDir:        "/var/lib/easyssl/work/dep-aaaaaaaaaaaaaa",
			Settings:       examples.NginxSettings(),
		},
	})
}
