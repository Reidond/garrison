package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/Reidond/garrison/internal/games"
)

type GameCmd struct {
	ListScenarios GameListScenariosCmd `cmd:"list-scenarios" help:"List available scenarios/maps for a game"`
	Template      GameTemplateCmd      `cmd:"template" help:"Print a ready-made YAML game configuration template"`
	InitConfig    GameInitConfigCmd    `cmd:"init-config" help:"Create a native game config file via the provider (if supported)"`
}

type GameCommon struct {
	Game string `help:"Game id (e.g., reforger)" default:"reforger"`
}

type GameListScenariosCmd struct {
	GameCommon
	Binary     string        `help:"Path or name of the server binary" default:""`
	InstallDir string        `help:"Install/working directory" type:"path"`
	Timeout    time.Duration `help:"Timeout for the query" default:"2m"`
	Raw        bool          `help:"Print raw output without parsing" default:"false"`
}

func (c *GameListScenariosCmd) Run(root *Root) error {
	p, err := games.Get(c.Game)
	if err != nil {
		return err
	}
	// Expand InstallDir to absolute for consistency in providers that might not expand
	if c.InstallDir != "" {
		if abs, err := filepath.Abs(c.InstallDir); err == nil {
			c.InstallDir = abs
		}
	}
	scenarios, raw, err := p.ListScenarios(root.Context(), games.ScenarioOptions{
		Binary:     c.Binary,
		InstallDir: c.InstallDir,
		Timeout:    c.Timeout,
		Raw:        c.Raw,
	})
	if err != nil {
		return err
	}
	if c.Raw {
		fmt.Print(raw)
		return nil
	}
	if len(scenarios) == 0 {
		// No parsed scenarios but maybe raw had info; print raw.
		fmt.Print(raw)
		return nil
	}
	fmt.Println("Scenarios:")
	for _, s := range scenarios {
		fmt.Printf("  %s\n", s)
	}
	return nil
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
	if err := p.InitConfig(root.Context(), out, opts); err != nil {
		if err == games.ErrNotSupported {
			return fmt.Errorf("game %q does not support init-config", c.Game)
		}
		return err
	}
	fmt.Printf("Wrote %s config to: %s\n", c.Game, out)
	return nil
}
