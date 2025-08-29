package linuxgsm

import (
	"fmt"
	"strings"
)

// Validate checks the configuration for required properties and supported GameID
func (c Config) Validate() error {
	if strings.TrimSpace(c.ContainerName) == "" {
		return fmt.Errorf("%w: missing ContainerName", ErrInvalidConfig)
	}
	if c.GameID == "" {
		return fmt.Errorf("%w: missing GameID", ErrInvalidConfig)
	}
	if !IsSupportedGame(c.GameID) {
		return fmt.Errorf("%w: %s", ErrUnsupportedGame, c.GameID)
	}
	if len(c.Volumes) == 0 {
		// strongly encourage data persistence
		return fmt.Errorf("%w: at least one volume bind (e.g., /host/path:/data) required", ErrInvalidConfig)
	}
	// quick sanity check for /data bind
	hasData := false
	for _, b := range c.Volumes {
		if strings.Contains(b, ":/data") || strings.HasSuffix(b, ":/data:rw") || strings.HasSuffix(b, ":/data:ro") {
			hasData = true
			break
		}
	}
	if !hasData {
		return fmt.Errorf("%w: a bind mount to /data is required", ErrInvalidConfig)
	}
	return nil
}
