package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	cfgpkg "github.com/example/garrison/internal/config"
	"github.com/example/garrison/internal/servers"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show server status",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverKey, _ := cmd.Flags().GetString("server")
		if serverKey == "" {
			return fmt.Errorf("--server is required. Available: %v", servers.Keys())
		}
		installDir, _ := cmd.Flags().GetString("install-dir")
		isoFlag, _ := cmd.Flags().GetString("isolation")
		var mode cfgpkg.IsolationMode
		switch isoFlag {
		case "none", "":
			if m, err := cfgpkg.GetIsolation(serverKey); err == nil && m != "" {
				mode = m
			} else {
				mode = cfgpkg.IsolationNone
			}
		case "systemd":
			return fmt.Errorf("systemd isolation removed; use --isolation=container or none")
		case "container":
			mode = cfgpkg.IsolationContainer
		default:
			return fmt.Errorf("unknown --isolation=%s", isoFlag)
		}
		impl, err := servers.Get(serverKey)
		if err != nil {
			return err
		}
		return impl.Status(installDir, mode)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().String("server", "", "Server key (e.g. arma-reforger)")
	statusCmd.Flags().StringP("install-dir", "d", "./servers/arma-reforger", "Install directory for server files")
	statusCmd.Flags().String("isolation", "", "Process/files isolation: none|systemd (defaults from saved config if empty)")
}
