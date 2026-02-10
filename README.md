# Rig - Tmux-based Development Workflow

Rig manages tmux-based development environments where each repo runs as an independent tmux session.

The backstory is [here](README.BACKGROUND.md).

## Installation

### Using go install

```bash
go install github.com/mstrand/rig/go/cmd/rig@latest
```

Make sure `$GOPATH/bin` (or `$HOME/go/bin`) is in your `PATH`.

### Build from source

```bash
git clone https://github.com/mstrand/rig
cd rig/go
go build -o ~/bin/rig ./cmd/rig
```

Or use the Makefile:

```bash
cd rig
make install
```

### Requirements

- Go 1.25.6 or later
- tmux
- git
- iTerm2 (optional, for RIG_USE_CC mode)

## Workflow Concept

Each repo gets its own tmux **session** (named after the repo). Rig supports two modes:

### Native Tmux Mode (Default)

Each session has two **windows**:

```
Session: myrepo
├─ Window 1: Claude Code
└─ Window 2: Terminal

Session: another-repo
├─ Window 1: Claude Code
└─ Window 2: Terminal
```

### iTerm2 Integration Mode (RIG_USE_CC=true)

Each session has one **window** split into two **panes**:

```
Session: myrepo
└─ Window: myrepo
   ├─ Pane: Claude Code (70%)
   └─ Pane: Terminal (30%)
```

### Key Design Principles

- **Session-per-repo architecture** - Each repo gets its own independent tmux session
- **Complete isolation** - Crashes or issues in one rig don't affect others
- **One Claude Code per repo** - Each repo has exactly one Claude Code instance, preventing conflicts
- **Independent lifecycle** - Detach from, restart, or kill individual rigs without affecting others
- **Easy switching** - Switch between rigs using `rig switch <name>` or tmux/iTerm2 native switching
- **Works anywhere** - Run `rig up` from inside or outside tmux, it handles the context automatically
- **Two modes** - Native tmux (default) or iTerm2 integration mode

## Tmux Terminology Quick Reference

- **Session** - An independent tmux workspace (each rig is one session, named after the repo)
- **Window** - A full-screen workspace within a session (native mode: 2 windows per rig)
- **Pane** - A split section within a window (iTerm2 mode: 2 panes per window)

## Modes

### Native Tmux Mode (Default)

Standard tmux behavior with two windows per rig:
- No flashing when switching sessions
- Use standard tmux keybindings to switch windows (prefix + 1/2)
- Works with any terminal emulator
- Set `RIG_USE_CC=false` or leave unset (default)

### iTerm2 Integration Mode

