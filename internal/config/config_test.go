package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLoadSaveIsolation(t *testing.T) {
	base := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(base, "cfg"))
	t.Setenv("HOME", base)

	if cfg, err := Load(); err != nil {
		t.Fatalf("Load error: %v", err)
	} else if len(cfg.Servers) != 0 {
		t.Fatalf("expected empty config")
	}

	if err := SetIsolation("srv", IsolationSystemd); err != nil {
		t.Fatalf("SetIsolation: %v", err)
	}
	m, err := GetIsolation("srv")
	if err != nil {
		t.Fatalf("GetIsolation: %v", err)
	}
	if m != IsolationSystemd {
		t.Fatalf("got %q want %q", m, IsolationSystemd)
	}

	if err := DeleteServer("srv"); err != nil {
		t.Fatalf("DeleteServer: %v", err)
	}
	cfgPath, err := configPath()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(cfgPath); err == nil {
		cfg, err := Load()
		if err != nil {
			t.Fatal(err)
		}
		if len(cfg.Servers) != 0 {
			t.Fatalf("expected servers empty post-delete")
		}
	}
}
