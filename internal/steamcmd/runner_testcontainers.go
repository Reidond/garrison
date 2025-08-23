//go:build integration
// +build integration

package steamcmd

import (
	"context"
	"fmt"
	"io"

	tc "github.com/testcontainers/testcontainers-go"
)

// containerRunner runs steamcmd inside a Docker container using testcontainers-go.
// It pulls and starts a container with steamcmd installed and executes commands with tc.Container.Exec.
type containerRunner struct {
	container tc.Container
	bin       string
}

// newContainerRunner starts a steamcmd container and returns a runner.
func newContainerRunner(ctx context.Context, image string) (*containerRunner, error) {
	if image == "" {
		// Default to cm2network/steamcmd:latest, a commonly used steamcmd image.
		image = "cm2network/steamcmd:latest"
	}
	req := tc.ContainerRequest{
		Image:        image,
		ExposedPorts: []string{},
		Entrypoint:   []string{"/bin/sh", "-lc", "sleep infinity"},
	}
	cont, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}
	return &containerRunner{container: cont, bin: "steamcmd"}, nil
}

func (r *containerRunner) Terminate(ctx context.Context) error {
	if r.container != nil {
		return r.container.Terminate(ctx)
	}
	return nil
}

func (r *containerRunner) Run(ctx context.Context, bin string, args ...string) (string, int, error) {
	if bin == "" {
		bin = r.bin
	}
	// Build command string to exec inside container.
	cmd := append([]string{bin}, args...)
	exitCode, rdr, err := r.container.Exec(ctx, cmd)
	var out string
	if rdr != nil {
		b, _ := io.ReadAll(rdr)
		out = string(b)
	}
	if err != nil {
		return out, exitCode, fmt.Errorf("exec error: %w; out=%s", err, out)
	}
	return out, exitCode, nil
}
