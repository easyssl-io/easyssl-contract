package plugincontract_test

import (
	"errors"
	"testing"

	"github.com/easyssl-io/easyssl-contract/l10n"
	"github.com/easyssl-io/easyssl-contract/plugincontract"
)

func decl(key string, summary bool) plugincontract.SettingDeclaration {
	return plugincontract.SettingDeclaration{Key: key, Type: plugincontract.SettingTypeText, Label: l10n.Text{"en": key}, Summary: summary}
}

// Each declaration a reader could not store or show against a key is refused,
// and a type the reader does not know is not: it is shown as text.
func TestValidateSettingsRefusesUnusableDeclarations(t *testing.T) {
	cases := map[string]plugincontract.SettingDeclaration{
		"an empty key":            {Key: "", Type: "text", Label: l10n.Text{"en": "x"}},
		"a key with a hyphen":     {Key: "ssl-cert", Type: "text", Label: l10n.Text{"en": "x"}},
		"a key starting with 1":   {Key: "1st", Type: "text", Label: l10n.Text{"en": "x"}},
		"no type":                 {Key: "k", Label: l10n.Text{"en": "x"}},
		"no label":                {Key: "k", Type: "text"},
		"a label without English": {Key: "k", Type: "text", Label: l10n.Text{"es": "x"}},
	}
	for name, d := range cases {
		if plugincontract.ValidateSettings([]plugincontract.SettingDeclaration{d}) == nil {
			t.Errorf("%s was accepted", name)
		}
	}
	if plugincontract.ValidateSettings([]plugincontract.SettingDeclaration{decl("k", false), decl("k", false)}) == nil {
		t.Error("a key declared twice was accepted")
	}
	future := decl("SSLCertificateFile", false)
	future.Type = "secret"
	if err := plugincontract.ValidateSettings([]plugincontract.SettingDeclaration{future}); err != nil {
		t.Errorf("a type this build does not know was refused: %v", err)
	}
}

func TestCodeLeavesTextWithABacktickOutOfASpan(t *testing.T) {
	for in, want := range map[string]string{"/etc/nginx": "`/etc/nginx`", "a`b": "a`b", "": ""} {
		if got := plugincontract.Code(in); got != want {
			t.Errorf("Code(%q) = %q, want %q", in, got, want)
		}
	}
}

// Two summaries parse and every field is valid, so nothing else would catch it.
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
