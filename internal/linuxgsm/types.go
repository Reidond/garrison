package linuxgsm

import "errors"

// GameID represents the short identifier used by LinuxGSM/server images
// e.g. "armar", "csgo", etc.
type GameID string

// ContainerStatus is a simplified container lifecycle state
type ContainerStatus string

const (
	StatusRunning ContainerStatus = "running"
	StatusExited  ContainerStatus = "exited"
	StatusCreated ContainerStatus = "created"
	StatusUnknown ContainerStatus = "unknown"
)

// Library common errors
var (
	ErrContainerNotFound = errors.New("container not found")
	ErrExecFailed        = errors.New("command execution failed")
	ErrUnsupportedGame   = errors.New("unsupported game id")
	ErrInvalidConfig     = errors.New("invalid config")
)

// ServerInfo describes a supported game server from serverlist.csv
type ServerInfo struct {
	ShortName      string // e.g. armar
	GameServerName string // e.g. armarserver
	GameName       string // e.g. Arma Reforger
	OS             string // e.g. ubuntu-24.04
}
