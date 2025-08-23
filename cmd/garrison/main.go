package main

import (
	"github.com/alecthomas/kong"
)

// entrypoint: parse CLI and dispatch command Run methods
func main() {
	var cli Root
	ctx := kong.Parse(&cli,
		kong.Name("garrison"),
		kong.Description("Garrison - manage dedicated Steam game servers (create, install, update, start, stop, monitor)."),
		kong.UsageOnError(),
		kong.Bind(&cli), // allow injecting *Root into commands' Run methods
	)

	// Initialize global context (logging, signals) before executing commands
	runCtx, cancel := cli.Init()
	defer cancel()

	// Execute selected command
	err := ctx.Run(runCtx)
	ctx.FatalIfErrorf(err)
}

// end
