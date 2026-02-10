# Rig - Development Environment Manager

## Overview

Rig is a tmux-based development environment manager that creates isolated, project-specific workspaces for software development. It manages the lifecycle of tmux sessions that are tightly coupled to git repositories, providing a consistent interface for spinning up development environments with integrated tooling (Claude Code, terminal access).

**Version:** 1.0  
**Language:** Bash  
**Dependencies:** tmux, git  
**Target Platform:** macOS (with iTerm2 integration support)

---

## Problem Statement

Modern development workflows often involve:
- Working across multiple projects simultaneously
- Context switching between repositories
- Running AI coding assistants (Claude Code) alongside traditional terminals
- Losing terminal state when closing windows or switching contexts
- Manually recreating development environments each time

**Rig solves this by:**
- Providing persistent tmux sessions tied to git repositories
- Automating the setup of multi-pane/multi-window development environments
- Making it trivial to switch between project contexts
- Preserving session state across reconnections

---

## Core Concepts

### Rig
A "rig" is a tmux session that represents a single project workspace. Each rig:
- Maps 1:1 to a git repository in `$RIGS_BASE` (default: `~/git/`)
- Has a predictable structure (Claude Code + Terminal)
- Can be brought up/down independently
- Persists across terminal disconnections

### Repository Discovery
Rigs are auto-discovered from directories in `$RIGS_BASE` that contain `.git/`. The directory name becomes the rig name.

### Session Naming
Rig sessions use the repository directory name directly:
- Repository: `~/git/notes` → Session: `notes`
- Repository: `~/git/myapp` → Session: `myapp`

This simple 1:1 mapping makes it easy to identify and switch between rigs.

---

## Architecture

### Directory Structure
```
~/git/                      # RIGS_BASE - all repos live here
  ├── notes/               # Repo 1
  ├── myapp/         # Repo 2
  └── my-project/          # Repo 3

# Tmux sessions created on-demand:
# - notes (session)
# - myapp (session)
# - my-project (session)
```

### Session Layouts

Rig supports two modes, controlled by `RIG_USE_CC` environment variable:

#### Native Mode (default, `RIG_USE_CC=false`)
```
Session: <repo-name>
├── Window 1: "Claude Code"     (runs claude CLI)
└── Window 2: "Terminal"        (bash/zsh in repo root)
```

**Navigation:**
- `prefix + n/p` - Next/previous window
- `prefix + 1/2` - Jump to specific window

#### iTerm2 Integration Mode (`RIG_USE_CC=true`)
```
Session: <repo-name>
├── Window 1: <repo-name>
    ├── Pane 1 (left 70%):  "Claude Code"
    └── Pane 2 (right 30%): "Terminal"
```

**Navigation:**
- Each session appears as a separate iTerm2 window (via `tmux -CC`)
- `Cmd + ` ` - Cycle through iTerm2 windows
- Native macOS window management

---

## Commands

### `rig up <name>`
Bring up a rig (create if new, switch if existing).

**Behavior:**
- If session exists: switch to it
- If session doesn't exist: create new session with layout, start Claude Code, attach

**Examples:**
```bash
rig up notes              # Create/switch to notes rig
rig up myapp        # Create/switch to myapp rig
```

**Validation:**
- Checks that `~/git/<name>` exists and contains `.git/`
- Fails if repository not found

---

### `rig down <name>`
Shut down a rig (kill tmux session).

**Behavior:**
- Kills the tmux session immediately
- Does not affect git repository or any uncommitted work
- Session state is lost (can be recreated with `rig up`)

**Examples:**
```bash
rig down notes            # Kill notes session
```

**Validation:**
- Fails if session doesn't exist

---

### `rig status` / `rig ls`
Show all active rigs.

**Output:**
```
=== Active Rigs ===

✓ notes                  (session)
  └─ ~/git/notes

  myapp            (session)
  └─ ~/git/myapp

Total: 2 active rigs
```

**Details:**
- `✓` indicates currently active session (if in tmux)
- Shows full path to repository
- Only shows sessions that map to valid git repos in `$RIGS_BASE`

---

### `rig list`
List all available repositories in `$RIGS_BASE`.

**Output:**
```
=== Available Repos in ~/git ===

  notes (running)
  myapp
  my-project

