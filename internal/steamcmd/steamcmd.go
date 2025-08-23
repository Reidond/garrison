package steamcmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

type Client struct {
	cfg Config
}

func New(cfg Config) *Client { return &Client{cfg: cfg} }

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
		cmd := exec.CommandContext(ctx, bin, args...)
		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()
		if err := cmd.Start(); err != nil {
			lastErr = err
			time.Sleep(time.Duration((attempt+1)*backoff) * time.Millisecond)
			continue
		}

		var ok bool
		var fail bool
		scan := func(rdr *bufio.Reader) {
			for {
				line, err := rdr.ReadString('\n')
				if len(line) > 0 {
					switch classifyLine(line) {
					case outcomeSuccess:
						ok = true
					case outcomeError:
						fail = true
					}
				}
				if err != nil {
					return
				}
			}
		}
		go scan(bufio.NewReader(stdout))
		go scan(bufio.NewReader(stderr))

		if err := cmd.Wait(); err != nil {
			lastErr = err
			time.Sleep(time.Duration((attempt+1)*backoff) * time.Millisecond)
			continue
		}
		if ok && !fail {
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
