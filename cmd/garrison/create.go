package main

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/Reidond/garrison/internal/server"
	"github.com/Reidond/garrison/internal/utils"
)

type CreateCmd struct {
	Name  string `arg:"" help:"Server name"`
	AppID int    `required:"" help:"Steam App ID"`
	Dir   string `help:"Installation directory (default: ~/.garrison/instances/<name>/install)" type:"path"`
}

func (c *CreateCmd) Run(root *Root) error {
	// Determine installation directory
	var dir string
	var err error
	if strings.TrimSpace(c.Dir) == "" {
		// Default to ~/.garrison/instances/<name>/install
		base, berr := utils.ExpandPath("~/.garrison/instances")
		if berr != nil {
			return berr
		}
		dir = filepath.Join(base, c.Name, "install")
	} else {
		dir, err = utils.ExpandPath(c.Dir)
		if err != nil {
			return err
		}
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
