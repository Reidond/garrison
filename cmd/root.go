package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "garrison",
	Short: "Garrison - game server manager & monitor",
	Long:  "Garrison manages SteamCMD-backed game servers: install/update/start/stop/status.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
