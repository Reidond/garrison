// Package steamcmd contains helpers to invoke SteamCMD for installing and
// updating dedicated server binaries.
// NOTE: no shell wrapping; path passed directly as argument to steamcmd
package steamcmd

import (
	"fmt"
	"os"
	"strings"

	execu "github.com/example/garrison/internal/executil"
)

// SteamCmdBinDefault is the default binary name searched on PATH.
const SteamCmdBinDefault = "steamcmd"

// Run executes SteamCMD to install or update an app into installDir.
func Run(installDir string, appID string, validate bool) error {
	if err := os.MkdirAll(installDir, 0o755); err != nil {
		return err
	}
	args := BuildArgs(installDir, appID, validate)
	return execu.Default.Run(steamcmdBin(), args...)
}

func steamcmdBin() string {
	if b := os.Getenv("GARRISON_STEAMCMD_BIN"); strings.TrimSpace(b) != "" {
		return b
	}
	return SteamCmdBinDefault
}

// BuildArgs returns the recommended argument list for a non-interactive steamcmd run.
func BuildArgs(installDir string, appID string, validate bool) []string {
	args := []string{
		"+force_install_dir", installDir,
		"+login", "anonymous",
		"+app_update", appID,
	}
	if validate {
		args = append(args, "+validate")
	}
	args = append(args, "+quit")
	return args
}

// BuildRunscriptContent returns a non-interactive steamcmd runscript.
func BuildRunscriptContent(appID string, validate bool, containerInstallDir string) string {
	validateSuffix := ""
	if validate {
		validateSuffix = " validate"
	}
	// containerInstallDir is the path inside steamcmd context (e.g., /data/app)
	return fmt.Sprintf(`@ShutdownOnFailedCommand 1
@NoPromptForPassword 1
force_install_dir %s
login anonymous
app_update %s%s
quit`, containerInstallDir, appID, validateSuffix)
}
