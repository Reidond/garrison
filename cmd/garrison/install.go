package main

import (
	"context"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/Reidond/garrison/internal/server"
	"github.com/Reidond/garrison/internal/steamcmd"
	"github.com/Reidond/garrison/internal/utils"
)

type InstallCmd struct {
	Name     string        `arg:"" help:"Server name"`
	Validate bool          `help:"Validate files during install"`
	Timeout  time.Duration `help:"Timeout for the install operation" default:"2h"`
}

func (c *InstallCmd) Run(root *Root) error {
	md, err := server.LoadMetadata(c.Name)
	if err != nil {
		return err
	}
	// Configure streaming and tee if stderr is a TTY
	cfg := steamcmd.Config{}
	if utils.DefaultStderrIsTTY() {
		logDir, _ := utils.ExpandPath("~/.garrison/logs")
		_ = utils.EnsureDir(logDir, 0o755)
		logFile := filepath.Join(logDir, "steamcmd-install-"+c.Name+".log")
		cb, closeFn, err := utils.NewLineTee("[steamcmd] ", logFile)
		if err == nil {
			defer closeFn()
			cfg.OnOutput = cb
			cfg.OutputPrefix = ""
			cfg.StreamOutput = true
		}
	}
	cli := steamcmd.New(cfg)
	// Wrap context with timeout if provided
	ctx := root.Context()
	if c.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.Timeout)
		defer cancel()
	}
	if err := cli.InstallApp(ctx, md.AppID, md.InstallDir, c.Validate); err != nil {
		return err
	}
	slog.Info("installed server", "name", c.Name, "dir", md.InstallDir)
	return nil
}
