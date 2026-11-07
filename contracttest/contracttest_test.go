package contracttest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
