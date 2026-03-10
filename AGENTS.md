# Rig

Tmux-based development workflow tool written in Go. Manages isolated tmux sessions per repo, crew workspaces via git worktrees, and a document-driven feature development workflow with ephemeral workers ("polecats").

## Build and Test

```bash
make build          # Build to go/bin/rig
make test           # Run all tests
make install        # Install to $GOPATH/bin
```

All Go source is under `go/`. The module path is `github.com/mstrand/rig`. The single binary entry point is `go/cmd/rig/main.go`.

## Architecture

```
go/cmd/rig/main.go      CLI entry point, all cobra commands defined here
go/pkg/config/           Config from env vars (RIGS_BASE, CREW_BASE, RIG_USE_CC, etc.)
go/pkg/tmux/             Tmux session lifecycle (create, kill, list, attach)
go/pkg/crew/             Crew workspace management (add, start, remove via git worktrees)
go/pkg/git/              Git operations (worktrees, branches, repo detection)
go/pkg/polecat/          Ephemeral worker name generation (polecat_<name> format)
go/pkg/work/             Work directory scaffolding, progress parsing, hook/formula system
```

### Key Concepts

- **Rig** - A tmux session bound to a repo in `$RIGS_BASE` (default `~/git`). Session name = repo directory name.
- **Crew** - Git worktree-based workspaces under `$CREW_BASE` (default `~/crew/<rig>/<name>/`). Each gets a tmux session named `<rig>@<name>`.
- **Polecat** - An ephemeral crew member with an auto-generated name (`polecat_emma`, `polecat_maya`, etc.). Created by `rig sling` for one-off work assignments.
- **Work** - Document-driven feature development. Lives in `work/<feature>/` with spec, design, breakdown, progress, and hook files. Feature branches use `feat/<name>` convention.
- **Formula** - Reusable workflow template in `work/formula/*.md` defining development phases.

### Session Naming

- Rig sessions: `<repo-name>` (e.g., `myapp`)
- Crew sessions: `<rig>@<name>` (e.g., `myapp@tracy`, `myapp@polecat_emma`)
- Tmux replaces dots with underscores in session names; `NormalizeSessionName()` handles this.

### Discovery Model

`rig status` discovers sessions via `tmux list-sessions` and filters by type (rigs have no `@`, crew sessions do). `rig crew ls` discovers workspaces by scanning the `$CREW_BASE` filesystem. After a restart, tmux sessions are gone so `rig status` won't show crew workspaces that still exist on disk -- `rig crew ls` will show them as `[stopped]`.

### Two Display Modes

- **Native tmux** (default): Two windows per session (Claude Code + Terminal)
- **iTerm2 integration** (`RIG_USE_CC=true`): Single window with two panes (70/30 split), uses `tmux -CC`

## Conventions

- Tests use standard `testing` package with table-driven subtests. No test framework dependencies.
- Config is entirely from environment variables, no config files.
- All commands use `cobra`. Every command is defined as a function returning `*cobra.Command` in `main.go`.
- Git operations shell out to `git` via `exec.Command`. Tmux operations shell out to `tmux`.
- Crew branch naming: `<name>/work`. Feature branch naming: `feat/<name>`.
- Polecat names come from a hardcoded pool of 24 names in `polecat.go`.
