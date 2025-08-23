package main

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/Reidond/garrison/internal/server"
	"github.com/Reidond/garrison/internal/utils"
)

type CreateCmd struct {
	Name  string `arg:"" help:"Server name"`
	AppID int    `required:"" help:"Steam App ID"`
	Dir   string `help:"Installation directory" default:"." type:"path"`
}

func (c *CreateCmd) Run(root *Root) error {
	// Expand paths
	dir, err := utils.ExpandPath(c.Dir)
	if err != nil {
		return err
	}
	if err := utils.EnsureDir(dir, 0o755); err != nil {
		return err
	}

	// optional default: configs/<appid>.yaml if exists
	cfgPath := filepath.Join("configs", fmt.Sprintf("%d.yaml", c.AppID))

	md := server.Metadata{
		Name:       c.Name,
		InstallDir: dir,
		AppID:      c.AppID,
		ConfigPath: cfgPath,
	}
	if err := server.SaveMetadata(md); err != nil {
		return err
	}
	slog.Info("created server", "name", c.Name, "dir", dir, "app_id", c.AppID)
	return nil
}
