package linuxgsm

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	imgt "github.com/docker/docker/api/types/image"
)

// CreateContainer pulls the image (optionally), creates the container and stores container ID.
func (gs *GameServer) CreateContainer() error {
	img := gs.config.ImageOrDefault()
	if img == "" {
		return fmt.Errorf("%w: missing image and GameID", ErrInvalidConfig)
	}
	// pull when requested or when image missing; for simplicity, pull when PullAlways is true
	if gs.config.PullAlways {
		rc, err := gs.client.ImagePull(gs.ctx, img, imgt.PullOptions{})
		if err != nil {
			return fmt.Errorf("pull image: %w", err)
		}
		io.Copy(io.Discard, rc)
		rc.Close()
	}

	env := append([]string{}, gs.config.EnvVars...)

	hostCfg := gs.toHostConfig()
	netCfg, netMode := gs.toNetworkConfig()
	if netMode != nil && *netMode != container.NetworkMode("") {
		hostCfg.NetworkMode = *netMode
	}

	// configure container create
	created, err := gs.client.ContainerCreate(
		gs.ctx,
		&container.Config{
			Image:    img,
			Hostname: gs.config.ContainerName,
			Env:      env,
			Tty:      false,
			// Always run as linuxgsm user in exec; container default user may be root depending on image
			// Cmd left empty to use image default entrypoint (LinuxGSM startup script)
			Labels: map[string]string{
				"garrison.module": "linuxgsm",
				"garrison.gameid": string(gs.config.GameID),
			},
		},
		hostCfg,
		netCfg,
		nil,
		gs.config.ContainerName,
	)
	if err != nil {
		return fmt.Errorf("container create: %w", err)
	}
	gs.ID = created.ID
	return nil
}

// StartContainer starts the container if created
func (gs *GameServer) StartContainer() error {
	if gs.ID == "" {
		return fmt.Errorf("%w: container not created", ErrContainerNotFound)
	}
	if err := gs.client.ContainerStart(gs.ctx, gs.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("container start: %w", err)
	}
	return nil
}

// StopContainer gracefully stops the container within timeout
func (gs *GameServer) StopContainer(timeout time.Duration) error {
	if gs.ID == "" {
		return fmt.Errorf("%w: container not created", ErrContainerNotFound)
	}
	to := container.StopOptions{Timeout: ptr(int(timeout.Seconds()))}
	if err := gs.client.ContainerStop(gs.ctx, gs.ID, to); err != nil {
		return fmt.Errorf("container stop: %w", err)
	}
	return nil
}

// RemoveContainer removes the container (force=false by default)
func (gs *GameServer) RemoveContainer(force bool) error {
	if gs.ID == "" {
		return fmt.Errorf("%w: container not created", ErrContainerNotFound)
	}
	if err := gs.client.ContainerRemove(gs.ctx, gs.ID, container.RemoveOptions{Force: force}); err != nil {
		return fmt.Errorf("container remove: %w", err)
	}
	gs.ID = ""
	return nil
}

// GetStatus inspects container and returns a simplified status
func (gs *GameServer) GetStatus() (ContainerStatus, error) {
	if gs.ID == "" {
		return StatusUnknown, ErrContainerNotFound
	}
	insp, err := gs.client.ContainerInspect(gs.ctx, gs.ID)
	if err != nil {
		if clientNotFound(err) {
			return StatusUnknown, ErrContainerNotFound
		}
		return StatusUnknown, fmt.Errorf("container inspect: %w", err)
	}
	if insp.ContainerJSONBase == nil || insp.State == nil {
		return StatusUnknown, nil
	}
	switch {
	case insp.State.Running:
		return StatusRunning, nil
	case insp.State.Status == "created":
		return StatusCreated, nil
	default:
		return StatusExited, nil
	}
}

func clientNotFound(err error) bool {
	// duck-type docker not found errors by message
	if err == nil {
		return false
	}
	s := err.Error()
	return s != "" && (strings.Contains(s, "No such container") || strings.Contains(s, "not found"))
}

func ptr[T any](v T) *T { return &v }
