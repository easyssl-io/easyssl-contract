package catalog_test

import (
	"encoding/json"
	"testing"

	"github.com/easyssl-io/easyssl-contract/catalog"
)

// The Agent entry is how an Agent a Server no longer speaks to finds its way
// back, so its JSON shape never changes, whatever a later breaking release
// does to the rest of the contract. This literal is written once and never
// edited: if it fails, the change is what is wrong.
func TestTheAgentEntryShapeIsFrozen(t *testing.T) {
	const frozen = `{"plugins":[],"agent":{"version":"0.2.2","artifacts":[{"os":"linux","arch":"amd64","sha256":"5a713b52153da8911d0abdb7f32894138ebb58e533bb257021dcc912b98e02cb","url":"/api/v1/download/easyssl-agent/linux/amd64"}]}}`
	var r catalog.Response
	if err := json.Unmarshal([]byte(frozen), &r); err != nil {
		t.Fatal(err)
	}
	a := r.Agent
	if a == nil || a.Version != "0.2.2" || len(a.Artifacts) != 1 ||
		a.Artifacts[0] != (catalog.Artifact{OS: "linux", Arch: "amd64", SHA256: "5a713b52153da8911d0abdb7f32894138ebb58e533bb257021dcc912b98e02cb", URL: "/api/v1/download/easyssl-agent/linux/amd64"}) {
		t.Fatalf("decoded = %+v", a)
	}
	back, err := json.Marshal(r)
	if err != nil || string(back) != frozen {
		t.Fatalf("encoded = %s, %v", back, err)
	}
	if a.Artifacts[0].URL != "/api/v1"+catalog.AgentArtifactPath("linux", "amd64") {
		t.Fatalf("the route moved: %s", catalog.AgentArtifactPath("linux", "amd64"))
	}
}
