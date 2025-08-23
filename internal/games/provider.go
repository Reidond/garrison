package games

import (
	"context"
	"fmt"
	"time"
)

// ScenarioOptions captures common flags for listing scenarios/maps.
type ScenarioOptions struct {
	Binary     string
	InstallDir string
	Timeout    time.Duration
	Raw        bool
}

// Provider defines common helpers a game integration can implement.
type Provider interface {
	// ID is a short machine id like "reforger".
	ID() string
	// Name is a human-friendly name.
	Name() string
	// ListScenarios returns parsed scenario identifiers and the raw output captured.
	ListScenarios(ctx context.Context, opts ScenarioOptions) (scenarios []string, raw string, err error)
	// TemplateYAML returns a ready-made YAML game config template.
	TemplateYAML() []byte
	// InitConfig writes a native game config (e.g., JSON) to the given path with provided options.
	// Providers that don't support scaffolding should return ErrNotSupported.
	InitConfig(ctx context.Context, outputPath string, options map[string]string) error
}

// ErrNotSupported is returned when a provider doesn't implement a capability.
var ErrNotSupported = fmt.Errorf("not supported")
