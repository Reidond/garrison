package server

import (
	"os"

	"gopkg.in/yaml.v3"
)

type GameConfig struct {
	AppID      int      `yaml:"app_id"`
	StartArgs  string   `yaml:"start_args"`
	Ports      []string `yaml:"ports"`
	Executable string   `yaml:"executable"`  // optional path to the server binary relative to install dir or absolute
	WorkingDir string   `yaml:"working_dir"` // optional working directory relative to install dir or absolute
}

func LoadGameConfig(path string) (GameConfig, error) {
	var cfg GameConfig
	b, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// end
