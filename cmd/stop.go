package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	cfgpkg "github.com/example/garrison/internal/config"
	"github.com/example/garrison/internal/servers"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a dedicated server",
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
			mode = cfgpkg.IsolationSystemd
		default:
			return fmt.Errorf("unknown --isolation=%s", isoFlag)
		}
		impl, err := servers.Get(serverKey)
		if err != nil {
			return err
		}
		if err := impl.Stop(installDir, mode); err != nil {
			return err
		}
		fmt.Println("Server stopped")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
	stopCmd.Flags().String("server", "", "Server key (e.g. arma-reforger)")
	stopCmd.Flags().StringP("install-dir", "d", "./servers/arma-reforger", "Install directory for server files")
	stopCmd.Flags().String("isolation", "", "Process/files isolation: none|systemd (defaults from saved config if empty)")
}
