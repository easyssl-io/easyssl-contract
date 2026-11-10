package plugincontract

import "testing"

func TestAtLeast(t *testing.T) {
	for _, c := range []struct {
		v, min string
		want   bool
	}{
		{"0.1.1", "0.1.1", true},
		{"0.1.0", "0.1.1", false},
		{"0.2.0", "0.1.1", true},
		{"1.0.0", "0.9.9", true},
		{"0.1.10", "0.1.9", true},
		{"0.1.1+dev.a128250e9719", "0.1.1", true},
		{"0.1.0+dev.a128250e9719", "0.1.1", false},
		{"0.1", "0.1.1", false},
		{"", "0.1.1", false},
		{"v0.1.1", "0.1.1", false},
	} {
		if got := AtLeast(c.v, c.min); got != c.want {
			t.Errorf("AtLeast(%q, %q) = %v, want %v", c.v, c.min, got, c.want)
		}
	}
}
