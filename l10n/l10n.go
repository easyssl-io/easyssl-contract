// Package l10n is how every Go component of easySSL says something to a person
// in that person's language. A message is an explicit ID with its English text
// beside it at the call site:
//
//	p.T("register.done", "{name} is authorized and active.", "name", name)
//
// The ID is never derived from the English, values are named placeholders a
// translation may reorder, and translated fragments are never concatenated.
// Translations live in one `<lang>.json` per language, mapping an ID to its
// text or to its plural forms. English has no file, and a message with no
// translation is shown in English, never as its ID.
//
// Standard library only, deliberately: a translation library would add itself
// and its dependencies to the Agent and to every Plugin, which are meant to be
// read end to end by whoever audits them.
package l10n

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"strings"
)

// English is the source language and the fallback.
const English = "en"

// MaxLanguages bounds the languages one Text may carry.
const MaxLanguages = 20

// supported is every language easySSL speaks, the source language first. The
// web console holds the same list; if either moves, the other moves.
var supported = []string{English, "es"}

// names is each language under its own name, which is how a language is
// offered to someone who may not read the one currently on screen.
var names = map[string]string{English: "English", "es": "Español"}

// Supported returns the supported languages, the source language first.
func Supported() []string { return append([]string(nil), supported...) }

// Name returns a language's own name for itself, or the code for a language
// that is not supported.
func Name(lang string) string {
	if n, ok := names[lang]; ok {
		return n
	}
	return lang
}

// Base resolves a language tag or a POSIX locale to the supported language it
// belongs to: `es-PE`, `es_419` and `es_PE.UTF-8` are all `es`. It reports
// false for anything else, `C` and `POSIX` included.
func Base(tag string) (string, bool) {
	tag = strings.ToLower(strings.TrimSpace(tag))
	if i := strings.IndexAny(tag, "-_.@"); i >= 0 {
		tag = tag[:i]
	}
	for _, l := range supported {
		if l == tag {
			return l, true
		}
	}
	return "", false
}

// FromLocale reads the language of the operating system's locale the way
// POSIX orders it: the first of LC_ALL, LC_MESSAGES and LANG that is set
// decides, so `LC_ALL=C` beside `LANG=es_PE.UTF-8` is not Spanish.
func FromLocale(getenv func(string) string) (string, bool) {
	for _, name := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if v := getenv(name); v != "" {
			return Base(v)
		}
	}
	return "", false
}

// Text is one text in several languages: a base language code to the text in
// that language. Empty means there is no text; one that is not empty holds
// English, the fallback for every other language. Every language travels
// because whoever reports a text cannot know who will read it.
type Text map[string]string

var languageCode = regexp.MustCompile(`^[a-z]{2,3}$`)

// In returns the text in lang, else in English, else "".
func (t Text) In(lang string) string {
	if s := t[lang]; s != "" {
		return s
	}
	return t[English]
}

// Validate reports a Text nobody could show: one holding text without
// English, a key that is not a base language code, or more languages than
// MaxLanguages.
func (t Text) Validate() error {
	if len(t) == 0 {
		return nil
	}
	if len(t) > MaxLanguages {
		return fmt.Errorf("l10n: a text carries %d languages, more than %d", len(t), MaxLanguages)
	}
	for lang := range t {
		if !languageCode.MatchString(lang) {
			return fmt.Errorf("l10n: %q is not a base language code", lang)
		}
	}
	if t[English] == "" {
		return errors.New("l10n: a text must hold English (\"en\")")
	}
	return nil
}

// entry is one translation: Other alone for a plain message, One and Other
// for a message that depends on a count.
type entry struct {
	One, Other string
	Plural     bool
}

func (e *entry) UnmarshalJSON(b []byte) error {
	var s string
	if json.Unmarshal(b, &s) == nil {
		e.Other = s
		return nil
	}
	var forms struct {
		One   string `json:"one"`
		Other string `json:"other"`
	}
	if err := json.Unmarshal(b, &forms); err != nil {
		return errors.New("a translation is a string, or an object with \"one\" and \"other\"")
	}
	e.One, e.Other, e.Plural = forms.One, forms.Other, true
	return nil
}

