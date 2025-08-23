package steamcmd

import "strings"

type outcome int

const (
	outcomeUnknown outcome = iota
	outcomeSuccess
	outcomeError
)

// Very small helpers to detect success and errors from steamcmd output.
func classifyLine(s string) outcome {
	l := strings.ToLower(s)
	switch {
	case strings.Contains(l, "success"), strings.Contains(l, "download complete"), strings.Contains(l, "update complete"), strings.Contains(l, "loading steam api... ok"):
		return outcomeSuccess
	case strings.Contains(l, "error"), strings.Contains(l, "failed"), strings.Contains(l, "invalid"), strings.Contains(l, "timeout"):
		return outcomeError
	default:
		return outcomeUnknown
	}
}
