package main

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"os"

	"github.com/Reidond/garrison/internal/server"
)

type StartCmd struct {
	Name string `arg:"" help:"Server name"`
	Args string `help:"Additional start arguments"`
}

func (c *StartCmd) Run(root *Root) error {
	md, err := server.LoadMetadata(c.Name)
	if err != nil {
		return err
	}
	// Load game config if present; otherwise assume run.sh
	exe := filepath.Join(md.InstallDir, "run.sh")
	workDir := md.InstallDir
	if md.ConfigPath != "" {
		cfg, err := server.LoadGameConfig(md.ConfigPath)
		if err == nil {
			if cfg.Executable != "" {
				if filepath.IsAbs(cfg.Executable) {
					exe = cfg.Executable
				} else {
					exe = filepath.Join(md.InstallDir, cfg.Executable)
				}
			}
			if cfg.WorkingDir != "" {
				if filepath.IsAbs(cfg.WorkingDir) {
					workDir = cfg.WorkingDir
				} else {
					workDir = filepath.Join(md.InstallDir, cfg.WorkingDir)
				}
			}
			if cfg.StartArgs != "" {
				c.Args = strings.TrimSpace(strings.Join([]string{cfg.StartArgs, c.Args}, " "))
			}
		}
	}

	// Ensure working directory exists
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return err
	}
	args := []string{}
	if c.Args != "" {
		args = append(args, strings.Fields(c.Args)...)
	}
	m := server.NewManager(md)
	// temporarily set the executable's working dir using manager.Metadata.InstallDir; change to workDir
	m.Metadata.InstallDir = workDir
	if err := m.Start(exe, args...); err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}
	slog.Info("started server", "name", c.Name)
	return nil
}
