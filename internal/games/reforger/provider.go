package reforger

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Reidond/garrison/internal/games"
)

// Provider implements games.Provider for Arma Reforger.
// It shells out to the server binary for scenario listing and serves a YAML template.

type Provider struct{}

func (Provider) ID() string   { return "reforger" }
func (Provider) Name() string { return "Arma Reforger" }

func init() {
	games.Register(Provider{})
}

func (Provider) ListScenarios(ctx context.Context, opts games.ScenarioOptions) ([]string, string, error) {
	// Resolve working dir and binary
	workDir := ""
	if opts.InstallDir != "" {
		workDir = opts.InstallDir
	}
	bin := opts.Binary
	if bin == "" {
		bin = "ArmaReforgerServer"
	}
	if workDir != "" && !filepath.IsAbs(bin) {
		candidate := filepath.Join(workDir, bin)
		if _, err := os.Stat(candidate); err == nil {
			bin = candidate
		} else {
			if lp, lerr := exec.LookPath(bin); lerr == nil {
				bin = lp
			} else if errors.Is(err, os.ErrNotExist) {
				return nil, "", fmt.Errorf("binary not found: %s (also not in PATH)", candidate)
			} else {
				return nil, "", fmt.Errorf("cannot access binary: %w", err)
			}
		}
	} else {
		if _, err := os.Stat(bin); err != nil {
			if lp, lerr := exec.LookPath(bin); lerr == nil {
				bin = lp
			} else if errors.Is(err, os.ErrNotExist) {
				return nil, "", fmt.Errorf("binary not found: %s (also not in PATH)", bin)
			} else {
				return nil, "", fmt.Errorf("cannot access binary: %w", err)
			}
		}
	}

	// Prepare context timeout
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, bin, "-listScenarios")
	if workDir != "" {
		cmd.Dir = workDir
	}
	var combined bytes.Buffer
	cmd.Stdout = &combined
	cmd.Stderr = &combined
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, combined.String(), fmt.Errorf("timed out after %s", timeout)
		}
		// Proceed with parsing whatever we got
	}
	out := combined.String()
	if opts.Raw {
		return nil, out, nil
	}

	// Parse for .conf occurrences
	re := regexp.MustCompile(`(?i)([A-Za-z0-9_./\\-]+\\.conf)`) // escape for Go string
	matches := re.FindAllStringSubmatch(out, -1)
	set := map[string]struct{}{}
	for _, m := range matches {
		if len(m) > 1 {
			set[m[1]] = struct{}{}
		} else if len(m) > 0 {
			set[m[0]] = struct{}{}
		}
	}
	if len(set) == 0 {
		// Fallback: return any lines mentioning scenario
		lines := []string{}
		sc := bufio.NewScanner(strings.NewReader(out))
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if strings.Contains(strings.ToLower(line), "scenario") {
				if line != "" {
					lines = append(lines, line)
				}
			}
		}
		return lines, out, nil
	}
	list := make([]string, 0, len(set))
	for k := range set {
		list = append(list, k)
	}
	sort.Strings(list)
	return list, out, nil
}

func (Provider) TemplateYAML() []byte {
	// Reuse the repository template, but return inline to avoid imports.
	return []byte(`# Example garrison game config for Arma Reforger server
app_id: 1874900
executable: "ArmaReforgerServer"
working_dir: "."
ports:
  - "2001/udp"
  - "17777/udp"
  - "19999/udp"
command_template: >-
  {{install_dir}}/ArmaReforgerServer
  -profile {{install_dir}}/profile
  -config {{install_dir}}/Configs/ServerConfig.json
  -maxFPS 120
  {{args}}
start_args: ""
`)
}

func (Provider) InitConfig(ctx context.Context, outputPath string, options map[string]string) error {
	// Expected options (stringly-typed): name, scenarioId, a2sPort, publicPort, rconPort, rconPassword, admins (comma-separated), fastValidation (true/false)
	get := func(k, def string) string {
		if v, ok := options[k]; ok {
			return v
		}
		return def
	}
	atoi := func(s string, def int) int {
		if s == "" {
			return def
		}
		if n, err := strconv.Atoi(s); err == nil {
			return n
		}
		return def
	}
	atob := func(s string, def bool) bool {
		if s == "" {
			return def
		}
		if v, err := strconv.ParseBool(s); err == nil {
			return v
		}
		return def
	}
	splitCSV := func(s string) []string {
		if strings.TrimSpace(s) == "" {
			return nil
		}
		parts := strings.Split(s, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
		return out
	}
	cfg := RootConfig{
		PublicPort: atoi(get("publicPort", "2001"), 2001),
		A2S:        &A2SConfig{Port: atoi(get("a2sPort", "17777"), 17777)},
		Game: GameConfig{
			Name:       get("name", "My Server"),
			Admins:     splitCSV(get("admins", "")),
			ScenarioID: get("scenarioId", ""),
			GameProperties: GameProperties{
				FastValidation: atob(get("fastValidation", "true"), true),
			},
		},
	}
	if pw := get("rconPassword", ""); pw != "" {
		cfg.RCON = &RCONConfig{Port: atoi(get("rconPort", "19999"), 19999), Password: pw}
	}
	return cfg.WriteJSON(outputPath)
}
