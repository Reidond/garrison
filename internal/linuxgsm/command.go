package linuxgsm

import (
	"bytes"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/container"
)

// ExecCommand executes a LinuxGSM subcommand inside the container and returns combined output
// e.g., ExecCommand("details") will run: ./<gameid>server details
func (gs *GameServer) ExecCommand(subcommand string) (string, error) {
	if gs.ID == "" {
		return "", ErrContainerNotFound
	}
	bin, err := ServerBinaryName(gs.config.GameID)
	if err != nil {
		return "", err
	}
	cmd := []string{"/bin/bash", "-lc", "./" + bin + " " + subcommand}

	execCfg := container.ExecOptions{
		User:         "linuxgsm",
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  false,
		Tty:          false,
		Cmd:          cmd,
		WorkingDir:   "/home/linuxgsm",
	}
	execResp, err := gs.client.ContainerExecCreate(gs.ctx, gs.ID, execCfg)
	if err != nil {
		return "", fmt.Errorf("exec create: %w", err)
	}
	if execResp.ID == "" {
		return "", fmt.Errorf("%w: empty exec id", ErrExecFailed)
	}
	attach, err := gs.client.ContainerExecAttach(gs.ctx, execResp.ID, container.ExecAttachOptions{Tty: false})
	if err != nil {
		return "", fmt.Errorf("exec attach: %w", err)
	}
	defer attach.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, attach.Reader); err != nil {
		return "", fmt.Errorf("exec read: %w", err)
	}
	// Optionally check exit status
	insp, err := gs.client.ContainerExecInspect(gs.ctx, execResp.ID)
	if err == nil && insp.ExitCode != 0 {
		return buf.String(), fmt.Errorf("%w: exit code %d", ErrExecFailed, insp.ExitCode)
	}
	return buf.String(), nil
}

// ExecRaw runs an arbitrary command in the container as linuxgsm user.
func (gs *GameServer) ExecRaw(cmd []string, workdir string) (string, error) {
	if gs.ID == "" {
		return "", ErrContainerNotFound
	}
	if len(cmd) == 0 {
		return "", fmt.Errorf("empty command")
	}
	if workdir == "" {
		workdir = "/home/linuxgsm"
	}
	execCfg := container.ExecOptions{
		User:         "linuxgsm",
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmd,
		WorkingDir:   workdir,
	}
	created, err := gs.client.ContainerExecCreate(gs.ctx, gs.ID, execCfg)
	if err != nil {
		return "", err
	}
	att, err := gs.client.ContainerExecAttach(gs.ctx, created.ID, container.ExecAttachOptions{})
	if err != nil {
		return "", err
	}
	defer att.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, att.Reader); err != nil {
		return "", err
	}
	insp, err := gs.client.ContainerExecInspect(gs.ctx, created.ID)
	if err == nil && insp.ExitCode != 0 {
		return buf.String(), fmt.Errorf("%w: exit code %d", ErrExecFailed, insp.ExitCode)
	}
	return buf.String(), nil
}
