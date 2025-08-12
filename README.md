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

## Testing

Run all tests:

- `go test -v ./...`

## SteamCMD (user-only install)

Install SteamCMD without root, into your home directory:

1) Dependencies (32-bit libs)
- Debian/Ubuntu: `sudo apt-get update && sudo apt-get install -y lib32gcc-s1`
- Fedora/RHEL: `sudo dnf install -y glibc.i686 libstdc++.i686`
- Arch: `sudo pacman -S --needed lib32-gcc-libs`

2) Download & unpack
```bash
mkdir -p "$HOME/.local/opt/steamcmd"
cd "$HOME/.local/opt/steamcmd"
curl -fsSL https://steamcdn.cloudflare.steamstatic.com/client/installer/steamcmd_linux.tar.gz | tar -xz
```

3) Create a user-local shim and add to PATH
```bash
mkdir -p "$HOME/.local/bin"
cat > "$HOME/.local/bin/steamcmd" <<'EOF'
#!/usr/bin/env bash
exec "$HOME/.local/opt/steamcmd/steamcmd.sh" "$@"
EOF
chmod +x "$HOME/.local/bin/steamcmd"
echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.bashrc"
export PATH="$HOME/.local/bin:$PATH"
```

4) First run (updates SteamCMD)
```bash
steamcmd +quit
```

Optional: pin for garrison
```bash
export GARRISON_STEAMCMD_BIN=steamcmd
```

## Temporary test user shell

Two safe options depending on whether you have sudo:

- With sudo (true throwaway user):
```bash
sudo useradd -m -s /bin/bash garrison-tmp
sudo -iu garrison-tmp bash
# inside the shell, repeat the user-only SteamCMD install steps above
# when done:
exit
sudo userdel -r garrison-tmp
```

- Without sudo (ephemeral clean home for the current user):
```bash
mkdir -p /tmp/garrison-tmp-home
env -i HOME=/tmp/garrison-tmp-home \
    XDG_DATA_HOME=/tmp/garrison-tmp-xdg/data \
    XDG_CACHE_HOME=/tmp/garrison-tmp-xdg/cache \
    XDG_STATE_HOME=/tmp/garrison-tmp-xdg/state \
    XDG_CONFIG_HOME=/tmp/garrison-tmp-xdg/config \
    PATH=/usr/local/bin:/usr/bin:/bin \
    SHELL=/bin/bash USER=$USER \
    bash --noprofile --norc
# inside that shell, do the user-only SteamCMD install
```

## Manual testing (Arma Reforger example, safe/stubbed)

Use a stub server so nothing real is downloaded or executed. For direct mode:
```bash
mkdir -p /tmp/garrison-test/arma/profiles
printf '{}' > /tmp/garrison-test/arma/profiles/config.json
cat > /tmp/garrison-test/arma/ArmaReforgerServer <<'EOF'
#!/usr/bin/env bash
echo "Stub server started: $0 $@" >&2
sleep 300
EOF
chmod +x /tmp/garrison-test/arma/ArmaReforgerServer

./garrison start \
  --server arma-reforger --isolation none \
  --install-dir /tmp/garrison-test/arma \
  --config /tmp/garrison-test/arma/profiles/config.json \
  --profile /tmp/garrison-test/arma/profiles \
  --port 21001 --query-port 27016 --browser-port 17777

./garrison status --server arma-reforger --isolation none --install-dir /tmp/garrison-test/arma
./garrison stop   --server arma-reforger --isolation none --install-dir /tmp/garrison-test/arma
./garrison delete --server arma-reforger --isolation none --install-dir /tmp/garrison-test/arma
```

For systemd user mode (no downloads):
```bash
export GARRISON_STEAMCMD_BIN=true
export XDG_DATA_HOME=/tmp/garrison-xdg/data
export XDG_CACHE_HOME=/tmp/garrison-xdg/cache
export XDG_STATE_HOME=/tmp/garrison-xdg/state
export XDG_CONFIG_HOME=/tmp/garrison-xdg/config
mkdir -p "$XDG_DATA_HOME/garrison/arma-reforger/app" "$XDG_DATA_HOME/garrison/arma-reforger/profiles"
printf '{}' > "$XDG_DATA_HOME/garrison/arma-reforger/profiles/config.json"
cat > "$XDG_DATA_HOME/garrison/arma-reforger/app/ArmaReforgerServer" <<'EOF'
#!/usr/bin/env bash
echo "Stub systemd server: $0 $@" >&2
sleep 300
EOF
chmod +x "$XDG_DATA_HOME/garrison/arma-reforger/app/ArmaReforgerServer"

./garrison start --server arma-reforger --isolation systemd \
  --port 21001 --query-port 27016 --browser-port 17777

systemctl --user status garrison-arma-reforger.service
./garrison stop   --server arma-reforger --isolation systemd
./garrison delete --server arma-reforger --isolation systemd
```
