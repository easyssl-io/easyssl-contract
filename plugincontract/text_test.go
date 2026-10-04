package plugincontract

import (
	"testing"

	"github.com/easyssl-io/easyssl-contract/l10n"
)

// A text nobody could show is refused wherever it sits in the answer: one
// without English cannot be read by a reader of any other language.
func TestValidateDiscoverRefusesTextWithoutEnglish(t *testing.T) {
	spanishOnly := l10n.Text{"es": "solo español"}
	good := l10n.Text{"en": "English", "es": "Español"}
	setting := SettingDeclaration{Key: "path", Label: good, Description: good, Identity: true}
	target := DiscoveredTarget{PluginSettings: map[string]string{"path": "/x"}}
	cases := map[string]DiscoverResponse{
		"description":         {Description: spanishOnly},
		"warning":             {Warnings: []l10n.Text{good, spanishOnly}},
		"label":               {Settings: []SettingDeclaration{{Key: "k", Label: spanishOnly}}},
		"setting description": {Settings: []SettingDeclaration{{Key: "k", Label: good, Description: spanishOnly}}},
		"extra-data label": {Settings: []SettingDeclaration{setting}, DiscoveredTargets: []DiscoveredTarget{{
			PluginSettings: target.PluginSettings, PluginExtraData: []ExtraDatum{{Label: spanishOnly, Value: "v"}},
		}}},
	}
	for name, d := range cases {
		if ValidateDiscover(d) == nil {
			t.Errorf("%s without English was accepted", name)
		}
	}
	ok := DiscoverResponse{Description: good, Warnings: []l10n.Text{good}, Settings: []SettingDeclaration{setting},
		DiscoveredTargets: []DiscoveredTarget{{PluginSettings: target.PluginSettings, PluginExtraData: []ExtraDatum{{Label: good, Value: "v"}}}}}
	if err := ValidateDiscover(ok); err != nil {
		t.Errorf("a valid answer was refused: %v", err)
	}
}
