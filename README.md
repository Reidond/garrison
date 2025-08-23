
# Garrison

Garrison is a Go-based command-line tool for managing dedicated Steam game servers, inspired by LinuxGSM. It enables users to create, install, update, start, stop, restart, monitor, and delete server instances with full integration to SteamCMD. Designed for modularity and cross-platform use (primarily Linux), it follows Go best practices for reliability and extensibility.

## Features

- **Server Creation and Management**: Easily set up new servers with custom directories and configurations.
- **SteamCMD Integration**: Automated installation, updates, and validation of game servers.
- **Lifecycle Controls**: Start, stop, restart, and monitor servers with process handling and logging.
- **Configuration Support**: YAML-based configs for game-specific settings (e.g., ports, arguments).
- **Persistence**: Stores server metadata in `~/.garrison/` for easy listing and management.
- **Extensibility**: Add support for new games via configuration files.

## Supported games and provider features

Garrison includes a small provider layer for game-specific helpers. Current providers:

- Reforger (Arma Reforger)
  - list-scenarios: Queries the server binary with `-listScenarios` and prints parsed `.conf` scenario paths.
  - template: Prints a ready-made YAML template for starting the server via `garrison start` (see `configs/reforger-example.yaml`).
  - init-config: Creates the native JSON server config (e.g., `Configs/ServerConfig.json`). Required option: `scenarioId`.

Generic commands that use providers:

- `garrison game list-scenarios --game reforger [--install-dir ...] [--binary ...] [--timeout ...] [--raw]`
- `garrison game template --game reforger`
- `garrison game init-config --game reforger --scenario-id ... --output ./Configs/ServerConfig.json`

The legacy alias `garrison reforger` remains for convenience and delegates to the same provider implementations.

## Installation

### Prerequisites

- Go 1.21 or later.
- SteamCMD installed (Garrison will detect or download it if configured).

### Building from Source

- Clone the repository:

```sh
git clone https://github.com/user/garrison.git
cd garrison
```

- Install dependencies:

```sh
go mod tidy
```

- Build the binary:

```sh
go build -o garrison cmd/garrison/main.go
```

- Move the binary to your PATH (optional):

```sh
sudo mv garrison /usr/local/bin/
```

## Usage

Garrison uses a subcommand structure. Run `garrison --help` for full details.

### Global Flags

- `--config <path>`: Path to configuration file (default: `~/.garrison/config.yaml`).
- `--verbose`: Enable detailed logging.

### Commands

- **create \<name\>**: Create a new server instance.
  - `--app-id <id>`: Steam App ID (required).
  - `--dir <path>`: Installation directory (default: current dir).
  - Example: `garrison create myserver --app-id 740 --dir ~/servers/cs2`

- **install \<name\>**: Install the server using SteamCMD.
  - `--validate`: Validate files during install.
  - Example: `garrison install myserver --validate`

- **update \<name\>**: Update the server.
  - Example: `garrison update myserver`

- **start \<name\>**: Start the server.
  - `--args <extra>`: Additional command-line arguments.
  - Example: `garrison start myserver`

- **stop \<name\>**: Stop the server gracefully.
  - Example: `garrison stop myserver`

- **restart \<name\>**: Restart the server.
  - Example: `garrison restart myserver`

- **monitor \<name\>**: Monitor server status.
  - `--watch`: Continuous monitoring.
  - Example: `garrison monitor myserver --watch`

- **details \<name\>**: Display server details and status.
  - Example: `garrison details myserver`

- **list**: List all managed servers.
  - Example: `garrison list`

- **delete \<name\>**: Delete a server instance.
  - Example: `garrison delete myserver`

## Configuration

Server configs are stored in YAML format. Example for Counter-Strike 2 (appID 740):

```yaml
app_id: 740
start_args: "-dedicated -port 27015"
ports:
  - 27015/udp
executable: run.sh          # optional: binary or script; relative to install dir unless absolute
working_dir: .              # optional: working directory; relative to install dir unless absolute
command_template: |
  ./run.sh {{args}}         # optional: full command template; supports {{install_dir}}, {{app_id}}, {{name}}, {{args}}
env:                        # optional: environment variables
  LD_LIBRARY_PATH: ".:${LD_LIBRARY_PATH}"
log_file: logs/server.log   # optional: log file path to tail in monitor
```

Place game-specific configs in `configs/` or override via flags. If `command_template` is provided, it overrides `executable` + args assembly.

## Contributing

Contributions are welcome. Please follow these steps:

1. Fork the repository.
2. Create a feature branch.
3. Commit changes with clear messages.
4. Open a pull request.

Ensure tests pass: `go test ./...`

## License

This project is licensed under the MIT License. See `LICENSE` for details.

## Acknowledgments

- Inspired by LinuxGSM.
- Built with Kong for CLI parsing.
