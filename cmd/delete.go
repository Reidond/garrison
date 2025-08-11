package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	cfgpkg "github.com/example/garrison/internal/config"
	"github.com/example/garrison/internal/servers"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete server files and/or systemd user unit",
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
		if err := impl.Delete(installDir, mode); err != nil {
			return err
		}
		_ = cfgpkg.DeleteServer(serverKey)
		fmt.Println("Deleted server artifacts")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().String("server", "", "Server key (e.g. arma-reforger)")
	deleteCmd.Flags().StringP("install-dir", "d", "./servers/arma-reforger", "Install directory for server files")
	deleteCmd.Flags().String("isolation", "", "Process/files isolation: none|systemd (defaults from saved config if empty)")
}
