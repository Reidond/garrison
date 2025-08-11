package servers

import (
	"fmt"
	"os"
	"sort"

	arma "github.com/example/garrison/internal/arma"
	cfgpkg "github.com/example/garrison/internal/config"
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
				return arma.InstallOrUpdateSystemd(validate)
			default:
				return arma.InstallOrUpdateDirect(dir, validate)
			}
		},
		Start: func(iso cfgpkg.IsolationMode, opts StartOptions) error {
			switch iso {
			case cfgpkg.IsolationSystemd:
				return arma.StartSystemd(opts.Port, opts.QueryPort, opts.BrowserPort, opts.Extra)
			default:
				return arma.StartDirect(opts.InstallDir, opts.ConfigPath, opts.ProfileDir, opts.Port, opts.QueryPort, opts.BrowserPort, opts.Extra)
			}
		},
		Stop: func(dir string, iso cfgpkg.IsolationMode) error {
			switch iso {
			case cfgpkg.IsolationSystemd:
				return arma.StopSystemd()
			default:
				return arma.StopDirect(dir)
			}
		},
		Status: func(dir string, iso cfgpkg.IsolationMode) error {
			switch iso {
			case cfgpkg.IsolationSystemd:
				return arma.StatusSystemd()
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
				return arma.DeleteAllSystemd()
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
