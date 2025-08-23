package main

import (
	"log/slog"

	"github.com/Reidond/garrison/internal/server"
)

type StopCmd struct {
	Name string `arg:"" help:"Server name"`
}

func (c *StopCmd) Run(root *Root) error {
	md, err := server.LoadMetadata(c.Name)
	if err != nil {
		return err
	}
	m := server.NewManager(md)
	if err := m.Stop(); err != nil {
		return err
	}
	slog.Info("stopped server", "name", c.Name)
	return nil
}
