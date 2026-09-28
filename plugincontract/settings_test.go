package plugincontract_test

import (
	"errors"
	"testing"

	"github.com/easyssl-io/easyssl-contract/plugincontract"
)

func decl(key string, summary bool) plugincontract.SettingDeclaration {
	return plugincontract.SettingDeclaration{Key: key, Summary: summary}
}

// The case the rule exists for. It is not a shape problem — every field is
// valid and it parses — so nothing else would catch it, and what a reader
// would see is an unlabeled Target named after whichever setting the Plugin's
// author happened to list first.
func TestValidateSettingsRefusesTwoSummaries(t *testing.T) {
	err := plugincontract.ValidateSettings([]plugincontract.SettingDeclaration{
		decl("server_name", true), decl("reload", false), decl("site", true),
	})
	var tooMany *plugincontract.TooManySummaryError
	if !errors.As(err, &tooMany) {
		t.Fatalf("two summary settings were accepted: %v", err)
	}
	// Both keys, because the Plugin's author is the one who has to choose and
	// the keys are how they find them.
	if len(tooMany.Keys) != 2 || tooMany.Keys[0] != "server_name" || tooMany.Keys[1] != "site" {
		t.Fatalf("the error does not name both offenders: %v", tooMany.Keys)
	}
}

func TestValidateSettingsAcceptsOneOrNone(t *testing.T) {
	for name, decls := range map[string][]plugincontract.SettingDeclaration{
		"none":                    {decl("reload", false), decl("path", false)},
		"exactly one":             {decl("reload", false), decl("server_name", true)},
		"no settings at all":      nil,
		"one setting, summarized": {decl("site", true)},
	} {
		if err := plugincontract.ValidateSettings(decls); err != nil {
			t.Errorf("%s was refused: %v", name, err)
		}
	}
}

// SummarySetting is only meaningful once the rule above holds, which is what
// makes "the summary setting" one thing rather than "the first of them".
func TestSummarySetting(t *testing.T) {
	got := plugincontract.SummarySetting([]plugincontract.SettingDeclaration{
		decl("reload", false), decl("server_name", true),
	})
	if got != "server_name" {
		t.Fatalf("SummarySetting = %q, want server_name", got)
	}
	if got := plugincontract.SummarySetting([]plugincontract.SettingDeclaration{decl("reload", false)}); got != "" {
		t.Fatalf("a Plugin declaring no summary got %q", got)
	}
}
