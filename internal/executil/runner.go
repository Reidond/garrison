package executil

import (
	"os"
	"os/exec"
)

// Runner abstracts command execution for production and tests.
type Runner interface {
	Run(name string, args ...string) error
	CombinedOutput(name string, args ...string) (string, error)
}

// HostRunner executes commands on the host using os/exec.
type HostRunner struct{}

func (HostRunner) Run(name string, args ...string) error {
	c := exec.Command(name, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func (HostRunner) CombinedOutput(name string, args ...string) (string, error) {
	c := exec.Command(name, args...)
	b, err := c.CombinedOutput()
	return string(b), err
}

// Default is the default runner, initialised to HostRunner.
var Default Runner = HostRunner{}

// SetDefault replaces the default runner (useful in tests).
func SetDefault(r Runner) { Default = r }
