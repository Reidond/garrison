package utils

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// ExpandPath expands ~ to user home and returns an absolute path.
func ExpandPath(p string) (string, error) {
	if p == "" {
		return "", errors.New("empty path")
	}
	if strings.HasPrefix(p, "~/") || p == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		p = filepath.Join(home, strings.TrimPrefix(p, "~/"))
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	return abs, nil
}

// EnsureDir makes sure a directory exists with proper permissions.
func EnsureDir(dir string, perm os.FileMode) error {
	if dir == "" {
		return errors.New("empty dir")
	}
	if err := os.MkdirAll(dir, perm); err != nil {
		return err
	}
	return nil
}
