// Package arma contains helpers specific to the Arma Reforger dedicated server.
package arma

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	scmd "github.com/example/garrison/internal/steamcmd"
)

// AppID is the Steam App ID for Arma Reforger dedicated server.
const AppID = "1874900"

// ServerKey is the canonical key used across config and systemd helpers.
const ServerKey = "arma-reforger"

// BinaryPath returns the expected absolute path to the server binary.
func BinaryPath(installDir string) string {
	return filepath.Join(installDir, "ArmaReforgerServer")
}

// StartDirect launches the server process directly, writing pid and logs
// under the provided install/profile directories.
func StartDirect(installDir, config, profile string, port, qport, bport int, extra string) error {
	bin := BinaryPath(installDir)
	if _, err := os.Stat(bin); err != nil {
		return fmt.Errorf("server binary not found at %s; run install first: %w", bin, err)
	}
	if err := os.MkdirAll(profile, 0o755); err != nil {
		return err
	}
	pidFile := filepath.Join(installDir, "arma-reforger.pid")
	if _, err := os.Stat(pidFile); err == nil {
		return fmt.Errorf("pid file exists: %s (is server already running?)", pidFile)
	}
	args := []string{
		fmt.Sprintf("-config=%s", config),
		fmt.Sprintf("-profile=%s", profile),
		fmt.Sprintf("-port=%d", port),
		fmt.Sprintf("-queryPort=%d", qport),
		fmt.Sprintf("-steamQueryPort=%d", qport),
		fmt.Sprintf("-serverBrowserPort=%d", bport),
	}
	if strings.TrimSpace(extra) != "" {
		args = append(args, strings.Fields(extra)...)
	}
	proc := exec.Command(bin, args...)
	proc.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	logPath := filepath.Join(profile, "server.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer logFile.Close()
	proc.Stdout = logFile
	proc.Stderr = logFile
	if err := proc.Start(); err != nil {
		return err
	}
	pid := proc.Process.Pid
	if err := os.WriteFile(pidFile, []byte(fmt.Sprint(pid)), 0o644); err != nil {
		_ = proc.Process.Kill()
		return err
	}
	time.Sleep(2 * time.Second)
	return nil
}

// StopDirect terminates the directly-launched server using its pid file.
func StopDirect(installDir string) error {
	pidFile := filepath.Join(installDir, "arma-reforger.pid")
	pid, err := readPid(pidFile)
	if err != nil {
		return err
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	if err := p.Signal(syscall.SIGTERM); err != nil {
		_ = p.Kill()
	}
	_ = os.Remove(pidFile)
	return nil
}

// StatusDirect returns a human-readable status string for the direct mode.
func StatusDirect(installDir string) (string, error) {
	pidFile := filepath.Join(installDir, "arma-reforger.pid")
	pid, err := readPid(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "stopped", nil
		}
		return "", err
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return "stopped", nil
	}
	if err := p.Signal(syscall.Signal(0)); err != nil {
		return "stopped", nil
	}
	return fmt.Sprintf("running (pid=%d)", pid), nil
}

func readPid(pidFile string) (int, error) {
	b, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, err
	}
	var pid int
	_, err = fmt.Sscanf(string(b), "%d", &pid)
	return pid, err
}

// InstallOrUpdateDirect installs or updates the server into installDir using steamcmd directly.
// When validate is true, the steamcmd run includes the validate flag.
func InstallOrUpdateDirect(installDir string, validate bool) error {
	return scmd.Run(installDir, AppID, validate)
}

// InstallOrUpdateSystemd installs or updates the server using a transient systemd-run sandbox.
// All systemd-based helpers were removed in favor of container-based isolation.
