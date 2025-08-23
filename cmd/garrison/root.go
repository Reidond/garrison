package main

import (
	"context"
	"time"

	"github.com/Reidond/garrison/internal/utils"
)

// Root defines global flags and subcommands for Kong
type Root struct {
	Config  string `help:"Path to configuration file" default:"~/.garrison/config.yaml" type:"path"`
	Verbose bool   `help:"Enable verbose logging" short:"v"`

	Create  CreateCmd  `cmd:"create" help:"Create a new server instance"`
	Install InstallCmd `cmd:"install" help:"Install a server via SteamCMD"`
	Update  UpdateCmd  `cmd:"update" help:"Update a server via SteamCMD"`
	Start   StartCmd   `cmd:"start" help:"Start a server"`
	Stop    StopCmd    `cmd:"stop" help:"Stop a server"`
	Restart RestartCmd `cmd:"restart" help:"Restart a server"`
	Monitor MonitorCmd `cmd:"monitor" help:"Monitor a server"`
	Details DetailsCmd `cmd:"details" help:"Show server details"`
	List    ListCmd    `cmd:"list" help:"List all managed servers"`
	Delete  DeleteCmd  `cmd:"delete" help:"Delete a server"`

	// internal runtime context
	runCtx context.Context    `kong:"-"`
	cancel context.CancelFunc `kong:"-"`
}

// Init initializes logging and signal handling, returning a context to run commands with.
func (r *Root) Init() (context.Context, context.CancelFunc) {
	utils.SetupLogger(r.Verbose)
	ctx, cancel := context.WithCancel(context.Background())
	// Attach signal cancellation
	utils.SetupSignalHandler(cancel)
	// Add a timeout safeguard for CLI operations to prevent hanging forever
	tctx, tcancel := context.WithTimeout(ctx, 24*time.Hour)
	r.runCtx = tctx
	r.cancel = func() { tcancel(); cancel() }
	return r.runCtx, r.cancel
}

// Context returns the runtime context for command execution.
func (r *Root) Context() context.Context { return r.runCtx }
