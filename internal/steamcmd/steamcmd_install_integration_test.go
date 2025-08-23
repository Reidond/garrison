//go:build integration
// +build integration

package steamcmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Reidond/garrison/internal/utils"
)

// To run: RUN_STEAMCMD_IT=1 SMALL_APP_ID=1007 go test -tags=integration ./internal/steamcmd -run InstallUpdate -v
// SMALL_APP_ID can be a small app like 1007 (Steamworks SDK redistributables). Adjust as needed.
func TestInstallUpdateApp(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode")
	}
	if os.Getenv("RUN_STEAMCMD_IT") == "" {
		t.Skip("set RUN_STEAMCMD_IT=1 to run")
	}
	appIDStr := os.Getenv("SMALL_APP_ID")
	if appIDStr == "" {
		// Default to a tiny app ID; override via env if needed
		appIDStr = "1007"
	}
	var appID int
	_, err := fmt.Sscanf(appIDStr, "%d", &appID)
	if err != nil || appID <= 0 {
		t.Fatalf("invalid SMALL_APP_ID: %q", appIDStr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	// Create a temporary host directory to mount as install target
	hostDir := t.TempDir()
	containerDir := "/data/install"

	image := os.Getenv("STEAMCMD_IMAGE")
	bin := os.Getenv("STEAMCMD_BIN")
	if bin == "" {
		// default for cm2network/steamcmd
		bin = "/home/steam/steamcmd/steamcmd.sh"
	}
	bind := hostDir + ":" + containerDir

	// Log environment/config used for visibility
	t.Logf("using image=%q bin=%q appID=%d hostDir=%q containerDir=%q", image, bin, appID, hostDir, containerDir)

	r, err := newContainerRunnerWithMounts(ctx, image, []string{bind})
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}
	defer func() { _ = r.Terminate(context.Background()) }()

	c := NewWithRunner(Config{BinaryPath: bin, Retries: 2, Backoff: 1500}, r)

	// Install app
	if err := c.InstallApp(ctx, appID, containerDir, false); err != nil {
		t.Fatalf("InstallApp failed: %v", err)
	}

	// Log installed files from host mount with pretty tree
	t.Logf("after install, listing files under %s:", hostDir)
	utils.DirLogger{Printf: t.Logf, Prefix: "  "}.LogDirTree(hostDir, 500)

	// Basic sanity: directory should have some content
	entries, err := os.ReadDir(hostDir)
	if err != nil {
		t.Fatalf("reading host dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatalf("expected content in %s after install", hostDir)
	}

	// Update app
	if err := c.UpdateApp(ctx, appID, containerDir); err != nil {
		t.Fatalf("UpdateApp failed: %v", err)
	}

	// Log files again after update
	t.Logf("after update, listing files under %s:", hostDir)
	utils.DirLogger{Printf: t.Logf, Prefix: "  "}.LogDirTree(hostDir, 500)

	// Re-check content remains
	if _, err := os.Stat(filepath.Join(hostDir, entries[0].Name())); err != nil {
		t.Fatalf("expected installed file to persist: %v", err)
	}
}
