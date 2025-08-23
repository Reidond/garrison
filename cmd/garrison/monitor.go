package main

import (
	"fmt"
	"time"

	"github.com/Reidond/garrison/internal/server"
)

type MonitorCmd struct {
	Name  string `arg:"" help:"Server name"`
	Watch bool   `help:"Continuous monitoring"`
}

func (c *MonitorCmd) Run(root *Root) error {
	md, err := server.LoadMetadata(c.Name)
	if err != nil {
		return err
	}
	m := server.NewManager(md)
	if c.Watch {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-root.Context().Done():
				return nil
			case <-ticker.C:
				status := "stopped"
				if m.IsRunning() {
					status = "running"
				}
				fmt.Printf("%s: %s\n", c.Name, status)
			}
		}
	}
	status := "stopped"
	if m.IsRunning() {
		status = "running"
	}
	fmt.Printf("%s: %s\n", c.Name, status)
	return nil
}