// Catalog holds a component's translations, one set per language.
type Catalog struct {
	translations map[string]map[string]entry
}

// Load reads `<lang>.json` from dir for every supported language but English.
// A language with no file is simply untranslated, which reads as English.
func Load(fsys fs.FS, dir string) (*Catalog, error) {
	c := &Catalog{translations: map[string]map[string]entry{}}
	for _, lang := range supported[1:] {
		name := path.Join(dir, lang+".json")
		raw, err := fs.ReadFile(fsys, name)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("l10n: %s: %w", name, err)
		}
		entries := map[string]entry{}
		if err := json.Unmarshal(raw, &entries); err != nil {
			return nil, fmt.Errorf("l10n: %s: %w", name, err)
		}
		c.translations[lang] = entries
	}
	return c, nil
}

// MustLoad is Load for a catalog embedded in the binary, where a failure is a
// build that should never have shipped.
func MustLoad(fsys fs.FS, dir string) *Catalog {
	c, err := Load(fsys, dir)
	if err != nil {
		panic(err)
	}
	return c
}

// In returns a Printer speaking lang, or English for a language that is not
// supported.
func (c *Catalog) In(lang string) Printer {
	if l, ok := Base(lang); ok {
		return Printer{c: c, lang: l}
	}
	return Printer{c: c, lang: English}
}

// Text renders one message in every supported language.
func (c *Catalog) Text(id, english string, args ...any) Text {
	t := Text{}
	for _, lang := range supported {
		t[lang] = c.In(lang).T(id, english, args...)
	}
	return t
}

// Printer renders a Catalog's messages in one language.
type Printer struct {
	c    *Catalog
	lang string
}

func (p Printer) Language() string { return p.lang }

// T renders the message id. english is its English text; args are
// name, value pairs filling its `{name}` placeholders.
func (p Printer) T(id, english string, args ...any) string {
	text := english
	if e, ok := p.lookup(id); ok && !e.Plural && e.Other != "" {
		text = e.Other
	}
	return fill(text, args)
}

// N renders a message whose wording depends on the count n, which also fills
// the `{n}` placeholder.
func (p Printer) N(id string, n int, one, other string, args ...any) string {
	if e, ok := p.lookup(id); ok && e.Plural && e.One != "" && e.Other != "" {
		one, other = e.One, e.Other
	}
	text := other
	if isOne(p.lang, n) {
		text = one
	}
	return fill(text, append([]any{"n", n}, args...))
}

func (p Printer) lookup(id string) (entry, bool) {
	if p.c == nil {
		return entry{}, false
	}
	e, ok := p.c.translations[p.lang][id]
	return e, ok
}

// isOne is the plural rule: which counts take the `one` form. English and
// Spanish agree. A language with more forms than two needs its forms added to
// entry and its rule here, and nothing changes where messages are used.
func isOne(_ string, n int) bool { return n == 1 }

var placeholder = regexp.MustCompile(`\{([a-zA-Z][a-zA-Z0-9]*)\}`)

// Placeholders returns the distinct placeholder names of a text, sorted by
// first appearance.
func Placeholders(text string) []string {
	var out []string
	seen := map[string]bool{}
	for _, m := range placeholder.FindAllStringSubmatch(text, -1) {
		if !seen[m[1]] {
			seen[m[1]] = true
			out = append(out, m[1])
		}
	}
	return out
}

// fill replaces each `{name}` with its value. A placeholder with no value is
// left as written, which is what makes a mistake visible instead of silent.
func fill(text string, args []any) string {
	if len(args) == 0 {
		return text
	}
	values := map[string]string{}
	for i := 0; i+1 < len(args); i += 2 {
		values[fmt.Sprint(args[i])] = fmt.Sprint(args[i+1])
	}
	return placeholder.ReplaceAllStringFunc(text, func(m string) string {
		if v, ok := values[m[1:len(m)-1]]; ok {
			return v
		}
		return m
	})
}
