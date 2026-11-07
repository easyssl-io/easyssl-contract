// Package contracttest pins each message of a contract to an example file, so
// a change to a message's JSON shows as a diff in review, and checks that the
// examples kept from every released version still decode, which is what makes
// a change additive rather than breaking.
//
// A package's examples live in `testdata/<message>.json`; a release copies them
// to `testdata/v<version>/`, and those copies never change. After an intended
// change, `go test ./... -update` rewrites the current ones.
package contracttest

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite the contract's example files from the values the tests build")

// Messages pins every message of a package: each value is checked against
// `<dir>/<name>.json` with Golden, and every file in each `<dir>/v*/` copy is
// checked with Decodes against the type of the message it is named after. A
// released file whose message no longer exists fails, since removing a
// message breaks its readers.
func Messages(t *testing.T, dir string, messages map[string]any) {
	t.Helper()
	names := make([]string, 0, len(messages))
	for name := range messages {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		t.Run(name, func(t *testing.T) { Golden(t, filepath.Join(dir, name+".json"), messages[name]) })
	}
	released, _ := filepath.Glob(filepath.Join(dir, "v*", "*.json"))
	for _, path := range released {
		name := strings.TrimSuffix(filepath.Base(path), ".json")
		t.Run(filepath.Base(filepath.Dir(path))+"/"+name, func(t *testing.T) {
			if err := checkReleased(path, messages); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func checkReleased(path string, messages map[string]any) error {
	v, ok := messages[strings.TrimSuffix(filepath.Base(path), ".json")]
	if !ok {
		return fmt.Errorf("%s is a released message that no longer exists", path)
	}
	return decodeFile(path, reflect.New(reflect.TypeOf(v)).Interface())
}

// Golden checks that v encodes exactly as the file at path, and that the file
// decodes strictly into v's type and encodes the same again, so a field that
// encodes but does not decode fails too. With -update it writes the file
// instead.
func Golden(t testing.TB, path string, v any) {
	t.Helper()
	if *update {
		if err := write(path, v); err != nil {
			t.Fatal(err)
		}
		return
	}
	if err := check(path, v); err != nil {
		t.Fatal(err)
	}
}

// Decodes checks that the file at path decodes into T with no field T does not
// know and nothing after the document, and returns the value.
func Decodes[T any](t testing.TB, path string) T {
	t.Helper()
	var v T
	if err := decodeFile(path, &v); err != nil {
		t.Fatal(err)
	}
	return v
}

func write(path string, v any) error {
	b, err := encode(v)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func check(path string, v any) error {
	got, err := encode(v)
	if err != nil {
		return err
	}
	want, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("%w; run `go test -update` to create it", err)
	}
	if !bytes.Equal(got, want) {
		return fmt.Errorf("%T no longer encodes as %s; if the change is intended, run `go test -update` and review the diff\ngot:\n%s", v, path, got)
	}
	back := reflect.New(reflect.TypeOf(v))
	if err := decodeFile(path, back.Interface()); err != nil {
		return err
	}
	again, err := encode(back.Elem().Interface())
	if err != nil {
		return err
	}
	if !bytes.Equal(again, want) {
		return fmt.Errorf("%s decodes into %T but encodes differently:\n%s", path, v, again)
	}
	return nil
}

func decodeFile(path string, ptr any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(ptr); err != nil {
		return fmt.Errorf("%s does not decode into %s: %w", path, reflect.TypeOf(ptr).Elem(), err)
	}
	if _, err := dec.Token(); err != io.EOF {
		return fmt.Errorf("%s holds more than one document", path)
	}
	return nil
}

func encode(v any) ([]byte, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	return append(b, '\n'), err
}
