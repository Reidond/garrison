package server

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

type Manager struct {
	Metadata Metadata
}

func NewManager(md Metadata) *Manager { return &Manager{Metadata: md} }

func (m *Manager) pidFile() (string, error) {
	base, err := dataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, fmt.Sprintf("%s.pid", m.Metadata.Name)), nil
}

func (m *Manager) writePID(pid int) error {
	p, err := m.pidFile()
	if err != nil {
		return err
	}
	return os.WriteFile(p, []byte(strconv.Itoa(pid)), 0o644)
}

func (m *Manager) readPID() (int, error) {
	p, err := m.pidFile()
	if err != nil {
		return 0, err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil {
		return 0, err
	}
	return pid, nil
}

func (m *Manager) removePID() { p, _ := m.pidFile(); _ = os.Remove(p) }

// Start launches the server using provided executable and args; here we assume executable lives in InstallDir
func (m *Manager) Start(executable string, args ...string) error {
	return m.StartWithOptions(ProcessOptions{}, executable, args...)
}

func (m *Manager) StartWithOptions(opts ProcessOptions, executable string, args ...string) error {
	if executable == "" {
		return errors.New("executable path required")
	}
	cmd := exec.Command(executable, args...)
	cmd.Dir = m.Metadata.InstallDir
	if opts.Env != nil {
		// Expand ${VAR} style env references against current environment
		expanded := make([]string, 0, len(opts.Env))
		for _, e := range opts.Env {
			expanded = append(expanded, os.ExpandEnv(e))
		}
		cmd.Env = expanded
	}
	// Drop privileges on Unix if requested
	if opts.UID > 0 || opts.GID > 0 || len(opts.Groups) > 0 {
		cmd.SysProcAttr = &syscall.SysProcAttr{Credential: &syscall.Credential{}}
		if opts.UID > 0 {
			cmd.SysProcAttr.Credential.Uid = uint32(opts.UID)
		}
		if opts.GID > 0 {
			cmd.SysProcAttr.Credential.Gid = uint32(opts.GID)
		}
		if len(opts.Groups) > 0 {
			gs := make([]uint32, len(opts.Groups))
			for i, g := range opts.Groups { gs[i] = uint32(g) }
			cmd.SysProcAttr.Credential.Groups = gs
		}
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	return m.writePID(cmd.Process.Pid)
}

func (m *Manager) Stop() error {
	pid, err := m.readPID()
	if err != nil {
		return err
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	// Send SIGTERM
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return err
	}
	m.removePID()
	return nil
}

func (m *Manager) IsRunning() bool {
	pid, err := m.readPID()
	if err != nil {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Unix, sending signal 0 checks process existence
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		return false
	}
	return true
}
