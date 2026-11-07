// Package contracttest pins each message of a contract to an example file, so
// a change to a message's JSON shows as a diff in review, and checks that the
// examples kept from every released version still decode, which is what makes
// a change additive rather than breaking.
//
// A package's examples live in `testdata/<message>.json`; a release copies them
// to `testdata/v<version>/`, and those copies never change. After an intended
// change, `go test ./<package>/ -update` rewrites the current ones; name the
// packages, since a test binary that does not import this package does not
// know the flag.
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
// `<dir>/<name>.json` with Golden, every optional field must be set by at least
// one value (optionalUnset), and every file in each `<dir>/v*/` copy is
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
	t.Run("every optional field", func(t *testing.T) {
		if unset := optionalUnset(messages); len(unset) > 0 {
			t.Fatalf("no example sets %s; give one example a value for each", strings.Join(unset, ", "))
		}
	})
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

// optionalUnset lists the optional fields (`omitempty` or `omitzero`) that no
// example sets, as `<type>.<json name>`. An example file holds only what its
// value encodes, so such a field is in no file, and renaming or dropping it
// would pass every check, the copies kept from released versions included.
//
// Every optional field reachable from a message's type counts, whether or not
// an example reaches it: one behind an empty list or a nil pointer is in no
// file either.
func optionalUnset(messages map[string]any) []string {
	set := map[string]bool{}
	seen := map[reflect.Type]bool{}
	for _, v := range messages {
		declare(reflect.TypeOf(v), set, seen)
		walk(reflect.ValueOf(v), set)
	}
	var unset []string
	for field, ok := range set {
		if !ok {
			unset = append(unset, field)
		}
	}
	sort.Strings(unset)
	return unset
}

var marshaler = reflect.TypeFor[json.Marshaler]()

// declare records every optional field reachable from t as unset, unless an
// example already set it.
func declare(t reflect.Type, set map[string]bool, seen map[reflect.Type]bool) {
	for {
		switch t.Kind() {
		case reflect.Pointer, reflect.Map:
			t = t.Elem()
			continue
		case reflect.Slice, reflect.Array:
			if t.Elem().Kind() == reflect.Uint8 {
				return
			}
			t = t.Elem()
			continue
		}
		break
	}
	if t.Kind() != reflect.Struct || seen[t] || encodesItself(t) {
		return
	}
	seen[t] = true
	for _, f := range jsonFields(t) {
		if f.optional {
			if key := t.String() + "." + f.name; !set[key] {
				set[key] = false
			}
		}
		declare(t.Field(f.index).Type, set, seen)
	}
}

// walk records, for every optional field reachable from v, whether v sets it.
// A type that encodes itself (a time, raw JSON) is a leaf.
func walk(v reflect.Value, set map[string]bool) {
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if !v.IsNil() {
			walk(v.Elem(), set)
		}
	case reflect.Slice, reflect.Array:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return
		}
		for i := range v.Len() {
			walk(v.Index(i), set)
		}
	case reflect.Map:
		for _, k := range v.MapKeys() {
			walk(v.MapIndex(k), set)
		}
	case reflect.Struct:
		t := v.Type()
		if encodesItself(t) {
			return
		}
		for _, f := range jsonFields(t) {
			if f.optional {
				key := t.String() + "." + f.name
				set[key] = set[key] || !omitted(v.Field(f.index), f.opts)
			}
			walk(v.Field(f.index), set)
		}
	}
}

func encodesItself(t reflect.Type) bool {
	return t.Implements(marshaler) || reflect.PointerTo(t).Implements(marshaler)
}

// jsonField is one field of a struct as encoding/json sees it. An untagged
// embedded struct is listed with no name: the encoder promotes its fields,
// even when the struct's own type is unexported.
type jsonField struct {
	index      int
	name, opts string
	optional   bool
}

func jsonFields(t reflect.Type) []jsonField {
	var out []jsonField
	for i := range t.NumField() {
		f := t.Field(i)
		tag := f.Tag.Get("json")
		name, opts, _ := strings.Cut(tag, ",")
		if f.Anonymous && name == "" && tag != "-" {
			out = append(out, jsonField{index: i})
			continue
		}
		if !f.IsExported() || tag == "-" {
			continue
		}
		if name == "" {
			name = f.Name
		}
		out = append(out, jsonField{index: i, name: name, opts: opts,
			optional: hasOption(opts, "omitempty") || hasOption(opts, "omitzero")})
	}
	return out
}

// omitted says whether encoding/json leaves a field with these options out:
// `omitempty` drops false, 0, a nil pointer or interface, and an empty string,
// slice, array or map, so an example that sets an optional map to `{}` has not
// set it; `omitzero` drops the zero value, or what the type's own IsZero calls
// zero.
func omitted(v reflect.Value, opts string) bool {
	if hasOption(opts, "omitzero") {
		if v.Kind() == reflect.Pointer && v.IsNil() {
			return true
		}
		if z, ok := isZeroer(v); ok {
			if z {
				return true
			}
		} else if v.IsZero() {
			return true
		}
	}
	if hasOption(opts, "omitempty") {
		switch v.Kind() {
		case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
			return v.Len() == 0
		case reflect.Struct:
			return false
		default:
			return v.IsZero()
		}
	}
	return false
}

// isZeroer calls v's own IsZero method when it has one.
func isZeroer(v reflect.Value) (zero, ok bool) {
	if !v.CanInterface() {
		return false, false
	}
	z, ok := v.Interface().(interface{ IsZero() bool })
	if !ok {
		return false, false
	}
	return z.IsZero(), true
}

func hasOption(opts, want string) bool {
	for _, o := range strings.Split(opts, ",") {
		if o == want {
			return true
		}
	}
	return false
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

// GoldenFile checks that got is exactly the file at path, for a contract that
// is not JSON: a generated script or unit file. With -update it writes the
// file instead.
func GoldenFile(t testing.TB, path string, got []byte) {
	t.Helper()
	if *update {
		if err := writeBytes(path, got); err != nil {
			t.Fatal(err)
		}
		return
	}
	if err := checkBytes(path, got); err != nil {
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
	return writeBytes(path, b)
}

func writeBytes(path string, b []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func checkBytes(path string, got []byte) error {
	want, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("%w; run `go test -update` to create it", err)
	}
	if !bytes.Equal(got, want) {
		return fmt.Errorf("%s changed; if the change is intended, run `go test -update` in its package and review the diff\ngot:\n%s", path, got)
	}
	return nil
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
		return fmt.Errorf("%T no longer encodes as %s; if the change is intended, run `go test -update` in its package and review the diff\ngot:\n%s", v, path, got)
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
