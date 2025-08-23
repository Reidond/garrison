package reforger

import (
	"context"
	"encoding/json"
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

func TestProvider_InitConfig_WritesJSON_Golden(t *testing.T) {
	td := t.TempDir()
	out := filepath.Join(td, "ServerConfig.json")
	p := Provider{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	opts := map[string]string{
		"name":           "My Server",
		"scenarioId":     "/Game/Scenarios/conf/everon.conf",
		"a2sPort":        "17777",
		"publicPort":     "2001",
		"rconPort":       "19999",
		"rconPassword":   "secret",
		"admins":         "123,456",
		"fastValidation": "true",
	}
	if err := p.InitConfig(ctx, out, opts); err != nil {
		t.Fatalf("InitConfig error: %v", err)
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	// Re-marshal to normalize spacing for both
	var gotObj map[string]any
	if err := json.Unmarshal(got, &gotObj); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, string(got))
	}
	gotNorm, _ := json.MarshalIndent(gotObj, "", "  ")
	wantRaw, err := os.ReadFile(filepath.Join("testdata", "config_golden.json"))
	if err != nil {
		t.Fatal(err)
	}
	var wantObj map[string]any
	if err := json.Unmarshal(wantRaw, &wantObj); err != nil {
		t.Fatalf("invalid golden JSON: %v\n%s", err, string(wantRaw))
	}
	wantNorm, _ := json.MarshalIndent(wantObj, "", "  ")
	if string(gotNorm) != string(wantNorm) {
		t.Fatalf("mismatch\nGot:\n%s\nWant:\n%s", string(gotNorm), string(wantNorm))
	}
}
