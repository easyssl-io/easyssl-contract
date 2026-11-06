// Package l10ncheck keeps a component's translations complete: it compares the
// message calls in a module's Go source with its catalogs, so a message added,
// changed or removed without its translations fails a test. It is called from
// one test per module:
//
//	func TestCatalog(t *testing.T) { l10ncheck.Check(t, ".", "locales") }
package l10ncheck

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/easyssl-io/easyssl-contract/l10n"
)

type message struct {
	id         string
	one, other string // one is empty for a plain message
	plural     bool
	pos        string
}

// Check fails t once per problem Problems finds.
func Check(t testing.TB, srcDir, localesDir string) {
	t.Helper()
	problems, err := Problems(srcDir, localesDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range problems {
		t.Error(p)
	}
}

// Problems walks every Go file under srcDir but tests and `testdata`, takes
// each call to a function or method named T, N or Text whose first argument
// is a string literal as a message, and compares the messages with
// `<lang>.json` in localesDir for every supported language but English. It
// reports:
//
//   - a message with no translation;
//   - a translation whose ID no message uses;
//   - one ID carrying two different English texts;
//   - a translation whose placeholders differ from the English ones, or whose
//     shape does (plural forms for a plain message, or the reverse);
//   - a message whose English is not a string literal, since nothing could
//     then be compared.
func Problems(srcDir, localesDir string) ([]string, error) {
	messages, problems, err := scan(srcDir)
	if err != nil {
		return nil, err
	}
	for _, lang := range l10n.Supported() {
		if lang == l10n.English {
			continue
		}
		name := filepath.Join(localesDir, lang+".json")
		raw, err := os.ReadFile(name)
		if err != nil {
			return nil, fmt.Errorf("l10ncheck: %w", err)
		}
		var entries map[string]json.RawMessage
		if err := json.Unmarshal(raw, &entries); err != nil {
			return nil, fmt.Errorf("l10ncheck: %s: %w", name, err)
		}
		for _, id := range sortedKeys(messages) {
			m := messages[id]
			raw, ok := entries[id]
			if !ok {
				problems = append(problems, fmt.Sprintf("%s: %q has no %s translation (%s)", m.pos, id, lang, name))
				continue
			}
			if p := compare(m, raw); p != "" {
				problems = append(problems, fmt.Sprintf("%s: %q: %s", name, id, p))
			}
		}
		for _, id := range sortedKeys(entries) {
			if _, ok := messages[id]; !ok {
				problems = append(problems, fmt.Sprintf("%s: %q is translated but no message uses it", name, id))
			}
		}
	}
	return problems, nil
}

func compare(m message, raw json.RawMessage) string {
	var plain string
	if json.Unmarshal(raw, &plain) == nil {
		if m.plural {
			return "the message depends on a count, so its translation needs \"one\" and \"other\""
		}
		if plain == "" {
			return "the translation is empty"
		}
		return placeholders(m.other, plain, "")
	}
	var forms struct{ One, Other string }
	if err := json.Unmarshal(raw, &forms); err != nil {
		return "a translation is a string, or an object with \"one\" and \"other\""
	}
	if !m.plural {
		return "the message does not depend on a count, so its translation is a string"
	}
	if forms.One == "" || forms.Other == "" {
		return "both \"one\" and \"other\" are needed"
	}
	if p := placeholders(m.one, forms.One, "one: "); p != "" {
		return p
	}
	return placeholders(m.other, forms.Other, "other: ")
}

// placeholders compares the names two texts use, and the code spans: a span is
// a literal the reader types or looks for, the same in every language, so a
// translation keeps each one exactly. `{n}` is left out: a form for exactly one
// may spell the number as a word.
func placeholders(english, translated, prefix string) string {
	a, b := names(english), names(translated)
	if a != b {
		return fmt.Sprintf("%splaceholders differ: English has {%s}, the translation has {%s}", prefix, a, b)
	}
	if a, b := codeSpans(english), codeSpans(translated); a != b {
		return fmt.Sprintf("%scode spans differ: English has %s, the translation has %s", prefix, a, b)
	}
	return ""
}

// codeSpans is every backtick span of a text, sorted, as one comparable string.
func codeSpans(text string) string {
	parts := strings.Split(text, "`")
	var spans []string
	for i := 1; i < len(parts); i += 2 {
		spans = append(spans, "`"+parts[i]+"`")
	}
	sort.Strings(spans)
	return "[" + strings.Join(spans, " ") + "]"
}

func names(text string) string {
	var out []string
	for _, p := range l10n.Placeholders(text) {
		if p != "n" {
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return strings.Join(out, ", ")
}

func scan(srcDir string) (map[string]message, []string, error) {
	messages := map[string]message{}
	var problems []string
	fset := token.NewFileSet()
	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "testdata" || (strings.HasPrefix(d.Name(), ".") && path != srcDir) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) < 2 {
				return true
			}
			name := callee(call.Fun)
			if name != "T" && name != "N" && name != "Text" {
				return true
			}
			id, ok := literal(call.Args[0])
			if !ok {
				return true
			}
			m := message{id: id, pos: fset.Position(call.Pos()).String()}
			if name == "N" {
				if len(call.Args) < 4 {
					return true
				}
				one, ok1 := literal(call.Args[2])
				other, ok2 := literal(call.Args[3])
				if !ok1 || !ok2 {
					problems = append(problems, fmt.Sprintf("%s: %q: the English forms must be string literals", m.pos, id))
					return true
				}
				m.one, m.other, m.plural = one, other, true
			} else {
				english, ok := literal(call.Args[1])
				if !ok {
					problems = append(problems, fmt.Sprintf("%s: %q: the English text must be a string literal", m.pos, id))
					return true
				}
				m.other = english
			}
			if prev, seen := messages[id]; seen && (prev.one != m.one || prev.other != m.other || prev.plural != m.plural) {
				problems = append(problems, fmt.Sprintf("%s: %q carries a different English text than at %s", m.pos, id, prev.pos))
				return true
			}
			messages[id] = m
			return true
		})
		return nil
	})
	return messages, problems, err
}

func callee(fun ast.Expr) string {
	switch f := fun.(type) {
	case *ast.Ident:
		return f.Name
	case *ast.SelectorExpr:
		return f.Sel.Name
	}
	return ""
}

// literal folds a string literal, or literals joined by +, into its value.
func literal(e ast.Expr) (string, bool) {
	switch v := e.(type) {
	case *ast.BasicLit:
		if v.Kind != token.STRING {
			return "", false
		}
		s, err := strconv.Unquote(v.Value)
		return s, err == nil
	case *ast.BinaryExpr:
		if v.Op != token.ADD {
			return "", false
		}
		a, ok1 := literal(v.X)
		b, ok2 := literal(v.Y)
		return a + b, ok1 && ok2
	case *ast.ParenExpr:
		return literal(v.X)
	}
	return "", false
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
