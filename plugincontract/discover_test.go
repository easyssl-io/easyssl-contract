package plugincontract

import "testing"

func TestValidateDiscoverRequiresIdentityForSuggestions(t *testing.T) {
	d := DiscoverResponse{Settings: []SettingDeclaration{{Key: "ssl_certificate"}, {Key: "reload"}}}
	if err := ValidateDiscover(d); err != nil {
		t.Fatalf("no suggestions, no identity: %v", err)
	}
	d.SuggestedTargets = []SuggestedTarget{{PluginSettings: map[string]string{"ssl_certificate": "/a.crt"}}}
	if err := ValidateDiscover(d); err == nil {
		t.Fatal("suggestions with no identity setting were accepted")
	}
	d.Settings[0].Identity = true
	if err := ValidateDiscover(d); err != nil {
		t.Fatalf("valid: %v", err)
	}
	d.SuggestedTargets[0].PluginSettings["ssl_certificate"] = ""
	if err := ValidateDiscover(d); err == nil {
		t.Fatal("an empty identity value was accepted")
	}
	d.SuggestedTargets[0].PluginSettings = map[string]string{"ssl_certificate": "/a.crt", "bogus": "x"}
	if err := ValidateDiscover(d); err == nil {
		t.Fatal("an undeclared setting key was accepted")
	}
}
