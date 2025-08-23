package utils

import (
	"log/slog"
	"os"
	"time"
)

// SetupLogger configures the global slog logger with readable time, source, and level.
func SetupLogger(verbose bool) {
	level := slog.LevelInfo
	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey && a.Value.Kind() == slog.KindTime {
				t := a.Value.Time()
				a.Value = slog.StringValue(t.Format(time.RFC3339))
			}
			return a
		},
	}
	if verbose {
		level = slog.LevelDebug
		opts.AddSource = true
	}
	handler := slog.NewTextHandler(os.Stderr, opts)

	slog.SetDefault(slog.New(handler))
}

// LoggerFor returns a namespaced logger with a component field.
func LoggerFor(component string) *slog.Logger {
	return slog.Default().With("component", component)
}
