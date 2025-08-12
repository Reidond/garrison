package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	cfgpkg "github.com/example/garrison/internal/config"
	"github.com/example/garrison/internal/servers"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a dedicated server",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverKey, _ := cmd.Flags().GetString("server")
		if serverKey == "" {
			return fmt.Errorf("--server is required. Available: %v", servers.Keys())
		}
		installDir, _ := cmd.Flags().GetString("install-dir")
		configPath, _ := cmd.Flags().GetString("config")
		profile, _ := cmd.Flags().GetString("profile")
		port, _ := cmd.Flags().GetInt("port")
		qport, _ := cmd.Flags().GetInt("query-port")
		bport, _ := cmd.Flags().GetInt("browser-port")
		extra, _ := cmd.Flags().GetString("extra")
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
		opts := servers.StartOptions{InstallDir: installDir, ConfigPath: configPath, ProfileDir: profile, Port: port, QueryPort: qport, BrowserPort: bport, Extra: extra}
		if err := impl.Start(mode, opts); err != nil {
			return err
		}
		fmt.Println("Server started")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().String("server", "", "Server key (e.g. arma-reforger)")
	startCmd.Flags().StringP("install-dir", "d", "./servers/arma-reforger", "Install directory for server files")
	startCmd.Flags().String("config", "./servers/arma-reforger/profiles/config.json", "Path to server config.json")
	startCmd.Flags().String("profile", "./servers/arma-reforger/profiles", "Profile directory for logs & saves")
	startCmd.Flags().Int("port", 20001, "Game port (UDP)")
	startCmd.Flags().Int("query-port", 27016, "Steam query port (UDP)")
	startCmd.Flags().Int("browser-port", 17777, "Server browser port (UDP)")
	startCmd.Flags().String("extra", "", "Extra args to pass to server binary")
	startCmd.Flags().String("isolation", "", "Process/files isolation: none|systemd (defaults from saved config if empty)")
}
