package main

import (
	"log/slog"

	"github.com/Reidond/garrison/internal/server"
)

type RestartCmd struct {
	Name string `arg:"" help:"Server name"`
}

func (c *RestartCmd) Run(root *Root) error {
	md, err := server.LoadMetadata(c.Name)
	if err != nil {
		return err
	}
	m := server.NewManager(md)
	_ = m.Stop()
	// For a real implementation, we'd re-use Start logic and config
	slog.Info("restarting server", "name", c.Name)
	return nil
}
