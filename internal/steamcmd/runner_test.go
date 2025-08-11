package steamcmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunCreatesInstallDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GARRISON_STEAMCMD_BIN", "true")
	// This will run bash -lc which may fail in CI if bash is missing; but Go official images have bash.
	_ = Run(filepath.Join(dir, "inst"), "999", false)
	if _, err := os.Stat(filepath.Join(dir, "inst")); err != nil {
		t.Fatalf("expected install dir created: %v", err)
	}
}
