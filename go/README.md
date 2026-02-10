# Rig - Go Implementation

A Go port of the rig crew management system. This provides a tmux-based development environment manager with support for crew workspaces (git worktrees).

## Features

- **Rig Management**: Create, switch, and manage tmux sessions for different repos
- **Crew Workspaces**: Git worktree-based parallel workspaces for team collaboration
- **Session Management**: Integrated tmux session creation and management
- **Environment Configuration**: Configurable base directories and default branches

## Installation

### From Source

```bash
go build -o bin/rig ./cmd/rig
# Optionally, copy to your PATH
cp bin/rig ~/bin/rig  # or /usr/local/bin/rig
```

### Using Go Install

```bash
go install github.com/mstrand/rig/cmd/rig@latest
```

## Configuration

Environment variables:

- `RIGS_BASE` - Base directory for repos (default: `~/git`)
- `CREW_BASE` - Base directory for crew workspaces (default: `~/crew`)
- `RIG_USE_CC` - Set to `true` for iTerm2 integration mode (default: `false`)
- `RIG_DEFAULT_BRANCH` - Default branch for crew worktrees (default: `main`)

## Usage

### Rig Commands

```bash
# Start or switch to a rig
rig up <name>

# Shut down a rig
rig down <name>

# Show all active rigs and crew
rig status
# or
rig ls

# List available repos
rig list

# Switch to a rig or crew
rig switch <name>

# Shut down all rigs
rig killall

# Shut down all rigs and crew
rig killall --crew

# Shut down only crew
rig killall --crew-only
```

### Crew Commands

```bash
# Create crew workspace (infers rig from current directory)
rig crew add <name>

# Create crew workspace (explicit rig)
rig crew add <name> --rig=<repo>

# Start/attach to crew workspace
rig crew start <name>

# Remove crew workspace
rig crew remove <name>
# or
rig crew rm <name>

# List all crew workspaces
rig crew ls

# List specific crew member's workspaces
rig crew ls <name>

# Show active crew sessions
rig crew status
```

## Directory Structure

```
~/git/                      # RIGS_BASE - main repos
  ├── notes/
  ├── myapp/
  └── myproject/

~/crew/                     # CREW_BASE - crew workspaces
  ├── notes/
  │   ├── tracy/           # git worktree for tracy
  │   └── alex/            # git worktree for alex
  └── myapp/
      └── tracy/           # git worktree for tracy
```

## Session Naming

- **Rig sessions**: `<repo-name>` (e.g., `notes`)
- **Crew sessions**: `<repo>@<name>` (e.g., `notes@tracy`)

## Architecture

### Package Structure

```
cmd/rig/          - CLI entry point and command definitions
pkg/
  ├── config/     - Configuration management
  ├── git/        - Git worktree operations
  ├── tmux/       - Tmux session management
  └── crew/       - Crew workspace logic
```

### Key Components

#### Config Package

Manages environment configuration and path generation:

- Load configuration from environment variables
- Generate repo and crew paths
- Generate session and branch names

#### Git Package

Handles all git worktree operations:

- Create and remove worktrees
- Branch management
- Repository validation
- Worktree status checking

#### Tmux Package

Manages tmux session lifecycle:

- Create rig sessions (native and iTerm2 modes)
- Create crew sessions
- Attach to sessions
- List and kill sessions

#### Crew Package

High-level crew workspace operations:

- Add crew workspaces
- Start existing workspaces
- Remove workspaces
- Validate crew names
- Infer rig from context

## Testing

The project includes comprehensive tests for all packages:

```bash
# Run all tests
go test ./pkg/...

# Run with verbose output
go test -v ./pkg/...

# Run specific package tests
go test ./pkg/config
go test ./pkg/git
go test ./pkg/crew
```

### Test Coverage

- **Config Package**: Environment variable handling, path generation
- **Git Package**: Git operations with real repositories
- **Crew Package**: Full workflow testing including worktree creation and removal

All tests use actual git operations on temporary repositories to ensure real behavior is tested, not just mocked interfaces.

## Implementation Notes

### Crew Path Structure

The crew directory structure follows the pattern: `~/crew/<repo>/<name>`

This groups all crew members working on the same repo together, making it easier to:
- See who's working on what repo
- Clean up empty repo directories
- Manage crew workspaces per-repo

### Rig Inference

The system can infer which rig you're working with from:

1. **Explicit flag**: `--rig=<name>`
2. **Current directory**: If in `~/git/<repo>/...`
3. **Crew directory**: If in `~/crew/<repo>/<name>/...`
4. **Tmux session**: If in a rig or crew tmux session

### Session Management

- **Native mode**: Each rig gets 2 tmux windows (Claude Code, Terminal)
- **iTerm2 mode** (`RIG_USE_CC=true`): Each rig gets 1 window with 2 panes

### Worktree Cleanup

When removing crew workspaces:
- Prompts before deleting branches
- Cleans up stale worktree metadata
- Removes empty repo directories
- Handles detached worktree states gracefully

## Differences from Bash Version

The Go implementation maintains complete CLI compatibility with the bash version while adding:

1. **Better error handling**: Structured errors with context
2. **Improved testing**: Unit and integration tests for all components
3. **Type safety**: Go's type system prevents many common bugs
4. **Better maintainability**: Modular package structure
5. **Cross-platform support**: Works on macOS, Linux, and Windows (with tmux)

## Development

### Building

```bash
go build -o bin/rig ./cmd/rig
```

### Running Tests

```bash
go test ./pkg/...
```

### Adding New Commands

1. Add command definition in `cmd/rig/main.go`
2. Implement logic in appropriate package
3. Add tests in package test file
4. Update this README

## License

Same as the bash version (see main project LICENSE).

## Credits

Go port of the original bash `rig` and `rig-crew` scripts.
