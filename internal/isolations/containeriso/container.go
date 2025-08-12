package containeriso

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	execu "github.com/example/garrison/internal/executil"
	scmd "github.com/example/garrison/internal/steamcmd"
)

// Note: This is a minimal scaffold that reuses direct-mode paths and binaries,
// but executes them inside a container runtime (docker/podman) via CLI.
// It assumes the user has a working runtime and permission to use it.

// runtimeBin returns the chosen container runtime (docker|podman), preferring docker.
func runtimeBin() string {
	if p := strings.TrimSpace(os.Getenv("GARRISON_CONTAINER_RUNTIME")); p != "" {
		return p
	}
	return "docker"
}

func InstallOrUpdate(serverKey, appID, installDir string, validate bool) error {
	cacheDir := computeCacheDir()
	absInstall, _ := filepath.Abs(installDir)
	absCache, _ := filepath.Abs(cacheDir)
	if err := os.MkdirAll(absInstall, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(absCache, 0o755); err != nil {
		return err
	}

	image := firstNonEmpty(os.Getenv("GARRISON_STEAMCMD_IMAGE"), "steamcmd/steamcmd:latest")

	// Install/update app and quit via runscript to ensure no interactive state
	rs := scmd.BuildRunscriptContent(appID, validate, "/data/app")
	tmp := filepath.Join(absInstall, ".garrison_steamcmd_run.txt")
	if err := os.WriteFile(tmp, []byte(rs), 0o600); err != nil {
		return err
	}

	var args []string
	base := []string{
		runtimeBin(), "run", "--rm",
		"-e", "HOME=/cache",
		"-w", "/data",
		"-v", absInstall + ":/data",
		"-v", absCache + ":/cache",
		image,
	}

	args = append(base, "steamcmd", "+runscript", "/data/.garrison_steamcmd_run.txt")
	err := execu.Default.Run(args[0], args[1:]...)
	_ = os.Remove(tmp)
	return err
}

// StartGeneric runs the provided command inside a container with standard UDP port mappings
// and a bind-mounted install/profile directory. The command should be an absolute path inside
// the container (e.g., /data/app/ArmaReforgerServer) with its flags.
func StartGeneric(serverKey, installDir string, containerCmd []string, udpPorts [][2]int) error {
	absInstall, _ := filepath.Abs(installDir)
	image := firstNonEmpty(os.Getenv("GARRISON_SERVER_IMAGE"), "debian:bookworm-slim")
	name := "garrison-" + serverKey
	// If a container already exists, reuse it: start if stopped, no-op if running.
	if out, err := execu.Default.CombinedOutput(runtimeBin(), "inspect", "-f", "{{.State.Running}}", name); err == nil {
		state := strings.TrimSpace(out)
		if state == "true" {
			return nil
		}
		if state == "false" {
			if err := execu.Default.Run(runtimeBin(), "start", name); err == nil {
				return nil
			}
			// start failed; remove and recreate fresh
			_ = execu.Default.Run(runtimeBin(), "rm", "-f", name)
		}
		// Unknown state: fall through to create a fresh container
	} else {
		// inspect failed: ensure no stale blocking name (best-effort)
		_ = execu.Default.Run(runtimeBin(), "rm", "-f", name)
	}
	// Ports are UDP; map host:container
	runArgs := []string{runtimeBin(), "run", "-d", "--name", name, "--restart", "unless-stopped", "-v", absInstall + ":/data"}
	for _, p := range udpPorts {
		runArgs = append(runArgs, "-p", fmt.Sprintf("%d:%d/udp", p[0], p[1]))
	}
	// image must come after all options
	runArgs = append(runArgs, image)
	// then the command and its args
	runArgs = append(runArgs, containerCmd...)
	if strings.TrimSpace(os.Getenv("GARRISON_DEBUG")) != "" {
		fmt.Println("[garrison]", strings.Join(runArgs, " "))
	}
	return execu.Default.Run(runArgs[0], runArgs[1:]...)
}

func Stop(serverKey string) error {
	name := "garrison-" + serverKey
	return execu.Default.Run(runtimeBin(), "stop", name)
}

func Status(serverKey string) error {
	name := "garrison-" + serverKey
	// Print friendly status
	if err := execu.Default.Run(runtimeBin(), "ps", "--filter", "name="+name); err != nil {
		return err
	}
	return nil
}

func Delete(serverKey, installDir string) error {
	name := "garrison-" + serverKey
	_ = execu.Default.Run(runtimeBin(), "rm", "-f", name)
	if installDir != "" {
		_ = os.RemoveAll(installDir)
	}
	return nil
}

// computeCacheDir returns a user cache dir for steamcmd
func computeCacheDir() string {
	if v := strings.TrimSpace(os.Getenv("GARRISON_STEAMCMD_CACHE")); v != "" {
		return v
	}
	home, _ := os.UserHomeDir()
	if ch := strings.TrimSpace(os.Getenv("XDG_CACHE_HOME")); ch != "" {
		return filepath.Join(ch, "garrison", "steamcmd")
	}
	return filepath.Join(home, ".cache", "garrison", "steamcmd")
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
