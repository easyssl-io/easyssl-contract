package plugincontract

import "testing"

func TestBreakingVersion(t *testing.T) {
	for v, want := range map[string]string{
		"0.0.1": "0.0", "0.1.0": "0.1", "0.1.7": "0.1", "1.0.0": "1", "1.4.2": "1", "12.0.3": "12",
		"": "", "1.2": "", "1.2.3.4": "", "v1.2.3": "", "1.2.x": "", "01.2.3": "", "1..3": "",
	} {
		got, ok := BreakingVersion(v)
		if got != want || ok != (want != "") {
			t.Errorf("BreakingVersion(%q) = %q, %v; want %q", v, got, ok, want)
		}
	}
}

func TestCompatible(t *testing.T) {
	own, _ := BreakingVersion(ContractVersion)
	if !Compatible(ContractVersion) {
		t.Error("a build does not speak its own version")
	}
	if Compatible("") || Compatible("garbage") {
		t.Error("a missing or malformed version is compatible")
	}
	if own == "0.0" && Compatible("0.1.0") {
		t.Error("another breaking version is compatible")
	}
	if own == "0.0" && !Compatible("0.0.9") {
		t.Error("a patch release of the same breaking version is not compatible")
	}
}
