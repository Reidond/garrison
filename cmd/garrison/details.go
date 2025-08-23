package main

import (
	"encoding/json"
	"fmt"

	"github.com/Reidond/garrison/internal/server"
)

type DetailsCmd struct {
	Name string `arg:"" help:"Server name"`
}

func (c *DetailsCmd) Run(root *Root) error {
	md, err := server.LoadMetadata(c.Name)
	if err != nil {
		return err
	}
	m := server.NewManager(md)
	payload := map[string]any{
		"metadata": md,
		"running":  m.IsRunning(),
	}
	b, _ := json.MarshalIndent(payload, "", "  ")
	fmt.Println(string(b))
	return nil
}
