package reforger

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Reidond/garrison/internal/games"
)

// Provider implements games.Provider for Arma Reforger.
// It shells out to the server binary for scenario listing and serves a YAML template.

type Provider struct{}

func (Provider) ID() string   { return "reforger" }
func (Provider) Name() string { return "Arma Reforger" }

func init() {
	games.Register(Provider{})
}

func (Provider) TemplateYAML() []byte {
	// Reuse the repository template, but return inline to avoid imports.
	return []byte(`# Example garrison game config for Arma Reforger server
app_id: 1874900
executable: "ArmaReforgerServer"
working_dir: "."
ports:
  - "2001/udp"
  - "17777/udp"
  - "19999/udp"
command_template: >-
  {{install_dir}}/ArmaReforgerServer
  -profile {{install_dir}}/profile
  -config {{install_dir}}/Configs/ServerConfig.json
  -maxFPS 120
  {{args}}
start_args: ""
`)
}

func (Provider) InitConfig(ctx context.Context, outputPath string, options map[string]string) error {
	// Expected options: name, scenarioId, a2sPort, publicPort, rconPort, rconPassword, admins (comma-separated), fastValidation (true/false)
	get := func(k, def string) string {
		if v, ok := options[k]; ok {
			return v
		}
		return def
	}
	// Required: scenarioId
	scenario := strings.TrimSpace(get("scenarioId", ""))
	if scenario == "" {
		return fmt.Errorf("scenarioId is required")
	}
	// Ports validation
	parsePort := func(key string, def int) (int, error) {
		if v, ok := options[key]; ok && strings.TrimSpace(v) != "" {
			n, err := strconv.Atoi(v)
			if err != nil {
				return 0, fmt.Errorf("invalid %s: %v", key, err)
			}
			if n < 1 || n > 65535 {
				return 0, fmt.Errorf("invalid %s: must be 1-65535", key)
			}
			return n, nil
		}
		return def, nil
	}
	publicPort, err := parsePort("publicPort", 2001)
	if err != nil {
		return err
	}
	a2sPort, err := parsePort("a2sPort", 17777)
	if err != nil {
		return err
	}
	// Booleans and lists
	fastValidation := true
	if v, ok := options["fastValidation"]; ok && strings.TrimSpace(v) != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("invalid fastValidation: %v", err)
		}
		fastValidation = b
	}
	splitCSV := func(s string) []string {
		if strings.TrimSpace(s) == "" {
			return nil
		}
		parts := strings.Split(s, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
		return out
	}
	cfg := RootConfig{
		PublicPort: publicPort,
		A2S:        &A2SConfig{Port: a2sPort},
		Game: GameConfig{
			Name:       get("name", "My Server"),
			Admins:     splitCSV(get("admins", "")),
			ScenarioID: scenario,
			GameProperties: GameProperties{
				FastValidation: fastValidation,
			},
		},
	}
	if pw := strings.TrimSpace(get("rconPassword", "")); pw != "" {
		rconPort, err := parsePort("rconPort", 19999)
		if err != nil {
			return err
		}
		cfg.RCON = &RCONConfig{Port: rconPort, Password: pw}
	}
	return cfg.WriteJSON(outputPath)
}
