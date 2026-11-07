package wire_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/easyssl-io/easyssl-contract/l10n"
	"github.com/easyssl-io/easyssl-contract/plugincontract"
	"github.com/easyssl-io/easyssl-contract/wire"
)

// The contract's encoding rules, checked on every message type: a list is
// always present, so a reader never tells "empty" from "not sent"; and a time
// is named with an At suffix (notBefore and notAfter, the X.509 names, apart).
// A text (l10n.Text) and a forwarded answer (json.RawMessage) are documents
// rather than lists, and may be absent.
func TestEncodingRules(t *testing.T) {
	documents := map[reflect.Type]bool{
		reflect.TypeFor[l10n.Text]():       true,
		reflect.TypeFor[json.RawMessage](): true,
	}
	timeNames := map[string]bool{"notBefore": true, "notAfter": true}
	for _, v := range []any{
		wire.Diagnostics{}, wire.RegistrationRequest{}, wire.RegistrationResponse{}, wire.CheckinRequest{},
		wire.RegistrationStatus{}, wire.CheckinResponse{}, wire.ErrorResponse{},
		wire.Command{}, wire.CommandsResponse{}, wire.AckRequest{},
		wire.CSRWork{}, wire.CSRSubmission{},
		wire.DeploymentWork{}, wire.DeploymentResult{}, wire.PluginRun{}, wire.OutputChunk{},
		wire.Inventory{}, wire.PluginInventory{}, wire.KeyInfo{},
		plugincontract.SettingDeclaration{}, plugincontract.DescribeResponse{}, plugincontract.DiscoverResponse{},
		plugincontract.DiscoveredTarget{}, plugincontract.DiscoveredCertificate{}, plugincontract.ExtraDatum{},
		plugincontract.DeployRequest{},
	} {
		typ := reflect.TypeOf(v)
		for i := range typ.NumField() {
			f := typ.Field(i)
			name, opts, _ := strings.Cut(f.Tag.Get("json"), ",")
			kind := f.Type.Kind()
			if (kind == reflect.Slice || kind == reflect.Map) && !documents[f.Type] && strings.Contains(opts, "omitempty") {
				t.Errorf("%s.%s is a list and may be omitted", typ.Name(), f.Name)
			}
			if f.Type == reflect.TypeFor[time.Time]() || f.Type == reflect.TypeFor[*time.Time]() {
				if !strings.HasSuffix(name, "At") && !timeNames[name] {
					t.Errorf("%s.%s is a time named %q, without the At suffix", typ.Name(), f.Name, name)
				}
			}
		}
	}
}

// The Organization's name and id are read best effort, so a Server that could
// not read them omits them rather than sending "" for unknown, as the README's "One encoding" rule says.
func TestUnknownOrganizationIsOmitted(t *testing.T) {
	for _, v := range []any{wire.RegistrationResponse{}, wire.RegistrationStatus{}, wire.CheckinResponse{}} {
		raw, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		for _, key := range []string{`"organizationName"`, `"organizationId"`} {
			if strings.Contains(string(raw), key) {
				t.Errorf("%T sends %s when it is unknown: %s", v, key, raw)
			}
		}
	}
}
