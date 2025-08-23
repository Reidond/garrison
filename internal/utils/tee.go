package utils

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// NewLineTee creates a line-oriented tee callback that writes each line to slog
// (with an optional prefix) and to the specified log file. It returns the callback
// and a close function to close the file when finished.
func NewLineTee(prefix, logFile string) (func(string), func() error, error) {
	if err := os.MkdirAll(filepath.Dir(logFile), 0o755); err != nil {
		return nil, nil, err
	}
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, err
	}
	mu := &sync.Mutex{}
	logger := slog.Default()
	cb := func(line string) {
		mu.Lock()
		defer mu.Unlock()
		// Slog
		if prefix != "" {
			logger.Info(prefix + line)
		} else {
			logger.Info(line)
		}
		// File (append newline)
		ts := time.Now().Format(time.RFC3339)
		_, _ = f.WriteString(fmt.Sprintf("%s %s%s\n", ts, prefix, line))
	}
	closer := func() error { return f.Close() }
	return cb, closer, nil
}
