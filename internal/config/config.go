package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type IsolationMode string

const (
	IsolationNone    IsolationMode = "none"
	IsolationSystemd IsolationMode = "systemd"
)

type ServerConfig struct {
	Isolation IsolationMode `json:"isolation,omitempty"`
}

type Config struct {
	Servers map[string]ServerConfig `json:"servers,omitempty"`
}

// Load reads the global garrison config.
func Load() (Config, error) {
	cfgPath, err := configPath()
	if err != nil {
		return Config{}, err
	}
	b, err := os.ReadFile(cfgPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, nil
		}
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Save writes the global garrison config.
func Save(cfg Config) error {
	cfgPath, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cfgPath, b, 0o644)
}

// GetIsolation returns the saved isolation mode for a server key (e.g. "arma-reforger").
func GetIsolation(serverKey string) (IsolationMode, error) {
	cfg, err := Load()
	if err != nil {
		return "", err
	}
	if cfg.Servers == nil {
		return "", nil
	}
	return cfg.Servers[serverKey].Isolation, nil
}

// SetIsolation persists the isolation mode for a server key.
func SetIsolation(serverKey string, mode IsolationMode) error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	if cfg.Servers == nil {
		cfg.Servers = make(map[string]ServerConfig)
	}
	c := cfg.Servers[serverKey]
	c.Isolation = mode
	cfg.Servers[serverKey] = c
	return Save(cfg)
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configHome := firstNonEmpty(os.Getenv("XDG_CONFIG_HOME"), filepath.Join(home, ".config"))
	return filepath.Join(configHome, "garrison", "config.json"), nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
