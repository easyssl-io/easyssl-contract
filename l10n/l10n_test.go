package l10n

import (
	"testing"
	"testing/fstest"
)

func TestBase(t *testing.T) {
	for tag, want := range map[string]string{
		"es": "es", "es-PE": "es", "es_419": "es", "es_PE.UTF-8": "es", "ES": "es",
		"en_US.UTF-8": "en", "en-GB": "en",
		"C": "", "POSIX": "", "": "", "fr_FR.UTF-8": "", "español": "",
	} {
		got, ok := Base(tag)
		if got != want || ok != (want != "") {
			t.Errorf("Base(%q) = %q, %v; want %q", tag, got, ok, want)
		}
	}
}

func TestFromLocale(t *testing.T) {
	env := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }
	if l, ok := FromLocale(env(map[string]string{"LANG": "es_PE.UTF-8"})); !ok || l != "es" {
		t.Errorf("LANG alone = %q, %v", l, ok)
	}
	// LC_ALL decides when set, even when what it names is not a language.
	if _, ok := FromLocale(env(map[string]string{"LC_ALL": "C", "LANG": "es_PE.UTF-8"})); ok {
		t.Error("LC_ALL=C must not resolve to the LANG behind it")
	}
	if _, ok := FromLocale(env(nil)); ok {
		t.Error("no locale must not resolve")
	}
}

func TestTextValidate(t *testing.T) {
	if err := (Text{}).Validate(); err != nil {
		t.Errorf("empty: %v", err)
	}
	if err := (Text{"en": "a", "es": "b", "pt": "c"}).Validate(); err != nil {
		t.Errorf("valid: %v", err)
	}
	if (Text{"es": "b"}).Validate() == nil {
		t.Error("a text without English must be refused")
	}
	if (Text{"en": "a", "es-PE": "b"}).Validate() == nil {
		t.Error("a regional key must be refused")
	}
	if got := (Text{"en": "a"}).In("es"); got != "a" {
		t.Errorf("fallback = %q", got)
	}
}

func TestPrinter(t *testing.T) {
	c := MustLoad(fstest.MapFS{"locales/es.json": {Data: []byte(`{
		"a.hello": "Hola, {name}, desde {host}.",
		"a.keys": {"one": "Queda {n} clave en {dir}.", "other": "Quedan {n} claves en {dir}."},
		"a.empty": ""
	}`)}}, "locales")
	es, en := c.In("es_PE.UTF-8"), c.In("fr")
	if en.Language() != "en" {
		t.Fatalf("unsupported language speaks %q", en.Language())
	}
	if got := es.T("a.hello", "Hello from {host}, {name}.", "name", "web-01", "host", "h"); got != "Hola, web-01, desde h." {
		t.Errorf("es = %q", got)
	}
	if got := en.T("a.hello", "Hello from {host}, {name}.", "name", "web-01", "host", "h"); got != "Hello from h, web-01." {
		t.Errorf("en = %q", got)
	}
	// Untranslated and empty both read as English, never as the ID.
	for _, id := range []string{"a.missing", "a.empty"} {
		if got := es.T(id, "English."); got != "English." {
			t.Errorf("%s = %q", id, got)
		}
	}
	if got := es.N("a.keys", 1, "{n} key remains in {dir}.", "{n} keys remain in {dir}.", "dir", "/k"); got != "Queda 1 clave en /k." {
		t.Errorf("one = %q", got)
	}
	if got := es.N("a.keys", 3, "{n} key remains in {dir}.", "{n} keys remain in {dir}.", "dir", "/k"); got != "Quedan 3 claves en /k." {
		t.Errorf("other = %q", got)
	}
	text := c.Text("a.hello", "Hello from {host}, {name}.", "name", "n", "host", "h")
	if text["en"] != "Hello from h, n." || text["es"] != "Hola, n, desde h." || text.Validate() != nil {
		t.Errorf("Text = %v", text)
	}
}
