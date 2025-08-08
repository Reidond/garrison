// Package steamcmd contains helpers to invoke SteamCMD for installing and
// updating dedicated server binaries.
package steamcmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// SteamCmdBinDefault is the default binary name searched on PATH.
const SteamCmdBinDefault = "steamcmd"

// Run executes SteamCMD to install or update an app into installDir.
func Run(installDir string, appID string, validate bool) error {
	if err := os.MkdirAll(installDir, 0o755); err != nil {
		return err
	}
	validateFlag := ""
	if validate {
		validateFlag = " validate"
	}
	cmdline := fmt.Sprintf("%s +force_install_dir %s +login anonymous +app_update %s%s +quit", steamcmdBin(), shellQuote(installDir), appID, validateFlag)
	c := exec.Command("bash", "-lc", cmdline)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func steamcmdBin() string {
	if b := os.Getenv("GARRISON_STEAMCMD_BIN"); strings.TrimSpace(b) != "" {
		return b
	}
	return SteamCmdBinDefault
}

func shellQuote(p string) string {
	if strings.ContainsAny(p, " \t\n\"") {
		return fmt.Sprintf("\"%s\"", strings.ReplaceAll(p, "\"", "\\\""))
	}
	return p
}
