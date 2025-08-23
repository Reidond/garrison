package server

// ProcessOptions controls how the server process is started.
// More fields (UID/GID, nice level, rlimits) can be added in the future.
type ProcessOptions struct {
	Env    []string // environment variables for the process (nil to inherit default)
	UID    int      // drop privileges to this user ID (Unix)
	GID    int      // drop privileges to this group ID (Unix)
	Groups []int    // supplementary groups (Unix)
}
