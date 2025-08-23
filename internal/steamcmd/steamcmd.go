package steamcmd

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
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

// StreamingRunner can stream stdout/stderr while the process is running.
// Implementations should invoke the provided callbacks with raw byte chunks as they arrive.
// The function returns when the command exits, yielding the exit code and any error.
type StreamingRunner interface {
	RunStreaming(ctx context.Context, bin string, args []string, onStdout func([]byte), onStderr func([]byte)) (exitCode int, err error)
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

func (localRunner) RunStreaming(ctx context.Context, bin string, args []string, onStdout func([]byte), onStderr func([]byte)) (int, error) {
	cmd := exec.CommandContext(ctx, bin, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 0, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return 0, err
	}
	if err := cmd.Start(); err != nil {
		return 0, err
	}
	var wg sync.WaitGroup
	emit := func(r io.Reader, cb func([]byte)) {
		defer wg.Done()
		if cb == nil {
			io.Copy(io.Discard, r)
			return
		}
		buf := make([]byte, 4096)
		var acc bytes.Buffer
		for {
			n, readErr := r.Read(buf)
			if n > 0 {
				// accumulate and emit complete lines
				acc.Write(buf[:n])
				for {
					b := acc.Bytes()
					idx := bytes.IndexByte(b, '\n')
					if idx < 0 {
						break
					}
					line := make([]byte, idx+1)
					copy(line, b[:idx+1])
					// consume
					acc.Next(idx + 1)
					cb(line)
				}
			}
			if readErr != nil {
				if readErr != io.EOF {
					// flush remainder on non-EOF too
				}
				// flush any remainder without newline
				if acc.Len() > 0 {
					cb(append(acc.Bytes(), '\n'))
					acc.Reset()
				}
				return
			}
		}
	}
	wg.Add(2)
	go emit(stdout, onStdout)
	go emit(stderr, onStderr)
	waitErr := cmd.Wait()
	wg.Wait()
	exitCode := 0
	if waitErr != nil {
		if ee, ok := waitErr.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		}
	} else if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}
	return exitCode, waitErr
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
		// If streaming is requested and supported, prefer streaming mode.
		if c.cfg.StreamOutput {
			if sr, ok := c.runner.(StreamingRunner); ok {
				var okFlag, failFlag bool
				classify := func(line string) {
					switch classifyLine(line) {
					case outcomeSuccess:
						okFlag = true
					case outcomeError:
						failFlag = true
					}
				}
				// Prepare line handlers for stdout/stderr
				handle := func(b []byte) {
					s := string(b)
					// emit to OnOutput without trailing newline
					if cb := c.cfg.OnOutput; cb != nil {
						text := strings.TrimRight(s, "\n")
						if p := c.cfg.OutputPrefix; p != "" {
							cb(p + text)
						} else {
							cb(text)
						}
					}
					// classify with original including newline
					classify(s)
				}
				exitCode, err := sr.RunStreaming(ctx, bin, args, handle, handle)
				if err == nil {
					if okFlag && !failFlag {
						return nil
					}
					if exitCode == 0 && !failFlag {
						return nil
					}
					lastErr = errors.New("steamcmd did not report success")
				} else {
					lastErr = err
				}
				time.Sleep(time.Duration((attempt+1)*backoff) * time.Millisecond)
				continue
			}
		}
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
