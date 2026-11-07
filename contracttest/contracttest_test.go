package contracttest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type message struct {
	Name  string   `json:"name"`
	Items []string `json:"items"`
}

type renamed struct {
	Title string   `json:"title"`
	Items []string `json:"items"`
}

type writeOnly struct {
	Name string `json:"name"`
}

func (w writeOnly) MarshalJSON() ([]byte, error) {
	return []byte(`{"name":"` + w.Name + `","extra":1}`), nil
}

func TestCheck(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "message.json")
	if err := write(path, message{Name: "a", Items: []string{}}); err != nil {
		t.Fatal(err)
	}
	if err := check(path, message{Name: "a", Items: []string{}}); err != nil {
		t.Fatalf("the value it was written from: %v", err)
	}
	if err := check(path, message{Name: "b", Items: []string{}}); err == nil || !strings.Contains(err.Error(), "no longer encodes") {
		t.Fatalf("a changed value = %v, want a mismatch", err)
	}
	// A renamed tag: the old file has a field the new type does not know.
	if err := check(path, renamed{Title: "a", Items: []string{}}); err == nil {
		t.Fatal("a renamed tag passed")
	}
	if err := decodeFile(path, &renamed{}); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("decoding the old file into the renamed type = %v, want an unknown field", err)
	}
	// A field that encodes but does not decode.
	wo := filepath.Join(dir, "write-only.json")
	if err := write(wo, writeOnly{Name: "a"}); err != nil {
		t.Fatal(err)
	}
	if err := check(wo, writeOnly{Name: "a"}); err == nil {
		t.Fatal("a field that does not decode passed")
	}
	if err := check(filepath.Join(dir, "missing.json"), message{}); err == nil || !strings.Contains(err.Error(), "-update") {
		t.Fatalf("a missing file = %v, want the -update hint", err)
	}
}

func TestDecodeFileRefusesTrailingData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "two.json")
	if err := os.WriteFile(path, []byte(`{"name":"a","items":[]} {}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := decodeFile(path, &message{}); err == nil {
		t.Fatal("two documents decoded")
	}
}

// A released copy keeps decoding into the current type, and one whose message
// was removed fails.
func TestCheckReleased(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "v0.1.0")
	kept, gone := filepath.Join(dir, "message.json"), filepath.Join(dir, "gone.json")
	for _, p := range []string{kept, gone} {
		if err := write(p, message{Name: "a", Items: []string{}}); err != nil {
			t.Fatal(err)
		}
	}
	messages := map[string]any{"message": message{}}
	if err := checkReleased(kept, messages); err != nil {
		t.Fatalf("a kept message = %v", err)
	}
	if err := checkReleased(gone, messages); err == nil || !strings.Contains(err.Error(), "no longer exists") {
		t.Fatalf("a removed message = %v", err)
	}
	if err := checkReleased(kept, map[string]any{"message": renamed{}}); err == nil {
		t.Fatal("a released message whose tag was renamed still decoded")
	}
}

func TestCheckBytes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "unit.service")
	if err := writeBytes(path, []byte("[Unit]\n")); err != nil {
		t.Fatal(err)
	}
	if err := checkBytes(path, []byte("[Unit]\n")); err != nil {
		t.Fatalf("the same bytes = %v", err)
	}
	if err := checkBytes(path, []byte("[Unit]\n\n")); err == nil || !strings.Contains(err.Error(), "changed") {
		t.Fatalf("one more byte = %v, want a mismatch", err)
	}
}

type inner struct {
	Note string `json:"note,omitempty"`
}

type embedded struct {
	Extra string `json:"extra,omitempty"`
}

type outer struct {
	embedded
	Name     string            `json:"name"`
	Optional string            `json:"optional,omitempty"`
	Zeroable int               `json:"zeroable,omitzero"`
	Items    []inner           `json:"items"`
	Ptr      *inner            `json:"ptr,omitempty"`
	ByKey    map[string]inner  `json:"byKey,omitempty"`
	Raw      json.RawMessage   `json:"raw,omitempty"`
	Skipped  string            `json:"-"`
	Text     map[string]string `json:"text,omitempty"`
}

// An optional field counts as set when any example sets it, at any depth;
// one no example sets is reported, a required field never is.
func TestOptionalUnset(t *testing.T) {
	unset := optionalUnset(map[string]any{
		"a": outer{Name: "x", Optional: "y", Items: []inner{{}}},
		"b": outer{Zeroable: 1, Items: []inner{{Note: "n"}}},
	})
	want := []string{
		"contracttest.embedded.extra", "contracttest.outer.byKey", "contracttest.outer.ptr",
		"contracttest.outer.raw", "contracttest.outer.text",
	}
	if strings.Join(unset, " ") != strings.Join(want, " ") {
		t.Fatalf("unset %v, want %v", unset, want)
	}

	if unset := optionalUnset(map[string]any{"full": full()}); len(unset) != 0 {
		t.Fatalf("a value setting every optional field reported %v", unset)
	}
}

// full sets every optional field of outer.
func full() outer {
	return outer{
		embedded: embedded{Extra: "e"}, Optional: "y", Zeroable: 1, Items: []inner{{Note: "n"}},
		Ptr: &inner{Note: "p"}, ByKey: map[string]inner{"k": {Note: "m"}}, Raw: json.RawMessage(`{}`),
		Text: map[string]string{"en": "t"},
	}
}

type stamped struct {
	At time.Time `json:"at,omitzero"`
}

// A value encoding/json omits is not set, though reflection calls it nonzero:
// an empty map, slice or string under omitempty, and under omitzero a value
// whose type's own IsZero says zero.
func TestOptionalUnsetFollowsTheEncoder(t *testing.T) {
	empty := full()
	empty.ByKey, empty.Text, empty.Raw = map[string]inner{}, map[string]string{}, json.RawMessage{}
	if got := strings.Join(optionalUnset(map[string]any{"empty": empty}), " "); got != "contracttest.outer.byKey contracttest.outer.raw contracttest.outer.text" {
		t.Fatalf("empty containers left unset %q, want byKey, raw and text", got)
	}
	if got := optionalUnset(map[string]any{"zero": stamped{At: time.Time{}.In(time.FixedZone("x", 3600))}}); len(got) != 1 {
		t.Fatalf("a time its own IsZero calls zero counted as set: %v", got)
	}
	if got := optionalUnset(map[string]any{"set": stamped{At: time.Unix(1, 0)}}); len(got) != 0 {
		t.Fatalf("a set time reported %v", got)
	}
}
