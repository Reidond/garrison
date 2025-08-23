package utils

import (
	"os"

	"golang.org/x/term"
)

// IsTerminal reports whether the given file descriptor refers to a terminal.
func IsTerminal(f *os.File) bool {
	if f == nil {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

// DefaultStderrIsTTY returns true if os.Stderr is a terminal.
func DefaultStderrIsTTY() bool { return IsTerminal(os.Stderr) }

// LineWriter is a utility to turn arbitrary writes into line-based callbacks.
// Useful to feed logs where consumers want complete lines.
type LineWriter struct {
	buf   []byte
	Flush func(line string)
}

func (w *LineWriter) Write(p []byte) (int, error) {
	w.buf = append(w.buf, p...)
	for {
		i := -1
		for idx, b := range w.buf {
			if b == '\n' {
				i = idx
				break
			}
		}
		if i < 0 {
			break
		}
		line := string(w.buf[:i])
		w.buf = w.buf[i+1:]
		if w.Flush != nil {
			w.Flush(line)
		}
	}
	return len(p), nil
}

// FlushRemainder emits any remaining buffered bytes as a line.
func (w *LineWriter) FlushRemainder() {
	if len(w.buf) > 0 {
		if w.Flush != nil {
			w.Flush(string(w.buf))
		}
		w.buf = nil
	}
}
