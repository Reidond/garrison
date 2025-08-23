package main

import (
	"log/slog"

	"github.com/Reidond/garrison/internal/server"
	"github.com/Reidond/garrison/internal/steamcmd"
)

type UpdateCmd struct {
	Name string `arg:"" help:"Server name"`
}

func (c *UpdateCmd) Run(root *Root) error {
	md, err := server.LoadMetadata(c.Name)
	if err != nil {
		return err
	}
	cli := steamcmd.New(steamcmd.Config{})
	if err := cli.UpdateApp(root.Context(), md.AppID, md.InstallDir); err != nil {
		return err
	}
	slog.Info("updated server", "name", c.Name)
	return nil
}
