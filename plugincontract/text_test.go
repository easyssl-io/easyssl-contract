package plugincontract

import (
	"testing"

	"github.com/easyssl-io/easyssl-contract/l10n"
)

// A text nobody could show is refused wherever it sits in either answer: one
// without English cannot be read by a reader of any other language.
func TestValidateRefusesTextWithoutEnglish(t *testing.T) {
	spanishOnly := l10n.Text{"es": "solo español"}
	good := l10n.Text{"en": "English", "es": "Español"}
	setting := SettingDeclaration{Key: "path", Label: good, Description: good, Identity: true}
	target := DiscoveredTarget{PluginSettings: map[string]string{"path": "/x"}}

	describes := map[string]DescribeResponse{
		"description":         {Version: "1.0.0", Description: spanishOnly},
		"label":               {Version: "1.0.0", Settings: []SettingDeclaration{{Key: "k", Label: spanishOnly}}},
		"setting description": {Version: "1.0.0", Settings: []SettingDeclaration{{Key: "k", Label: good, Description: spanishOnly}}},
	}
	for name, d := range describes {
		if ValidateDescribe(d) == nil {
			t.Errorf("%s without English was accepted", name)
		}
	}
	discovers := map[string]DiscoverResponse{
		"warning": {Warnings: []l10n.Text{good, spanishOnly}},
		"extra-data label": {DiscoveredTargets: []DiscoveredTarget{{
			PluginSettings: target.PluginSettings, PluginExtraData: []ExtraDatum{{Label: spanishOnly, Value: "v"}},
		}}},
	}
	for name, d := range discovers {
		if ValidateDiscover([]SettingDeclaration{setting}, d) == nil {
			t.Errorf("%s without English was accepted", name)
		}
	}

	if err := ValidateDescribe(DescribeResponse{Version: "1.0.0", Description: good, Settings: []SettingDeclaration{setting}}); err != nil {
		t.Errorf("a valid describe answer was refused: %v", err)
	}
	ok := DiscoverResponse{Warnings: []l10n.Text{good},
		DiscoveredTargets: []DiscoveredTarget{{PluginSettings: target.PluginSettings, PluginExtraData: []ExtraDatum{{Label: good, Value: "v"}}}}}
	if err := ValidateDiscover([]SettingDeclaration{setting}, ok); err != nil {
		t.Errorf("a valid discover answer was refused: %v", err)
	}
}