Total: 3 repos
```

**Details:**
- Shows all directories with `.git/` in `$RIGS_BASE`
- Marks running rigs with `(running)`
- Useful for discovering what rigs can be created

---

### `rig switch <name>`
Switch to a rig session.

```bash
rig switch notes          # Switch to notes rig
rig switch notes@tracy    # Switch to crew session
```

**Behavior:**
- If outside tmux: attaches to session
- If inside tmux: switches client to session
- Respects `RIG_USE_CC` mode
- Works with both rig sessions and crew sessions

---

### `rig killall`
Shut down all rigs.

**Behavior:**
- Kills all tmux sessions that map to repos in `$RIGS_BASE`
- Leaves other tmux sessions untouched
- Prompts for confirmation

**Safety:**
- Only kills sessions where session name matches a git repo directory
- Won't accidentally kill unrelated tmux sessions

---

## Environment Variables

### `RIGS_BASE`
Base directory for repositories.

**Default:** `~/git`  
**Usage:** `export RIGS_BASE=~/projects`

All repositories must live under this directory. Rig will only discover and manage repos here.

---

### `RIG_USE_CC`
Enable iTerm2 integration mode.

**Default:** `false`  
**Usage:** `RIG_USE_CC=true rig up notes`

**When true:**
- Uses `tmux -CC` for iTerm2 integration
- Each session becomes a native iTerm2 window
- Panes become native iTerm2 split panes
- Enables native macOS features (scrollback, search, copy/paste)

**When false (default):**
- Uses standard tmux
- All tmux keybindings work normally
- More portable (works in any terminal)

---

## Implementation Details

### Session Lifecycle

#### Creation (`rig_up`)
1. Check if session already exists
   - If yes: switch to it
   - If no: continue to creation
2. Validate repository exists in `$RIGS_BASE`
3. Create tmux session (detached)
4. Set up layout (mode-dependent)
5. Start Claude Code in appropriate pane/window
6. Initialize terminal pane with `git status`
7. Attach to session

#### Switching (`rig_switch`)
1. If in tmux: use `tmux switch-client`
2. If not in tmux: use `tmux attach-session` (or `tmux -CC attach`)

#### Teardown (`rig_down`)
1. Validate session exists
2. Kill session with `tmux kill-session`

### Repository Discovery
```bash
# Pseudocode
for dir in $RIGS_BASE/*; do
    if [ -d "$dir/.git" ]; then
        repo_name = basename($dir)
        # repo_name is a valid rig
    fi
done
```

Only directories with `.git/` are considered valid repositories.

### Session Detection
```bash
session_exists() {
    tmux has-session -t "$name" 2>/dev/null
}
```

Uses tmux's built-in session detection. Returns true if session exists, false otherwise.

---

## Error Handling

### Repository Not Found
```bash
$ rig up nonexistent
Error: Repo not found: ~/git/nonexistent
```

### Session Not Found (for down/switch)
```bash
$ rig down nonexistent
Error: Rig not found: nonexistent
```

### No Active Rigs
```bash
$ rig status
=== Active Rigs ===

No active rigs

Start a rig with: rig up <name>
```

---

## Design Decisions

### Why tmux?
- **Persistence:** Sessions survive terminal disconnections
- **Multiplexing:** Multiple panes/windows in one session
- **Remote-friendly:** Works over SSH
- **Scriptable:** Easy to automate with `tmux send-keys`
- **Ubiquitous:** Installed on most Unix systems

### Why 1:1 repo mapping?
- **Simplicity:** One session = one project
- **Predictability:** Session name always matches repo name
- **No conflicts:** Can't have multiple sessions for same repo
- **Mental model:** Easy to remember where things are

### Why two modes (native/iTerm2)?
- **Native mode:** Portable, works everywhere, full tmux features
- **iTerm2 mode:** Better macOS integration, native scrollback/search
- **User choice:** Different workflows prefer different approaches

### Why `$RIGS_BASE`?
- **Organization:** All repos in one place
- **Discovery:** Easy to scan for available projects
- **Isolation:** Doesn't interfere with other directories
- **Convention:** Matches common practice (`~/git/`, `~/projects/`)

---

## Limitations

### Current Limitations
1. **Single workspace per repo:** Can't have multiple tmux sessions for the same repo
2. **No nested rigs:** Can't manage rigs within rigs
3. **Tmux dependency:** Requires tmux to be installed
4. **Git dependency:** Only works with git repositories
5. **No remote rig support:** All repos must be local
6. **No session persistence config:** Layout is hardcoded (2 windows/panes)

### Known Issues
1. **Window/pane titles:** In native mode, automatic window renaming can interfere with titles
2. **Session name collisions:** If two repos have the same name in different parent dirs, only one can be managed
3. **No cleanup on repository deletion:** If you delete a repo, you must manually `rig down` the session

---

## Future Enhancements

### Planned Features
- [ ] **Crew support:** Multi-user workspaces with git worktrees (see crew design doc)
- [ ] **Custom layouts:** Per-repo configuration for window/pane setup
- [ ] **Hooks:** Pre/post scripts for rig up/down
- [ ] **Remote rigs:** SSH to remote machines and manage rigs there
- [ ] **Rig templates:** Pre-configured setups for different project types
- [ ] **Session restoration:** Save/restore exact session state

### Possible Future Directions
- [ ] **TUI dashboard:** Visual interface showing all rigs (like `lazygit`)
- [ ] **Rig groups:** Manage sets of related rigs together
- [ ] **Cloud sync:** Share rig state across machines
- [ ] **Plugin system:** Extensibility for custom workflows
- [ ] **Non-git support:** Manage sessions for non-repo directories

---

## Usage Patterns

### Daily Workflow
```bash
# Morning: Start work on notes project
rig up notes

# Switch to another project
rig switch myapp

# Check what's running
rig status

# End of day: shut down everything
rig killall
```

### iTerm2 Integration Workflow
```bash
# Set environment variable (add to ~/.zshrc for persistence)
export RIG_USE_CC=true

# Start rig - opens in new iTerm2 window
rig up notes

# Each rig gets its own native window
rig up myapp
rig up my-project

# Switch with Cmd+` (native macOS)
```

### Discovering Projects
```bash
# See what repos are available
rig list

# Start a rig from the list
rig up <name>
```

---

## Testing Checklist

### Basic Operations
- [ ] `rig up <name>` creates new session
- [ ] `rig up <name>` switches to existing session
- [ ] `rig down <name>` kills session
- [ ] `rig status` shows active rigs
- [ ] `rig list` shows available repos
- [ ] `rig switch <name>` switches to session
- [ ] `rig killall` kills all rigs

### Mode Switching
- [ ] Native mode creates 2 windows
- [ ] iTerm2 mode creates 2 panes
- [ ] iTerm2 mode opens new native window
- [ ] Can switch between modes

### Error Cases
- [ ] `rig up nonexistent` fails gracefully
- [ ] `rig down nonexistent` fails gracefully
- [ ] Repository without `.git/` is rejected
- [ ] Empty `$RIGS_BASE` is handled

### Edge Cases
- [ ] Works when already in tmux
- [ ] Works when outside tmux
- [ ] Session name with hyphens works
- [ ] Session name with underscores works
- [ ] Switching between rigs preserves state

---

## Installation

### Prerequisites
- Bash 4.0+
- tmux 2.0+
- git
- (Optional) iTerm2 3.0+ for integration mode
- (Optional) fzf for interactive switching

### Setup
```bash
# 1. Create script
cat > ~/bin/rig << 'EOF'
[paste rig script here]
EOF

# 2. Make executable
chmod +x ~/bin/rig

# 3. Ensure ~/bin is in PATH
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc

# 4. (Optional) Set RIGS_BASE
echo 'export RIGS_BASE="$HOME/git"' >> ~/.zshrc

# 5. (Optional) Enable iTerm2 mode
echo 'export RIG_USE_CC=true' >> ~/.zshrc
```

### Verification
```bash
# Check installation
which rig

# Try it out
rig list
```

---

## Configuration

### Environment Variables (Summary)
| Variable | Default | Description |
|----------|---------|-------------|
| `RIGS_BASE` | `~/git` | Base directory for repositories |
| `RIG_USE_CC` | `false` | Enable iTerm2 integration mode |

### Customization
Add to `~/.zshrc` or `~/.bashrc`:
```bash
# Use different base directory
export RIGS_BASE="$HOME/projects"

# Enable iTerm2 integration
export RIG_USE_CC=true

# Add aliases
alias rs="rig switch"
alias ru="rig up"
alias rd="rig down"
```

---

## Troubleshooting

### "Repo not found" error
- Check that directory exists in `$RIGS_BASE`
- Verify directory contains `.git/`
- Check `RIGS_BASE` is set correctly: `echo $RIGS_BASE`

### Session won't attach
- Verify tmux is running: `tmux ls`
- Try manually: `tmux attach -t <session-name>`
- Check for zombie sessions: `tmux kill-session -t <session-name>`

### Claude Code doesn't start
- Verify `claude` command is in PATH: `which claude`
- Check Claude Code is installed
- Manually attach and check error: `rig up <name>`, then check first window

### iTerm2 mode not working
- Verify iTerm2 is installed and running
- Check `RIG_USE_CC=true` is set
- Try manually: `tmux -CC attach`
- Ensure iTerm2 has tmux integration enabled

---

## Comparison to Alternatives

### vs. Manual tmux
**Pros:**
- Automates session setup
- Consistent layouts
- Repo discovery
- Easy switching

**Cons:**
- Less flexible
- Opinionated structure

### vs. tmuxinator/tmuxp
**Pros:**
- No config files needed
- Simpler mental model
- Automatic repo discovery
- Single command interface

**Cons:**
- Less customizable
- No per-project configs
- Fixed layout

### vs. IDE workspaces
**Pros:**
- Terminal-based (faster, lighter)
- Works over SSH
- Language/tool agnostic
- Scriptable

**Cons:**
- No GUI
- Requires tmux knowledge
- Less IDE integration

---

## Changelog

### v1.0 (Current)
- Initial implementation
- Support for native and iTerm2 modes
- Basic CRUD operations (up/down/status/list/switch/killall)
- Repository discovery from `$RIGS_BASE`

---

## Contributing

### Code Style
- Use shellcheck for linting
- Follow existing naming conventions
- Add comments for non-obvious logic
- Test on both native and iTerm2 modes

### Testing
- Test all commands manually
- Verify error cases
- Check both modes (native/iTerm2)
- Test on clean environment

---

## License

[Specify license]

---

## Credits

Inspired by:
- Gas Town by Steve Yegge (tmux-based agent orchestration)
- tmuxinator (tmux session management)
- The need for better development environment management
