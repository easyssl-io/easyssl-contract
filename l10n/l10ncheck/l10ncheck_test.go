package l10ncheck

import (
	"strings"
	"testing"
)

func TestProblems(t *testing.T) {
	if got, err := Problems("testdata/good", "testdata/good/locales"); err != nil || len(got) != 0 {
		t.Fatalf("good module: %v, %v", got, err)
	}
	got, err := Problems("testdata/bad", "testdata/bad/locales")
	if err != nil {
		t.Fatal(err)
	}
	all := strings.Join(got, "\n")
	for _, want := range []string{
		`"b.missing" has no es translation`,
		`"b.twice" carries a different English text`,
		`"b.placeholder": placeholders differ`,
		`"b.span": code spans differ`,
		`"b.shape": the message does not depend on a count`,
		`"b.count": the message depends on a count`,
		`"b.variable": the English text must be a string literal`,
		`"b.stale" is translated but no message uses it`,
	} {
		if !strings.Contains(all, want) {
			t.Errorf("missing problem %q in:\n%s", want, all)
		}
	}
	if len(got) != 8 {
		t.Errorf("%d problems, want 8:\n%s", len(got), all)
	}
}
