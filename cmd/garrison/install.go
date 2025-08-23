package main

import (
	"log/slog"

	"github.com/Reidond/garrison/internal/server"
	"github.com/Reidond/garrison/internal/steamcmd"
)

type InstallCmd struct {
	Name     string `arg:"" help:"Server name"`
	Validate bool   `help:"Validate files during install"`
}

func (c *InstallCmd) Run(root *Root) error {
	md, err := server.LoadMetadata(c.Name)
	if err != nil {
		return err
	}
	cli := steamcmd.New(steamcmd.Config{})
	if err := cli.InstallApp(root.Context(), md.AppID, md.InstallDir, c.Validate); err != nil {
		return err
	}
	slog.Info("installed server", "name", c.Name, "dir", md.InstallDir)
	return nil
}
