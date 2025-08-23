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

type UpdateCmd struct {
	Name    string        `arg:"" help:"Server name"`
	Timeout time.Duration `help:"Timeout for the update operation" default:"2h"`
}

func (c *UpdateCmd) Run(root *Root) error {
	md, err := server.LoadMetadata(c.Name)
	if err != nil {
		return err
	}
	cfg := steamcmd.Config{}
	if utils.DefaultStderrIsTTY() {
		logDir, _ := utils.ExpandPath("~/.garrison/logs")
		_ = utils.EnsureDir(logDir, 0o755)
		logFile := filepath.Join(logDir, "steamcmd-update-"+c.Name+".log")
		cb, closeFn, err := utils.NewLineTee("[steamcmd] ", logFile)
		if err == nil {
			defer closeFn()
			cfg.OnOutput = cb
			cfg.OutputPrefix = ""
			cfg.StreamOutput = true
		}
	}
	cli := steamcmd.New(cfg)
	ctx := root.Context()
	if c.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.Timeout)
		defer cancel()
	}
	if err := cli.UpdateApp(ctx, md.AppID, md.InstallDir); err != nil {
		return err
	}
	slog.Info("updated server", "name", c.Name)
	return nil
}
