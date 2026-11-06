package plugincontract

import (
	"strings"
	"testing"

	"github.com/easyssl-io/easyssl-contract/l10n"
)

// The bounds are the contract's, so the Agent refuses a Plugin past one on the
// host instead of the Server refusing the whole host report.
func TestAnswersPastTheBoundsAreRefused(t *testing.T) {
	text := l10n.Text{l10n.English: "x"}
	ok := DescribeResponse{Version: "1.0.0", Description: text}
	if err := ValidateDescribe(ok); err != nil {
		t.Fatalf("a valid answer was refused: %v", err)
	}
	for name, d := range map[string]DescribeResponse{
		"no version":        {Description: text},
		"long version":      {Version: strings.Repeat("1", MaxVersionBytes+1), Description: text},
		"too many settings": {Version: "1.0.0", Description: text, Declarations: make([]SettingDeclaration, MaxSettings+1)},
	} {
		if ValidateDescribe(d) == nil {
			t.Errorf("%s was accepted", name)
		}
	}
	if ValidateDiscover(nil, DiscoverResponse{DiscoveredTargets: make([]DiscoveredTarget, MaxDiscoveredTargets+1)}) == nil {
		t.Error("too many discovered Targets were accepted")
	}
}