Uses `tmux -CC` for iTerm2 integration:
- Each rig session appears as a separate native iTerm2 window
- Multiple rigs can be visible simultaneously (arrange windows side-by-side)
- Use native macOS window switching (Cmd+` or Mission Control)
- Single window with split panes per rig
- Set `RIG_USE_CC=true` to enable

## Usage

### Start or attach to a rig
```bash
rig up <repo-name>
```
Creates a new session if it doesn't exist, or switches to existing one.
Works from anywhere - inside or outside tmux.

### See what's running
```bash
rig status    # or: rig ls
```
Shows all active rig sessions.

### Switch between rigs

Using rig commands:
```bash
rig switch <repo-name>    # Switch to a different rig session
rig up <repo-name>        # Also switches if session already exists
```

Native tmux mode:
```bash
prefix + n/p              # Next/previous window (Claude Code / Terminal)
prefix + 1                # Jump to Claude Code window
prefix + 2                # Jump to Terminal window
tmux switch-client -t <repo-name>   # Switch to a different rig session
```

iTerm2 integration mode (RIG_USE_CC=true):
```bash
Cmd + `                   # Cycle through iTerm2 windows (each rig is a window)
Mission Control           # See all rig windows at once
```

### Shut down a rig
```bash
rig down <repo-name>
```
Kills the session for that repo.

### List all available repos
```bash
rig list
```
Shows all git repos in `$RIGS_BASE` and their status.

### Shut down all rigs
```bash
rig killall
```
Kills all rig sessions.

## Configuration

Environment variables:

```bash
# Base directory for repos (default: ~/git)
export RIGS_BASE="$HOME/projects"

# Use iTerm2 integration mode (default: false)
export RIG_USE_CC=true
```

Or set per-command:
```bash
RIG_USE_CC=true rig up myrepo
```

## Examples

```bash
# Start working on your first repo (creates a new session: "configs")
rig up configs

# Open another repo (creates a new session: "myapp")
rig up myapp

# Check what's running
rig status
# Output:
#   configs    (session)
#   myapp      (session)

# Switch between rigs using rig commands
rig switch configs

# Or use iTerm2 native window switching
# Cmd + `  (cycle through rig windows)

# Or use tmux directly
tmux switch-client -t myapp

# Shut down a specific rig
rig down myapp

# Shut down all rigs (kills all rig sessions)
rig killall
```

## Crew Workspaces

Crew workspaces enable multiple people (or personas) to work on the same repo simultaneously using git worktrees. Each crew member gets:

- **Isolated workspace** - No interference with the main repo or other crew members
- **Own branch** - `<name>/work` branch created from the base branch
- **Own tmux session** - Named `<rig>@<name>` (e.g., `notes@tracy`)
- **Zero copying** - Uses git worktrees (no file duplication or syncing)

### Crew Concept

```
~/git/notes/              # Main repo
~/crew/tracy/notes/       # Tracy's worktree (branch: tracy/work)
~/crew/alex/notes/        # Alex's worktree (branch: alex/work)

Sessions:
- notes              (main rig)
- notes@tracy        (Tracy's crew workspace)
- notes@alex         (Alex's crew workspace)
```

### Crew Commands

#### Create a crew workspace

```bash
# From within a repo
cd ~/git/notes
rig crew add tracy

# Or specify the rig explicitly
rig crew add tracy --rig=notes
```

This creates:
- `~/crew/tracy/notes/` (git worktree)
- `tracy/work` branch (from `main` or `$RIG_DEFAULT_BRANCH`)
- `notes@tracy` tmux session (Claude Code + Terminal)

#### Start an existing crew workspace

```bash
rig crew start tracy
# Or: rig crew start tracy --rig=notes
```

Recreates the tmux session if needed, checks branch, and attaches.

#### List crew workspaces

```bash
# List all crew
rig crew ls

# List specific crew member
rig crew ls tracy
```

Output:
```
=== Crew Workspaces ===

tracy
  - notes (notes@tracy) [running]
  - myapp (myapp@tracy) [stopped]

alex
  - notes (notes@alex) [running]
```

#### Show active crew sessions

```bash
rig crew status
```

Output:
```
=== Active Crew Sessions ===

notes@tracy      ~/crew/tracy/notes       tracy/work    [running]
notes@alex       ~/crew/alex/notes        alex/work     [running]
```

#### Remove a crew workspace

```bash
rig crew remove tracy
# Or: rig crew remove tracy --rig=notes
```

This:
1. Kills the tmux session
2. Removes the git worktree
3. Asks whether to delete the branch (defaults to yes)
4. Cleans up empty directories

### Crew with Main Rig Commands

Crew sessions integrate with main rig commands:

```bash
# Show all rigs AND crew
rig status

# Output:
# === Active Rigs ===
#   notes
#     └─ ~/git/notes
#
# === Crew ===
#   notes@tracy (tracy/notes)
#     └─ ~/crew/tracy/notes

# Switch to crew sessions
rig switch notes@tracy        # Switch to crew session

# Kill options
rig killall                   # Kill rigs only (default)
rig killall --crew            # Kill both rigs and crew
rig killall --crew-only       # Kill only crew sessions
```

### Crew Configuration

Environment variables:

```bash
# Base directory for crew workspaces (default: ~/crew)
export CREW_BASE="$HOME/crew"

# Base branch for crew worktrees (default: main)
export RIG_DEFAULT_BRANCH=main
```

### Crew Workflow Examples

#### Multiple people on same repo

```bash
# Main repo
cd ~/git/notes
rig up notes                  # Session: notes

# Create workspace for Tracy
rig crew add tracy            # Session: notes@tracy

# Create workspace for Alex
rig crew add alex             # Session: notes@alex

# Switch between them
rig switch notes              # Switch to main rig
rig switch notes@tracy        # Switch to Tracy's crew
rig switch notes@alex         # Switch to Alex's crew

# Check status
rig status
# Shows:
# - notes (main rig)
# - notes@tracy (crew)
# - notes@alex (crew)
```

#### Single person, multiple contexts

```bash
# Main work
rig up notes                  # Session: notes

# Experimental feature
rig crew add experiment       # Session: notes@experiment
# Work on experiment/work branch

# Switch back to main
rig switch notes

# Clean up when done
rig crew remove experiment
```

### Crew Tips

- **Branch naming**: Crew branches use `<name>/work` format (e.g., `tracy/work`)
- **Rig inference**: Run crew commands from anywhere in `~/git/` or `~/crew/` - rig is auto-detected
- **Idempotent add**: Running `rig crew add <name>` on existing workspace attaches instead of failing
- **Branch conflicts**: If branch exists, you're prompted to use it or cancel
- **Detached state**: `rig crew remove` handles cases where directories are gone but git still tracks them
- **Session naming**: The `@` separator makes crew sessions easy to identify and prevents collisions

## Work-Based Development Workflow

Rig provides a structured, document-driven workflow for feature development with git-native state tracking. Work is defined in markdown documents, tracked through checklists, and can be assigned to crew members or ephemeral "polecats" (temporary workers).

### Key Concepts

- **Work directories** - Each feature lives in `work/<feature-name>/` with spec, design, breakdown, and progress files
- **Feature branches** - Automatically created as `feat/<feature-name>`
- **Formulas** - Reusable workflow templates (e.g., `build.md`) that define the development process
- **Progress tracking** - Simple markdown checklists in `progress.md` track current status
- **Polecats** - Auto-named ephemeral workers for one-off tasks
- **Git-native** - All state lives in git; no external databases or hidden files

### Work Directory Structure

```
work/
├── formula/
│   ├── build.md         # Default workflow formula
│   └── hotfix.md        # Optional: fast-track workflow
│
└── build-frontend/      # Feature work directory
    ├── spec.md          # What we're building
    ├── design.md        # How we'll build it
    ├── breakdown.md     # Implementation tasks
    ├── progress.md      # Current status and checklist
    └── hook.md          # Worker startup instructions (created by rig sling)
```

### Creating New Work

```bash
# Create a new feature
rig work create build-frontend
```

This creates:
- `work/build-frontend/` directory with scaffolded files
- `feat/build-frontend` feature branch
- Formula directory `work/formula/` (if it doesn't exist)
- Default `work/formula/build.md` formula (if it doesn't exist)
- Initial commit on the feature branch

**Behavior:**
- Warns but continues if work directory already exists
- Only creates missing files (never overwrites)
- Installs missing formulas but never overwrites existing ones

### Viewing Work Status

```bash
rig work status
```

Shows all active work across all rigs:

```
Active Work:

  📦 myapp
    build-frontend      [In Design]      polecat_emma    feat/build-frontend
      → Design review
    optimize-queries    [Implementation] tracy           feat/optimize-queries
      → Code review

  📦 bonsai
    add-auth           [Blocked]        -               feat/add-auth
      → Awaiting backend API
```

### Assigning Work with Sling

The `rig sling` command assigns work to crew members or creates ephemeral polecats:

```bash
# Create ephemeral polecat (auto-named worker)
rig sling work/build-frontend

# Assign to named crew member
rig sling work/build-frontend --to=tracy

# Use specific formula instead of default
rig sling work/build-frontend --formula=hotfix

# Work on it yourself in current session
rig sling work/build-frontend --self
```

**What happens during sling:**
1. Creates `hook.md` file with workflow instructions
2. For new polecats:
   - Creates crew workspace and tmux session
   - Checks out feature branch via git worktree
   - Sends initial message: `"Check your hook: rig hook and follow instructions there"`
3. For existing crew or `--self`:
   - Creates/updates hook.md
   - Provides copy-paste instruction for user

**Re-slinging:**
If work is already assigned, you'll be warned and asked for confirmation before reassigning.

### Hook Instructions

```bash
# Display current hook
rig hook
```

Shows the hook instructions for the current work:

```
🪝 Hook: build-frontend

Follow the workflow defined in: work/formula/build.md
Your spec file is: work/build-frontend/spec.md

Additional context:
- Design: work/build-frontend/design.md
- Breakdown: work/build-frontend/breakdown.md
- Progress: work/build-frontend/progress.md

Read and understand both the formula and spec, then begin working.
Update progress.md as you complete each step.
Commit intermediate progress at each phase.
```

Agents run this command on startup to get their instructions.

### Progress Tracking

Each work directory has a `progress.md` file with a checklist:

```markdown
# Progress: Build Frontend

## Status: In Design
## Assigned to:

## Checklist
- [x] Spec review
- [x] Initial design
- [ ] Design review
- [ ] Design corrections
- [ ] Implementation breakdown
- [ ] Implementation
- [ ] Code review
- [ ] Fixes
- [ ] Push feature branch
- [ ] Cleanup: `rig crew remove <worker-name>`

## Notes
Add any relevant notes, decisions, or blockers here.
```

**Rules:**
- First unchecked item = current task
- Update as work progresses
- Commit intermediate progress at each phase
- Final items remind you to push and cleanup

### Formula System

Formulas define reusable workflows. They live in `work/formula/` and are referenced by `rig sling`:

```bash
# Use default formula (work/formula/build.md)
rig sling work/build-frontend

# Use specific formula
rig sling work/build-frontend --formula=hotfix
```

Formulas emphasize:
- Spec review → Design → Design review → Implementation → Code review → Fixes → Push
- Committing intermediate progress at each step
- Keeping progress.md checklist up to date
- Leaving work in consistent state at each phase

### Managing Polecats

Polecats are ephemeral workers with auto-generated names:

```bash
# List all crew including polecats
rig crew ls

# Attach to review polecat's work
rig crew attach polecat_emma

# Remove when done (removes worktree too)
rig crew remove polecat_emma

# Bulk cleanup
rig crew prune --polecats
```

**Polecat naming:**
- Random names from predefined pool
- Format: `polecat_<name>`
- Visually distinguished in `rig crew ls` output
- Must be manually cleaned up after work completes

### Work Workflow Example

```bash
# Create new feature work
rig work create user-auth

# Edit the spec
vim work/user-auth/spec.md

# Assign to a polecat
rig sling work/user-auth

# Check status across all work
rig work status

# Attach to review progress
rig crew attach polecat_maya

# When done, cleanup
rig crew remove polecat_maya
```

### Work Configuration

The workflow system uses the same base directories as crew:

```bash
# Base directory for crew workspaces (default: ~/crew)
export CREW_BASE="$HOME/crew"

# Base branch for feature branches (default: main)
export RIG_DEFAULT_BRANCH=main
```
