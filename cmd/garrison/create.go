package main

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Reidond/garrison/internal/games"
	"github.com/Reidond/garrison/internal/server"
	"github.com/Reidond/garrison/internal/utils"
)

type CreateCmd struct {
	Name       string `arg:"" help:"Server name"`
	AppID      int    `required:"" help:"Steam App ID"`
	Dir        string `help:"Installation directory (default: ~/.garrison/instances/<name>/install)" type:"path"`
	InitConfig bool   `help:"Initialize game config using parameters from the game's config file" default:"false"`
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

	// Initialize game config if requested
	if c.InitConfig {
		if err := c.initGameConfig(root, dir, cfgPath); err != nil {
			return fmt.Errorf("failed to initialize game config: %w", err)
		}
	}

	return nil
}

// initGameConfig initializes a game configuration file using parameters from reforger.yaml
func (c *CreateCmd) initGameConfig(root *Root, installDir string, cfgPath string) error {
	// Load game config which includes init config parameters
	gameConfig, err := server.LoadGameConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load game config %s: %w", cfgPath, err)
	}

	// Check if init config is available
	if gameConfig.Game == nil || gameConfig.Game.InitConfig == nil {
		return fmt.Errorf("no init config parameters found in %s", cfgPath)
	}

	initParams := gameConfig.Game.InitConfig

	// Find the game provider by iterating through all providers
	// and checking if any provider knows about this app_id
	var gameProvider string
	for _, provider := range games.All() {
		// We need to check if this provider handles this app_id
		// For now, we'll use the provider ID as a simple heuristic
		// In a real implementation, providers might have an AppIDs() method
		if provider.ID() == "reforger" && gameConfig.AppID == 1874900 {
			gameProvider = provider.ID()
			break
		}
		// Add more mappings as needed
	}

	if gameProvider == "" {
		return fmt.Errorf("no game provider found for app_id %d", gameConfig.AppID)
	}

	p, err := games.Get(gameProvider)
	if err != nil {
		return err
	}

	// Determine config output path
	configOutput := filepath.Join(installDir, "Configs", "ServerConfig.json")

	// Expand path if relative
	if abs, err := filepath.Abs(configOutput); err == nil {
		configOutput = abs
	}

	// Prepare server name - replace {{name}} placeholder if present
	serverName := initParams.Name
	if strings.Contains(serverName, "{{name}}") {
		serverName = strings.ReplaceAll(serverName, "{{name}}", c.Name)
	}
	if serverName == "" {
		serverName = c.Name
	}

	opts := map[string]string{
		"name":           serverName,
		"scenarioId":     initParams.ScenarioID,
		"a2sPort":        strconv.Itoa(initParams.A2SPort),
		"publicPort":     strconv.Itoa(initParams.PublicPort),
		"rconPort":       strconv.Itoa(initParams.RCONPort),
		"rconPassword":   initParams.RCONPassword,
		"admins":         strings.Join(initParams.Admins, ","),
		"fastValidation": strconv.FormatBool(initParams.FastValidation),
	}

	if err := p.InitConfig(root.Context(), configOutput, opts); err != nil {
		if err == games.ErrNotSupported {
			return fmt.Errorf("game %q does not support config initialization", gameProvider)
		}
		return err
	}

	slog.Info("initialized game config", "game", gameProvider, "config", configOutput)
	return nil
}
