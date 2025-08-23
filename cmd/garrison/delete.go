package main

import (
	"log/slog"

	"github.com/Reidond/garrison/internal/server"
)

type DeleteCmd struct {
	Name string `arg:"" help:"Server name"`
}

func (c *DeleteCmd) Run(root *Root) error {
	if err := server.DeleteMetadata(c.Name); err != nil {
		return err
	}
	slog.Info("deleted server", "name", c.Name)
	return nil
}
