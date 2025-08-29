# linuxgsm (internal package)

This internal package provides a small orchestration layer to run LinuxGSM-based game servers inside Docker containers.

It embeds LinuxGSM's server list (`serverlist.csv`) to validate game IDs and to resolve each game's LinuxGSM script/binary name for command execution.

> Note: This is an internal package and is not intended to be imported by external modules. The import path within this repo is `github.com/Reidond/garrison/internal/linuxgsm`.

## Features

- Parse and embed LinuxGSM's supported games list
- Validate configuration and enforce a `/data` bind mount
- Create / start / stop / remove containers
- Execute LinuxGSM commands in the container as the `linuxgsm` user
- Query simple container status

## Types overview

- `type GameID string`
- `type ContainerStatus string` with values: `running`, `created`, `exited`, `unknown`
- `type ServerInfo struct { ShortName, GameServerName, GameName, OS string }`
- `type Config struct { ContainerName, GameID, Image, Volumes, Network, RestartPolicy, Ports, EnvVars, PullAlways }`
- `func (c Config) ImageOrDefault() string`
- `type GameServer struct { ... }`

## Basic usage

```go
ctx := context.Background()
cli, err := linuxgsm.NewDockerClient()
if err != nil { /* handle */ }

cfg := linuxgsm.Config{
    ContainerName: "arma-reforger",
    GameID:        linuxgsm.GameID("armar"),
    Volumes:       []string{"/path/to/state:/data"},
    Network:       "host",
    RestartPolicy: "unless-stopped",
    PullAlways:    true,
}

srv, err := linuxgsm.NewGameServer(ctx, cli, cfg)
if err != nil { /* handle */ }

if err := srv.CreateContainer(); err != nil { /* handle */ }
if err := srv.StartContainer(); err != nil { /* handle */ }

out, err := srv.ExecCommand("details")
fmt.Println(out, err)

_ = srv.StopContainer(10 * time.Second)
_ = srv.RemoveContainer(true)
```

## Supported games

- `SupportedGames() map[GameID]ServerInfo`: returns a copy of the embedded map
- `IsSupportedGame(GameID) bool`
- `ServerBinaryName(GameID) (string, error)`

## Configuration notes

- `/data` bind: the validator requires at least one bind mount to `/data` inside the container since LinuxGSM installs and stores server data there.
- Image: if omitted, the default image is `gameservermanagers/gameserver:<GameID>`.
- Network: using `host` is convenient for servers that require many UDP/TCP ports. For bridge networks, configure port mappings.
- Environment: you can pass env vars like `TZ=UTC`.

## Error handling

- `ErrUnsupportedGame`, `ErrInvalidConfig`, `ErrContainerNotFound`, `ErrExecFailed` for common failure modes.

## Limitations and future work

- This package focuses on core lifecycle and command execution. It does not yet include:
  - log streaming APIs
  - event subscriptions
  - advanced health checks
  - concurrency primitives around multiple simultaneous operations

These can be added incrementally if needed.
