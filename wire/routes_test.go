package wire

import "testing"

func TestPathEscapesTheID(t *testing.T) {
	for id, want := range map[string]string{
		"dep-1":  "/agent/deployments/dep-1/result",
		"a/../b": "/agent/deployments/a%2F..%2Fb/result",
		"x?y#z":  "/agent/deployments/x%3Fy%23z/result",
		"":       "/agent/deployments//result",
	} {
		if got := Path(RouteAgentDeploymentResult, id); got != want {
			t.Errorf("Path(%q) = %q, want %q", id, got, want)
		}
	}
	if got := Path(RouteAgentCheckin, "ignored"); got != RouteAgentCheckin {
		t.Errorf("a route without a placeholder became %q", got)
	}
}
