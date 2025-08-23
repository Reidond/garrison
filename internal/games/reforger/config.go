package reforger

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// RootConfig models the Arma Reforger server JSON config.
type RootConfig struct {
	BindAddress   string           `json:"bindAddress,omitempty"`
	BindPort      int              `json:"bindPort,omitempty"`
	PublicAddress string           `json:"publicAddress,omitempty"`
	PublicPort    int              `json:"publicPort,omitempty"`
	A2S           *A2SConfig       `json:"a2s,omitempty"`
	RCON          *RCONConfig      `json:"rcon,omitempty"`
	Game          GameConfig       `json:"game"`
	Operating     *OperatingConfig `json:"operating,omitempty"`
}

type A2SConfig struct {
	Address string `json:"address,omitempty"`
	Port    int    `json:"port,omitempty"`
}

type RCONConfig struct {
	Address    string `json:"address,omitempty"`
	Port       int    `json:"port,omitempty"`
	Password   string `json:"password,omitempty"`
	MaxClients int    `json:"maxClients,omitempty"`
	Permission string `json:"permission,omitempty"`
}

type GameConfig struct {
	Name           string         `json:"name,omitempty"`
	Password       string         `json:"password,omitempty"`
	PasswordAdmin  string         `json:"passwordAdmin,omitempty"`
	Admins         []string       `json:"admins,omitempty"`
	ScenarioID     string         `json:"scenarioId"`
	GameProperties GameProperties `json:"gameProperties,omitempty"`
}

type GameProperties struct {
	FastValidation bool `json:"fastValidation"`
	BattlEye       bool `json:"battlEye,omitempty"`
}

type OperatingConfig struct {
	LobbyPlayerSynchronise bool `json:"lobbyPlayerSynchronise,omitempty"`
}

// WriteJSON writes the configuration to the given path creating parent dirs as needed.
func (rc RootConfig) WriteJSON(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(rc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
