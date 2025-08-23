# Project Implementation Plan

The project name has been updated to "garrison" as specified. This adjustment replaces all references to the tentative name "steamgsm" in the original plan. The CLI tool will now be referred to as "garrison," with corresponding changes to binary names, module paths, and directory structures where applicable. Below is the complete and exhaustive revised plan, incorporating this change while maintaining adherence to Go best practices, modularity, and the specified requirements.

## Project Overview

The Go CLI tool, named "garrison," will provide a command-line interface for creating, installing, updating, managing, and monitoring dedicated Steam game servers. It mimics the core functionalities of LinuxGSM, including server deployment, lifecycle management (start, stop, restart), monitoring, and configuration handling. Implemented in Go for cross-platform compatibility (primarily Linux, with potential extensions to Windows and macOS), it integrates fully with SteamCMD for server downloads and updates, ensuring robust error handling, logging, and idempotency.

Key design principles remain unchanged:

- **Modularity**: Separation of concerns using Go's package system, with non-exported library code in the `internal/` directory.
- **Best Practices**: Idiomatic Go error handling, context-aware operations, concurrent safety, and comprehensive testing. Directory structure follows standard Go layouts.
- **CLI Framework**: Kong library for declarative CLI parsing.
- **SteamCMD Integration**: Complete wrapper with subprocess execution, output parsing, and retries.
- **Security and Reliability**: Anonymous logins for public servers, configurable credentials, and signal handling.
- **Extensibility**: Support for game-specific configurations via YAML.

The implementation assumes a Unix-like environment initially, with configurable paths for broader OS support.

## Directory Structure

The structure adheres to Go idioms:

- **`cmd/garrison/`**: Main entry point and CLI command definitions.
  - `main.go`: Bootstraps the CLI using Kong, parses arguments, and dispatches commands.
  - `root.go`: Defines the root command struct (global flags like verbosity, config path).
  - `create.go`: Command for creating a new server instance.
  - `install.go`: Command for installing a server.
  - `update.go`: Command for updating a server.
  - `start.go`: Command for starting a server.
  - `stop.go`: Command for stopping a server.
  - `restart.go`: Command for restarting a server.
  - `monitor.go`: Command for monitoring server status.
  - `details.go`: Command for displaying server details.
  - `list.go`: Command for listing managed servers.
  - `delete.go`: Command for deleting a server instance.

- **`internal/`**: Private packages for application logic.
  - `steamcmd/`:
    - `steamcmd.go`: Core wrapper for SteamCMD execution (login, install, update, validate).
    - `parser.go`: Output parsing utilities (regex or string matching).
    - `config.go`: SteamCMD configuration structures.
  - `server/`:
    - `manager.go`: Server lifecycle logic (process management with PID files).
    - `config.go`: Server configuration handling (YAML-based).
    - `monitor.go`: Health checking and log tailing.
    - `storage.go`: Persistent metadata storage (JSON files in `~/.garrison/`).
  - `utils/`:
    - `logging.go`: Structured logging with `log/slog`.
    - `fs.go`: File system utilities.
    - `signals.go`: Signal handling.
    - `validation.go`: Input validation.

- **`go.mod` and `go.sum`**: Module definitions (e.g., `module github.com/user/garrison`).
- **`README.md`**: Project documentation (content provided below).
- **`LICENSE`**: MIT or Apache 2.0.
- **`.gitignore`**: Standard Go ignores.
- **`tests/`**: Optional top-level for integration tests; use in-package tests otherwise.
- **`configs/`**: Example YAML configs for games.

## Dependencies

Unchanged from the original plan:

- Standard library: `os`, `exec`, `io`, `context`, `sync`, `regexp`, `encoding/json`, `gopkg.in/yaml.v3`.
- External: `github.com/alecthomas/kong`.
- Vendoring and Go 1.21+.

## Main Components

1. **CLI Setup with Kong**:
   - Root struct in `cmd/garrison/root.go`, e.g.:

     ```go
     type CLI struct {
         Config string `help:"Path to config file" default:"~/.garrison/config.yaml"`
         Verbose bool `short:"v" help:"Enable verbose logging"`
         Create CreateCmd `cmd:"create" help:"Create a new server instance"`
         // Other commands...
     }
     ```

   - `main.go` uses `kong.Parse(&cli)`.

2. **SteamCMD Wrapper (in `internal/steamcmd/`)**:
   - `SteamCMD` struct and methods (e.g., `InstallApp(appID int, validate bool) error`).
   - Context-aware, with output parsing and retries.

3. **Server Management (in `internal/server/`)**:
   - Server struct and lifecycle methods.
   - YAML config loading for games (e.g., appID 740 for CS2).

## Commands List

Commands remain the same, e.g.:

- `garrison create <name> --app-id <id> --dir <path>`.
- Global flags: `--config`, `--verbose`.

## Implementation Steps

1. **Setup Project (1-2 days)**:
   - `go mod init github.com/user/garrison`.
   - Add Kong.
   - Create structure.

2. **SteamCMD Wrapper (3-5 days)**: Develop and test.

3. **Server Management Logic (4-6 days)**: Implement and integrate.

4. **CLI Commands (3-4 days)**: Define and wire.

5. **Utilities and Polish (2-3 days)**: Add features.

6. **Testing (4-5 days)**: Unit, integration, coverage.

7. **Documentation and Deployment (1-2 days)**: Update README.md, build binary (`go build -o garrison cmd/garrison/main.go`).
