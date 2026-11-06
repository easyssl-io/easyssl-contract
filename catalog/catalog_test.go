package catalog

import "testing"

func TestArtifactPathsAreEscaped(t *testing.T) {
	if got := PluginArtifactPath("generic-cli", "linux", "amd64"); got != "/download/plugins/generic-cli/linux/amd64" {
		t.Errorf("PluginArtifactPath = %q", got)
	}
	if got := PluginArtifactPath("../x", "linux", "a/b"); got != "/download/plugins/..%2Fx/linux/a%2Fb" {
		t.Errorf("an id or platform added a segment: %q", got)
	}
	if got := AgentArtifactPath("linux", "arm64"); got != "/download/easyssl-agent/linux/arm64" {
		t.Errorf("AgentArtifactPath = %q", got)
	}
}
