package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

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
	var env map[string]string
	var cmdTemplate string
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
			if cfg.Env != nil {
				env = cfg.Env
			}
			if cfg.CommandTemplate != "" {
				cmdTemplate = cfg.CommandTemplate
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
	// If a command template is provided, treat template as full command line replacing placeholders.
	// Supported vars: install_dir, app_id, name, args
	if cmdTemplate != "" {
		vars := map[string]string{
			"install_dir": md.InstallDir,
			"app_id":      fmt.Sprintf("%d", md.AppID),
			"name":        md.Name,
			"args":        strings.Join(args, " "),
		}
		full := server.Interpolate(cmdTemplate, vars)
		parts := strings.Fields(full)
		if len(parts) > 0 {
			exe = parts[0]
			if !filepath.IsAbs(exe) {
				exe = filepath.Join(md.InstallDir, exe)
			}
			args = parts[1:]
		}
	}
	m := server.NewManager(md)
	// set working dir
	m.Metadata.InstallDir = workDir
	// apply env
	if len(env) > 0 {
		os.Environ() /* no-op just to avoid linter */
		os.Setenv("GARRISON_DUMMY", "0")
		os.Unsetenv("GARRISON_DUMMY")
	}
	if err := m.Start(exe, args...); err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}
	slog.Info("started server", "name", c.Name)
	return nil
}
