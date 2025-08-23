package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Reidond/garrison/internal/games"
)

type GameCmd struct {
	Template   GameTemplateCmd   `cmd:"template" help:"Print a ready-made YAML game configuration template"`
	InitConfig GameInitConfigCmd `cmd:"init-config" help:"Create a native game config file via the provider (if supported)"`
}

type GameCommon struct {
	Game string `help:"Game id (e.g., reforger)" default:"reforger"`
}

type GameTemplateCmd struct {
	GameCommon
}

func (c *GameTemplateCmd) Run(root *Root) error {
	p, err := games.Get(c.Game)
	if err != nil {
		return err
	}
	fmt.Print(string(p.TemplateYAML()))
	return nil
}

type GameInitConfigCmd struct {
	GameCommon
	Output         string   `help:"Output config path" default:"./Configs/ServerConfig.json" type:"path"`
	Name           string   `help:"Server name" default:"My Server"`
	ScenarioID     string   `required:"" help:"Scenario identifier or .conf path"`
	A2SPort        int      `help:"A2S UDP port" default:"17777"`
	PublicPort     int      `help:"Public UDP port" default:"2001"`
	RCONPort       int      `help:"RCON UDP port (optional)" default:"19999"`
	RCONPassword   string   `help:"RCON password (required to enable RCON)"`
	Admins         []string `help:"List of admin IDs"`
	FastValidation bool     `help:"Enable fastValidation (if supported)" default:"true"`
	DryRun         bool     `help:"Print the generated config to stdout without writing the file" default:"false"`
}

func (c *GameInitConfigCmd) Run(root *Root) error {
	p, err := games.Get(c.Game)
	if err != nil {
		return err
	}
	out := c.Output
	if abs, err := filepath.Abs(out); err == nil {
		out = abs
	}
	opts := map[string]string{
		"name":           c.Name,
		"scenarioId":     c.ScenarioID,
		"a2sPort":        fmt.Sprintf("%d", c.A2SPort),
		"publicPort":     fmt.Sprintf("%d", c.PublicPort),
		"rconPort":       fmt.Sprintf("%d", c.RCONPort),
		"rconPassword":   c.RCONPassword,
		"admins":         strings.Join(c.Admins, ","),
		"fastValidation": fmt.Sprintf("%t", c.FastValidation),
	}
	if c.DryRun {
		// Generate to a temp file, print to stdout, then remove.
		tmpf, err := os.CreateTemp("", "garrison-config-*.json")
		if err != nil {
			return err
		}
		tmp := tmpf.Name()
		tmpf.Close()
		defer os.Remove(tmp)
		if err := p.InitConfig(root.Context(), tmp, opts); err != nil {
			if err == games.ErrNotSupported {
				return fmt.Errorf("game %q does not support init-config", c.Game)
			}
			return err
		}
		b, err := os.ReadFile(tmp)
		if err != nil {
			return err
		}
		fmt.Print(string(b))
		return nil
	} else {
		if err := p.InitConfig(root.Context(), out, opts); err != nil {
			if err == games.ErrNotSupported {
				return fmt.Errorf("game %q does not support init-config", c.Game)
			}
			return err
		}
		fmt.Printf("Wrote %s config to: %s\n", c.Game, out)
	}
	return nil
}
