package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Reidond/garrison/internal/utils"
)

// Metadata for a managed server instance
type Metadata struct {
	Name       string `json:"name"`
	InstallDir string `json:"install_dir"`
	AppID      int    `json:"app_id"`
	ConfigPath string `json:"config_path"`
}

func dataDir() (string, error) {
	p, err := utils.ExpandPath("~/.garrison")
	if err != nil {
		return "", err
	}
	if err := utils.EnsureDir(p, 0o755); err != nil {
		return "", err
	}
	return p, nil
}

func metadataPath(name string) (string, error) {
	base, err := dataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, fmt.Sprintf("%s.json", name)), nil
}

// SaveMetadata writes metadata for a server name.
func SaveMetadata(md Metadata) error {
	if md.Name == "" {
		return errors.New("metadata name is empty")
	}
	p, err := metadataPath(md.Name)
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(md, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0o644)
}

// LoadMetadata reads metadata for a server name.
func LoadMetadata(name string) (Metadata, error) {
	var md Metadata
	p, err := metadataPath(name)
	if err != nil {
		return md, err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return md, err
	}
	if err := json.Unmarshal(b, &md); err != nil {
		return md, err
	}
	return md, nil
}

// ListMetadata lists all saved servers.
func ListMetadata() ([]Metadata, error) {
	base, err := dataDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil, err
	}
	out := make([]Metadata, 0)
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		b, err := os.ReadFile(filepath.Join(base, e.Name()))
		if err != nil {
			continue
		}
		var md Metadata
		if err := json.Unmarshal(b, &md); err == nil {
			out = append(out, md)
		}
	}
	return out, nil
}

// DeleteMetadata removes a server entry.
func DeleteMetadata(name string) error {
	p, err := metadataPath(name)
	if err != nil {
		return err
	}
	return os.Remove(p)
}
