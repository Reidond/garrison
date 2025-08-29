package linuxgsm

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

// DockerClient is a thin wrapper around docker's client.Client
type DockerClient struct {
	*client.Client
}

// NewDockerClient creates a Docker client from environment and negotiates API version.
func NewDockerClient() (*DockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("docker client: %w", err)
	}
	return &DockerClient{Client: cli}, nil
}

// GameServer describes a concrete server container instance
type GameServer struct {
	ctx    context.Context
	client *DockerClient
	config Config
	ID     string // container ID after creation/lookup
}

// NewGameServer validates config and returns a GameServer handle
func NewGameServer(ctx context.Context, cli *DockerClient, cfg Config) (*GameServer, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &GameServer{ctx: ctx, client: cli, config: cfg}, nil
}

// toHostConfig builds host config from Config
func (gs *GameServer) toHostConfig() *container.HostConfig {
	hc := &container.HostConfig{}
	// volumes
	for _, b := range gs.config.Volumes {
		// docker supports simple bind notation through Mounts as well
		hc.Mounts = append(hc.Mounts, mount.Mount{Type: mount.TypeBind, Source: beforeColon(b), Target: afterColon(b)})
	}
	// restart policy
	if gs.config.RestartPolicy != "" {
		hc.RestartPolicy = container.RestartPolicy{Name: container.RestartPolicyMode(gs.config.RestartPolicy)}
	}
	// ports are managed via PortBindings in a complete impl; for simplicity, rely on network host or image defaults
	return hc
}

// toNetworkConfig builds network settings; host network when specified
func (gs *GameServer) toNetworkConfig() (*network.NetworkingConfig, *container.NetworkMode) {
	mode := container.NetworkMode("")
	if gs.config.Network != "" {
		mode = container.NetworkMode(gs.config.Network)
	}
	return &network.NetworkingConfig{}, &mode
}

// helpers for parsing binds in simple "src:dst[:opts]" form
func beforeColon(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			return s[:i]
		}
	}
	return s
}
func afterColon(s string) string {
	// return middle segment between first and second colon, or after first colon
	first := -1
	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			first = i
			break
		}
	}
	if first == -1 {
		return ""
	}
	// find second colon
	second := -1
	for i := first + 1; i < len(s); i++ {
		if s[i] == ':' {
			second = i
			break
		}
	}
	if second == -1 {
		return s[first+1:]
	}
	return s[first+1 : second]
}
