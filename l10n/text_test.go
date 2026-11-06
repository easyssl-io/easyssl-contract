package l10n

import (
	"encoding/json"
	"reflect"
	"testing"
)

// A reader keeps the languages it can read and drops the rest, rather than
// refusing a text a newer producer wrote.
func TestDecodingDropsWhatIsNotABaseLanguageCode(t *testing.T) {
	var got Text
	if err := json.Unmarshal([]byte(`{"en":"Reload","es":"Recargar","pt-BR":"Recarregar","pt":"Recarregar","":"x","EN":"y","zzzz":"z"}`), &got); err != nil {
		t.Fatal(err)
	}
	want := Text{"en": "Reload", "es": "Recargar", "pt": "Recarregar"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("decoded %v, want %v", got, want)
	}
	if err := got.Validate(); err != nil {
		t.Fatalf("a decoded text with a language this build does not speak is invalid: %v", err)
	}
	many := Text{"en": "x"}
	for _, c := range []string{"aa", "ab", "ac", "ad", "ae", "af", "ag", "ah", "ai", "aj", "ak", "al", "am", "an", "ao", "ap", "aq", "ar", "as", "at", "au"} {
		many[c] = "x"
	}
	if err := many.Validate(); err != nil {
		t.Fatalf("%d languages refused: %v", len(many), err)
	}
	var none Text
	if err := json.Unmarshal([]byte(`null`), &none); err != nil || none != nil {
		t.Fatalf("null decoded to %v, %v", none, err)
	}
}

// A value holding a backtick cannot sit in a code span, so a placeholder set
// as one loses its backticks for that value and keeps them for any other.
func TestFillLeavesAValueWithABacktickOutOfACodeSpan(t *testing.T) {
	for _, tc := range []struct{ value, want string }{
		{"/etc/nginx/site.conf", "reading `/etc/nginx/site.conf` failed"},
		{"/tmp/a`b", "reading /tmp/a`b failed"},
	} {
		if got := fill("reading `{path}` failed", []any{"path", tc.value}); got != tc.want {
			t.Errorf("fill with %q = %q, want %q", tc.value, got, tc.want)
		}
	}
}
