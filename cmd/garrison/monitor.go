package main

import (
	"fmt"
	"os"
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
	// Load config to know ports and log file
	var ports []struct {
		Port  int
		Proto string
	}
	var logFile string
	if md.ConfigPath != "" {
		if cfg, err := server.LoadGameConfig(md.ConfigPath); err == nil {
			if parsed, err := server.ParsePorts(cfg.Ports); err == nil {
				ports = parsed
			}
			logFile = cfg.LogFile
		}
	}
	if c.Watch {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		var logF *os.File
		var lastSize int64
		for {
			select {
			case <-root.Context().Done():
				if logF != nil {
					_ = logF.Close()
				}
				return nil
			case <-ticker.C:
				status := "stopped"
				if m.IsRunning() {
					status = "running"
				}
				msg := status
				// Probe first TCP port (if any tcp)
				for _, p := range ports {
					if p.Proto == "tcp" {
						if server.ProbePortTCP(p.Port, 500*time.Millisecond) {
							msg += fmt.Sprintf(" tcp:%d=up", p.Port)
						} else {
							msg += fmt.Sprintf(" tcp:%d=down", p.Port)
						}
						break
					}
				}
				fmt.Printf("%s: %s\n", c.Name, msg)
				// Tail log if configured
				if logFile != "" {
					if logF == nil {
						lf, err := os.Open(server.ResolvePath(md.InstallDir, logFile))
						if err == nil {
							logF = lf
						}
					}
					if logF != nil {
						if info, err := logF.Stat(); err == nil {
							size := info.Size()
							if size > lastSize {
								buf := make([]byte, size-lastSize)
								if _, err := logF.ReadAt(buf, lastSize); err == nil {
									os.Stdout.Write(buf)
								}
								lastSize = size
							}
						}
					}
				}
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
