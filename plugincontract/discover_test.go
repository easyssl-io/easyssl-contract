package plugincontract

import (
	"testing"

	"github.com/easyssl-io/easyssl-contract/l10n"
)

func TestValidateDiscoverRequiresIdentityForDiscoveredTargets(t *testing.T) {
	settings := []SettingDeclaration{{Key: "ssl_certificate"}, {Key: "reload"}}
	d := DiscoverResponse{}
	if err := ValidateDiscover(settings, d); err != nil {
		t.Fatalf("no discovered Targets, no identity: %v", err)
	}
	d.DiscoveredTargets = []DiscoveredTarget{{Settings: map[string]string{"ssl_certificate": "/a.crt"}}}
	if err := ValidateDiscover(settings, d); err == nil {
		t.Fatal("discovered Targets with no identity setting were accepted")
	}
	settings[0].Identity = true
	if err := ValidateDiscover(settings, d); err != nil {
		t.Fatalf("valid: %v", err)
	}
	d.DiscoveredTargets[0].Settings["ssl_certificate"] = ""
	if err := ValidateDiscover(settings, d); err == nil {
		t.Fatal("an empty identity value was accepted")
	}
	d.DiscoveredTargets[0].Settings = map[string]string{"ssl_certificate": "/a.crt", "bogus": "x"}
	if err := ValidateDiscover(settings, d); err == nil {
		t.Fatal("an undeclared setting key was accepted")
	}
}

func TestValidateDiscoverChecksReferencesAndWarnings(t *testing.T) {
	settings := []SettingDeclaration{{Key: "ssl_certificate", Identity: true}}
	target := func() DiscoveredTarget {
		return DiscoveredTarget{
			Settings:   map[string]string{"ssl_certificate": "/a.crt"},
			References: []Reference{{Setting: "ssl_certificate", File: "/etc/nginx/nginx.conf", Line: 3, Value: "a.crt"}},
		}
	}
	d := DiscoverResponse{DiscoveredTargets: []DiscoveredTarget{target()}}
	if err := ValidateDiscover(settings, d); err != nil {
		t.Fatalf("valid: %v", err)
	}
	for name, mutate := range map[string]func(*DiscoveredTarget){
		"undeclared setting": func(x *DiscoveredTarget) { x.References[0].Setting = "bogus" },
		"no file":            func(x *DiscoveredTarget) { x.References[0].File = "" },
		"line 0":             func(x *DiscoveredTarget) { x.References[0].Line = 0 },
		"no value":           func(x *DiscoveredTarget) { x.References[0].Value = "" },
		"warning without English": func(x *DiscoveredTarget) {
			x.Warnings = []l10n.Text{{"es": "solo español"}}
		},
		"too many references": func(x *DiscoveredTarget) {
			for range MaxReferences {
				x.References = append(x.References, x.References[0])
			}
		},
	} {
		x := target()
		mutate(&x)
		if err := ValidateDiscover(settings, DiscoverResponse{DiscoveredTargets: []DiscoveredTarget{x}}); err == nil {
			t.Errorf("%s was accepted", name)
		}
	}
}
