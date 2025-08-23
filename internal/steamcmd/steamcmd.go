package steamcmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Client struct {
	cfg    Config
	runner Runner
}

// Runner abstracts how steamcmd is executed. Implementations may run locally or in a container.
type Runner interface {
	// Run executes the binary with args and returns combined output and exit code.
	Run(ctx context.Context, bin string, args ...string) (output string, exitCode int, err error)
}

// localRunner executes processes on the host.
type localRunner struct{}

func (localRunner) Run(ctx context.Context, bin string, args ...string) (string, int, error) {
	cmd := exec.CommandContext(ctx, bin, args...)
	out, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		// Try to extract exit code; if unavailable, leave as 0 and return err.
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		}
	} else if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}
	return string(out), exitCode, err
}

// New creates a client using the local process runner.
func New(cfg Config) *Client { return &Client{cfg: cfg, runner: localRunner{}} }

// NewWithRunner creates a client using a custom runner (e.g., for integration tests).
func NewWithRunner(cfg Config, r Runner) *Client { return &Client{cfg: cfg, runner: r} }

func (c *Client) exec(ctx context.Context, args ...string) error {
	bin := c.cfg.BinaryPath
	if bin == "" {
		bin = "steamcmd"
	}
	var lastErr error
	retries := c.cfg.Retries
	if retries <= 0 {
		retries = 3
	}
	backoff := c.cfg.Backoff
	if backoff <= 0 {
		backoff = 1000
	}
	for attempt := 0; attempt < retries; attempt++ {
		out, exitCode, err := c.runner.Run(ctx, bin, args...)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration((attempt+1)*backoff) * time.Millisecond)
			continue
		}
		// Parse combined output for success/error signals.
		var ok bool
		var fail bool
		rdr := bufio.NewReader(strings.NewReader(out))
		for {
			line, rerr := rdr.ReadString('\n')
			if len(line) > 0 {
				// emit to OnOutput if configured
				if cb := c.cfg.OnOutput; cb != nil {
					if p := c.cfg.OutputPrefix; p != "" {
						cb(p + strings.TrimRight(line, "\n"))
					} else {
						cb(strings.TrimRight(line, "\n"))
					}
				}
				switch classifyLine(line) {
				case outcomeSuccess:
					ok = true
				case outcomeError:
					fail = true
				}
			}
			if rerr != nil {
				break
			}
		}
		if ok && !fail {
			return nil
		}
		// Treat exit code 0 with no explicit failure as success, because some commands
		// (e.g., simple login +quit) may not print canonical "success" tokens.
		if exitCode == 0 && !fail {
			return nil
		}
		lastErr = errors.New("steamcmd did not report success")
		time.Sleep(time.Duration((attempt+1)*backoff) * time.Millisecond)
	}
	if lastErr == nil {
		lastErr = errors.New("steamcmd failed without error detail")
	}
	return lastErr
}

func (c *Client) loginArgs() []string {
	if c.cfg.Username != "" {
		if c.cfg.Password != "" {
			return []string{"+login", c.cfg.Username, c.cfg.Password}
		}
		return []string{"+login", c.cfg.Username}
	}
	return []string{"+login", "anonymous"}
}

// InstallApp installs a Steam app into the given directory.
func (c *Client) InstallApp(ctx context.Context, appID int, installDir string, validate bool) error {
	args := []string{}
	args = append(args, c.loginArgs()...)
	args = append(args, "+force_install_dir", installDir)
	args = append(args, "+app_update", fmt.Sprintf("%d", appID))
	if validate {
		args = append(args, "validate")
	}
	args = append(args, "+quit")
	return c.exec(ctx, args...)
}

// UpdateApp updates a Steam app in the given directory.
func (c *Client) UpdateApp(ctx context.Context, appID int, installDir string) error {
	return c.InstallApp(ctx, appID, installDir, false)
}

// PingLogin attempts a simple steamcmd login (anonymous or user) and quits.
// Useful for integration tests and connectivity checks without downloading apps.
func (c *Client) PingLogin(ctx context.Context) error {
	args := append([]string{}, c.loginArgs()...)
	args = append(args, "+quit")
	return c.exec(ctx, args...)
}
