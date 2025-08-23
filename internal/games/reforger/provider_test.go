package reforger

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

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
