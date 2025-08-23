//go:build integration
// +build integration

package steamcmd

import (
	"context"
	"io"
	"os"
	"testing"
	"time"
)

// To run: RUN_STEAMCMD_IT=1 go test -run TestSteamcmdPingLogin -tags=integration ./internal/steamcmd -v
func TestSteamcmdPingLogin(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode")
	}
	if os.Getenv("RUN_STEAMCMD_IT") == "" {
		t.Skip("set RUN_STEAMCMD_IT=1 to run")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Use the runner implemented in runner_testcontainers.go (built with integration tag)
	r, err := newContainerRunner(ctx, "")
	if err != nil {
		t.Fatalf("failed to start steamcmd container: %v", err)
	}
	defer func() {
		_ = r.Terminate(context.Background())
	}()

	c := NewWithRunner(Config{BinaryPath: "/home/steam/steamcmd/steamcmd.sh", Retries: 1, Backoff: 500}, r)
	if err := c.PingLogin(ctx); err != nil {
		// Fetch logs for easier diagnosis
		logs, _ := r.container.Logs(ctx)
		var tail string
		if logs != nil {
			b, _ := io.ReadAll(logs)
			tail = string(b)
		}
		t.Fatalf("PingLogin failed: %v\ncontainer logs:\n%s", err, tail)
	}
}
