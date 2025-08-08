package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	arma "github.com/example/garrison/internal/arma"
	cfgpkg "github.com/example/garrison/internal/config"
	scmd "github.com/example/garrison/internal/steamcmd"
	sysd "github.com/example/garrison/internal/systemd"
)

const (
	steamCmdBinDefault = scmd.SteamCmdBinDefault
	armaReforgerAppID  = arma.AppID
)

func init() {
	rootCmd.AddCommand(armaReforgerCmd)
	armaReforgerCmd.AddCommand(installCmd)
	armaReforgerCmd.AddCommand(updateCmd)
	armaReforgerCmd.AddCommand(startCmd)
	armaReforgerCmd.AddCommand(stopCmd)
	armaReforgerCmd.AddCommand(statusCmd)

	startCmd.Flags().StringP("install-dir", "d", "./servers/arma-reforger", "Install directory for server files")
	startCmd.Flags().String("config", "./servers/arma-reforger/profiles/config.json", "Path to server config.json")
	startCmd.Flags().String("profile", "./servers/arma-reforger/profiles", "Profile directory for logs & saves")
	startCmd.Flags().Int("port", 20001, "Game port (UDP)")
	startCmd.Flags().Int("query-port", 27016, "Steam query port (UDP)")
	startCmd.Flags().Int("browser-port", 17777, "Server browser port (UDP)")
	startCmd.Flags().String("extra", "", "Extra args to pass to server binary")
	startCmd.Flags().String("isolation", "none", "Process/files isolation: none|systemd")

	stopCmd.Flags().StringP("install-dir", "d", "./servers/arma-reforger", "Install directory for server files")
	stopCmd.Flags().String("isolation", "none", "Process/files isolation: none|systemd")
	statusCmd.Flags().StringP("install-dir", "d", "./servers/arma-reforger", "Install directory for server files")
	statusCmd.Flags().String("isolation", "none", "Process/files isolation: none|systemd")

	installCmd.Flags().StringP("install-dir", "d", "./servers/arma-reforger", "Install directory for server files")
	installCmd.Flags().String("isolation", "none", "Process/files isolation: none|systemd")
	updateCmd.Flags().StringP("install-dir", "d", "./servers/arma-reforger", "Install directory for server files")
	updateCmd.Flags().String("isolation", "none", "Process/files isolation: none|systemd")
}

var armaReforgerCmd = &cobra.Command{
	Use:   "arma-reforger",
	Short: "Manage Arma Reforger dedicated server",
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Arma Reforger dedicated server via SteamCMD",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("install-dir")
		isolation, _ := cmd.Flags().GetString("isolation")
		switch isolation {
		case "none", "":
			if err := runSteamCmd(dir, true); err != nil {
				return err
			}
			// Persist choice
			_ = cfgpkg.SetIsolation("arma-reforger", cfgpkg.IsolationNone)
			return nil
		case "systemd":
			// Run steamcmd under systemd transient unit inside the sandbox
			if err := runSteamCmdSystemd(true); err != nil {
				return err
			}
			_ = cfgpkg.SetIsolation("arma-reforger", cfgpkg.IsolationSystemd)
			return nil
		default:
			return fmt.Errorf("unknown --isolation=%s", isolation)
		}
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Arma Reforger server via SteamCMD",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("install-dir")
		_ = dir // kept for flag compatibility (ignored with systemd isolation)
		isolation, _ := cmd.Flags().GetString("isolation")
		if isolation == "" {
			if mode, err := cfgpkg.GetIsolation("arma-reforger"); err == nil && mode != "" {
				isolation = string(mode)
			}
		}
		switch isolation {
		case "none", "":
			return runSteamCmd(dir, false)
		case "systemd":
			return runSteamCmdSystemd(false)
		default:
			return fmt.Errorf("unknown --isolation=%s", isolation)
		}
	},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Arma Reforger server",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("install-dir")
		config, _ := cmd.Flags().GetString("config")
		profile, _ := cmd.Flags().GetString("profile")
		port, _ := cmd.Flags().GetInt("port")
		qport, _ := cmd.Flags().GetInt("query-port")
		bport, _ := cmd.Flags().GetInt("browser-port")
		extra, _ := cmd.Flags().GetString("extra")
		isolation, _ := cmd.Flags().GetString("isolation")
		if isolation == "" {
			if mode, err := cfgpkg.GetIsolation("arma-reforger"); err == nil && mode != "" {
				isolation = string(mode)
			}
		}

		if isolation == "systemd" {
			return sysd.InstallAndStartUnit(armaReforgerAppID, port, qport, bport, extra)
		}
		if err := arma.StartDirect(dir, config, profile, port, qport, bport, extra); err != nil {
			return err
		}
		fmt.Println("Arma Reforger server started")
		return nil
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop Arma Reforger server",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("install-dir")
		isolation, _ := cmd.Flags().GetString("isolation")
		if isolation == "" {
			if mode, err := cfgpkg.GetIsolation("arma-reforger"); err == nil && mode != "" {
				isolation = string(mode)
			}
		}
		if isolation == "systemd" {
			return sysd.Stop()
		}
		if err := arma.StopDirect(dir); err != nil {
			return err
		}
		fmt.Println("Arma Reforger server stopped")
		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show Arma Reforger server status",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("install-dir")
		isolation, _ := cmd.Flags().GetString("isolation")
		if isolation == "" {
			if mode, err := cfgpkg.GetIsolation("arma-reforger"); err == nil && mode != "" {
				isolation = string(mode)
			}
		}
		if isolation == "systemd" {
			return sysd.Status()
		}
		status, err := arma.StatusDirect(dir)
		if err != nil {
			return err
		}
		fmt.Println(status)
		return nil
	},
}

func runSteamCmd(installDir string, validate bool) error {
	return scmd.Run(installDir, armaReforgerAppID, validate)
}

// Systemd helpers now live under internal/systemd
func runSteamCmdSystemd(validate bool) error {
	return sysd.RunSteamcmdTransient(armaReforgerAppID, validate)
}
