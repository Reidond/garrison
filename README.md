# Garrison - Game server manager

Garrison is a Go-based CLI for installing, updating, starting, stopping and checking status of SteamCMD-managed dedicated servers.

## Features
- Install/update via SteamCMD
  - Direct mode: runs `steamcmd` locally with `+force_install_dir`
  - Isolated mode: runs `steamcmd` inside a sandboxed systemd transient unit (`DynamicUser`, restricted FS)
- Start/stop/status
  - Direct mode: spawns the server binary; writes pid and logs under the chosen install/profile directory
  - Isolated mode: generates and manages a hardened systemd unit with `DynamicUser` and strong sandboxing
- Hardened sandbox (systemd)
  - `DynamicUser=yes`, `StateDirectory`, `CacheDirectory`, `LogsDirectory`
  - `ProtectSystem=strict`, `ProtectHome=yes`, `PrivateTmp=yes`, `NoNewPrivileges=yes`, `Restrict*`

## Currently supported
- Arma Reforger dedicated server (Steam App ID 1874900)

## Requirements
- Linux host with required UDP ports open (default: 20001, 17777, 27016)
- SteamCMD on PATH as `steamcmd` (for direct mode)
- systemd (for isolated mode) and sudo/root to install/manage the unit

## CLI

Commands are grouped under `arma-reforger`:
- `install` and `update`: install/update server files via SteamCMD
- `start`, `stop`, `status`: manage the server process

Common flags:
- `-d, --install-dir`: target directory (direct mode)
- `--config`, `--profile`, `--port`, `--query-port`, `--browser-port`, `--extra`
- `--isolation`: `none` (default) or `systemd`

### Examples (direct mode)
- Install/update:
  garrison arma-reforger install -d ./servers/arma-reforger
  garrison arma-reforger update -d ./servers/arma-reforger

- Start/stop/status:
  garrison arma-reforger start -d ./servers/arma-reforger \
    --config ./servers/arma-reforger/profiles/config.json \
    --profile ./servers/arma-reforger/profiles \
    --port 20001 --query-port 27016 --browser-port 17777
  garrison arma-reforger stop -d ./servers/arma-reforger
  garrison arma-reforger status -d ./servers/arma-reforger

### Examples (isolated mode)
- Install/update in sandbox:
  garrison arma-reforger install --isolation=systemd
  garrison arma-reforger update --isolation=systemd

- Start/stop/status via systemd unit:
  garrison arma-reforger start --isolation=systemd \
    --port 20001 --query-port 27016 --browser-port 17777
  garrison arma-reforger stop --isolation=systemd
  garrison arma-reforger status --isolation=systemd

Notes:
- Systemd unit and data paths (isolated mode):
  - Unit: `/etc/systemd/system/garrison-arma-reforger.service`
  - State: `/var/lib/garrison/arma-reforger` (app, profiles, config)
  - Cache: `/var/cache/garrison/steamcmd`
  - Logs: `/var/log/garrison/arma-reforger`
