package reforger

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/Reidond/garrison/internal/games"
)

// Test ListScenarios using a fake binary harness that prints conf names
func TestProvider_ListScenarios_FakeBinary(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake shell script not supported on windows in this test")
	}
	// Create temp dir and fake binary
	dir := t.TempDir()
	bin := filepath.Join(dir, "ArmaReforgerServer")
	script := "#!/usr/bin/env bash\necho 'Available scenarios:'\necho '/Game/Scenarios/conf/everon.conf'\necho 'some/other/SCENARIO.CONF'\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	p := Provider{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	scenarios, raw, err := p.ListScenarios(ctx, games.ScenarioOptions{Binary: bin, InstallDir: dir})
	if err != nil {
		t.Fatalf("ListScenarios error: %v\nraw:%s", err, raw)
	}
	if len(scenarios) == 0 {
		t.Fatalf("expected scenarios parsed, got none; raw: %s", raw)
	}
	// Ensure .conf entries normalized and present
	got := strings.Join(scenarios, "\n")
	if !strings.Contains(strings.ToLower(got), "everon.conf") {
		t.Fatalf("expected everon.conf in %q", got)
	}
}
