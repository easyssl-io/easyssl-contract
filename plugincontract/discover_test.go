package plugincontract

import "testing"

func TestValidateDiscoverRequiresIdentityForDiscoveredTargets(t *testing.T) {
	settings := []SettingDeclaration{{Key: "ssl_certificate"}, {Key: "reload"}}
	d := DiscoverResponse{}
	if err := ValidateDiscover(settings, d); err != nil {
		t.Fatalf("no discovered Targets, no identity: %v", err)
	}
	d.DiscoveredTargets = []DiscoveredTarget{{PluginSettings: map[string]string{"ssl_certificate": "/a.crt"}}}
	if err := ValidateDiscover(settings, d); err == nil {
		t.Fatal("discovered Targets with no identity setting were accepted")
	}
	settings[0].Identity = true
	if err := ValidateDiscover(settings, d); err != nil {
		t.Fatalf("valid: %v", err)
	}
	d.DiscoveredTargets[0].PluginSettings["ssl_certificate"] = ""
	if err := ValidateDiscover(settings, d); err == nil {
		t.Fatal("an empty identity value was accepted")
	}
	d.DiscoveredTargets[0].PluginSettings = map[string]string{"ssl_certificate": "/a.crt", "bogus": "x"}
	if err := ValidateDiscover(settings, d); err == nil {
		t.Fatal("an undeclared setting key was accepted")
	}
}
