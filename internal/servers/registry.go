package servers

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	arma "github.com/example/garrison/internal/arma"
	cfgpkg "github.com/example/garrison/internal/config"
	"github.com/example/garrison/internal/isolations/containeriso"
)

// StartOptions captures cross-game options required to start a server.
type StartOptions struct {
	InstallDir  string
	ConfigPath  string
	ProfileDir  string
	Port        int
	QueryPort   int
	BrowserPort int
	Extra       string
}

// Implementation contains the concrete operations for a specific game server.
// All methods should print user-friendly output to stdout where applicable.
type Implementation struct {
	InstallOrUpdate func(installDir string, isolation cfgpkg.IsolationMode, validate bool) error
	Start           func(isolation cfgpkg.IsolationMode, opts StartOptions) error
	Stop            func(installDir string, isolation cfgpkg.IsolationMode) error
	Status          func(installDir string, isolation cfgpkg.IsolationMode) error
	Delete          func(installDir string, isolation cfgpkg.IsolationMode) error
}

var registry = map[string]Implementation{
	arma.ServerKey: {
		InstallOrUpdate: func(dir string, iso cfgpkg.IsolationMode, validate bool) error {
			switch iso {
			case cfgpkg.IsolationSystemd:
				return fmt.Errorf("systemd isolation removed; use --isolation=container or none")
			case cfgpkg.IsolationContainer:
				return containeriso.InstallOrUpdate(arma.ServerKey, arma.AppID, dir, validate)
			default:
				return arma.InstallOrUpdateDirect(dir, validate)
			}
		},
		Start: func(iso cfgpkg.IsolationMode, opts StartOptions) error {
			switch iso {
			case cfgpkg.IsolationSystemd:
				return fmt.Errorf("systemd isolation removed; use --isolation=container or none")
			case cfgpkg.IsolationContainer:
				cmd := []string{
					"/data/app/" + filepath.Base(arma.BinaryPath(opts.InstallDir)),
					"-config=" + strings.ReplaceAll(opts.ConfigPath, opts.InstallDir, "/data"),
					"-profile=" + strings.ReplaceAll(opts.ProfileDir, opts.InstallDir, "/data"),
					fmt.Sprintf("-port=%d", opts.Port),
					fmt.Sprintf("-queryPort=%d", opts.QueryPort),
					fmt.Sprintf("-steamQueryPort=%d", opts.QueryPort),
					fmt.Sprintf("-serverBrowserPort=%d", opts.BrowserPort),
				}
				if strings.TrimSpace(opts.Extra) != "" {
					cmd = append(cmd, opts.Extra)
				}
				ports := [][2]int{{opts.Port, opts.Port}, {opts.QueryPort, opts.QueryPort}, {opts.BrowserPort, opts.BrowserPort}}
				return containeriso.StartGeneric(arma.ServerKey, opts.InstallDir, cmd, ports)
			default:
				return arma.StartDirect(opts.InstallDir, opts.ConfigPath, opts.ProfileDir, opts.Port, opts.QueryPort, opts.BrowserPort, opts.Extra)
			}
		},
		Stop: func(dir string, iso cfgpkg.IsolationMode) error {
			switch iso {
			case cfgpkg.IsolationSystemd:
				return fmt.Errorf("systemd isolation removed; use --isolation=container or none")
			case cfgpkg.IsolationContainer:
				return containeriso.Stop(arma.ServerKey)
			default:
				return arma.StopDirect(dir)
			}
		},
		Status: func(dir string, iso cfgpkg.IsolationMode) error {
			switch iso {
			case cfgpkg.IsolationSystemd:
				return fmt.Errorf("systemd isolation removed; use --isolation=container or none")
			case cfgpkg.IsolationContainer:
				return containeriso.Status(arma.ServerKey)
			default:
				s, err := arma.StatusDirect(dir)
				if err != nil {
					return err
				}
				fmt.Println(s)
				return nil
			}
		},
		Delete: func(dir string, iso cfgpkg.IsolationMode) error {
			switch iso {
			case cfgpkg.IsolationSystemd:
				return fmt.Errorf("systemd isolation removed; use --isolation=container or none")
			case cfgpkg.IsolationContainer:
				return containeriso.Delete(arma.ServerKey, dir)
			default:
				if dir == "" {
					return fmt.Errorf("install-dir required to delete in non-systemd mode")
				}
				return os.RemoveAll(dir)
			}
		},
	},
}

// Get returns the Implementation for a given server key.
func Get(serverKey string) (Implementation, error) {
	impl, ok := registry[serverKey]
	if !ok {
		return Implementation{}, fmt.Errorf("unknown server: %s", serverKey)
	}
	return impl, nil
}

// Keys returns a sorted list of available server keys.
func Keys() []string {
	keys := make([]string, 0, len(registry))
	for k := range registry {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
