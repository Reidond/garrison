package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	cfgpkg "github.com/example/garrison/internal/config"
	"github.com/example/garrison/internal/servers"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a dedicated server via SteamCMD",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverKey, _ := cmd.Flags().GetString("server")
		if serverKey == "" {
			return fmt.Errorf("--server is required. Available: %v", servers.Keys())
		}
		installDir, _ := cmd.Flags().GetString("install-dir")
		isoFlag, _ := cmd.Flags().GetString("isolation")
		var mode cfgpkg.IsolationMode
		switch isoFlag {
		case "", "none":
			mode = cfgpkg.IsolationNone
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
		if err := impl.InstallOrUpdate(installDir, mode, true); err != nil {
			return err
		}
		_ = cfgpkg.SetIsolation(serverKey, mode)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().String("server", "", "Server key (e.g. arma-reforger)")
	installCmd.Flags().StringP("install-dir", "d", "./servers/arma-reforger", "Install directory for server files")
	installCmd.Flags().String("isolation", "none", "Process/files isolation: none|systemd")
}
