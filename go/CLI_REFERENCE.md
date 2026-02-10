# CLI Reference

Complete command-line interface for the rig Go implementation.

## Main Commands

### rig up

Bring up a rig (creates or switches to existing session).

```bash
rig up <name>
```

**Examples**:
```bash
rig up notes              # Start/switch to notes rig
rig up myapp        # Start/switch to myapp rig
```

**Behavior**:
- If session exists: switches to it
- If session doesn't exist: creates it and attaches
- Creates 2 tmux windows: "Claude Code" and "Terminal"
- Starts `claude` in first window
- Runs `git status` in second window

---

### rig down

Shut down a rig session.

```bash
rig down <name>
```

**Examples**:
```bash
rig down notes            # Kill notes session
rig down myapp      # Kill myapp session
```

**Behavior**:
- Kills the tmux session
- Errors if session doesn't exist

---

### rig status / rig ls

Show all active rigs and crew sessions.

```bash
rig status
rig ls        # alias
```

**Output**:
```
=== Active Rigs ===

✓ notes
  └─ /Users/user/git/notes

  myapp
  └─ /Users/user/git/myapp

=== Crew ===

  notes@tracy (notes/tracy)
  └─ /Users/user/crew/notes/tracy
```

**Behavior**:
- Lists all rig sessions
- Lists all crew sessions
- Shows ✓ for current session
- Shows paths for each

---

### rig list

List available repos in RIGS_BASE.

```bash
rig list
```

**Output**:
```
=== Available Repos in /Users/user/git ===

  notes (running)
  myapp
  myproject

Total: 3 repos
```

**Behavior**:
- Scans RIGS_BASE for git repositories
- Shows which ones have active sessions
- Counts total repos

---

### rig switch

Switch to a rig or crew session.

```bash
rig switch <name>
```

**Examples**:
```bash
rig switch notes          # Switch to notes rig
rig switch notes@tracy    # Switch to crew session
```

**Behavior**:
- If in tmux: switches client
- If not in tmux: attaches to session
- Errors if session doesn't exist

---

### rig killall

Shut down multiple sessions.

```bash
rig killall [flags]
```

**Flags**:
- `--crew`: Kill both rigs and crew
- `--crew-only`: Kill only crew sessions

**Examples**:
```bash
rig killall               # Kill all rig sessions
rig killall --crew        # Kill all rigs and crew
rig killall --crew-only   # Kill only crew sessions
```

**Behavior**:
- Default: kills only rig sessions
- `--crew`: kills rigs and crew
- `--crew-only`: kills only crew sessions

---

## Crew Commands

### rig crew add

Create a new crew workspace.

```bash
rig crew add <name> [--rig=<repo>]
```

**Flags**:
- `--rig=<repo>`: Explicit repo name (optional, can be inferred)

**Examples**:
```bash
# From within ~/git/notes
rig crew add tracy        # Creates ~/crew/notes/tracy

# Explicit rig
rig crew add tracy --rig=notes    # Works from anywhere

# From within a crew workspace
cd ~/crew/notes/tracy
rig crew add alex         # Infers rig from git root
```

**Behavior**:
1. Validates crew name (no @, /, etc.)
2. Infers or uses explicit rig
3. Creates git worktree at `~/crew/<rig>/<name>`
4. Creates branch `<name>/work` from `main`
5. Creates tmux session `<rig>@<name>`
6. Starts `claude` in first window
7. Attaches to session

**Idempotent**: If workspace exists, recreates session and attaches

**Interactive**:
- Prompts if branch already exists: "Use existing branch? [Y/n]"

---

### rig crew start

Attach to an existing crew workspace.

```bash
rig crew start <name> [--rig=<repo>]
```

**Flags**:
- `--rig=<repo>`: Explicit repo name (optional, can be inferred)

**Examples**:
```bash
rig crew start tracy      # Attach to tracy's workspace
rig crew start tracy --rig=notes    # Explicit rig
```

**Behavior**:
1. Checks workspace exists
2. Verifies on correct branch
3. Creates session if doesn't exist
4. Attaches to session

**Interactive**:
- Prompts if on wrong branch: "Switch to <name>/work? [Y/n]"

---

### rig crew remove / rig crew rm

Remove a crew workspace.

```bash
rig crew remove <name> [--rig=<repo>]
rig crew rm <name> [--rig=<repo>]       # alias
```

**Flags**:
- `--rig=<repo>`: Explicit repo name (optional, can be inferred)

**Examples**:
```bash
rig crew remove tracy     # Remove tracy's workspace
rig crew rm alex          # Same, shorter alias
```

**Behavior**:
1. Warns if in current session
2. Asks about branch deletion
3. Kills tmux session
4. Removes git worktree
5. Prunes worktree metadata
6. Deletes branch if confirmed
7. Removes empty repo directory

**Interactive**:
- Prompts: "Delete branch <name>/work? [Y/n]"
- Warns: "You are currently in session '...' - removing it will disconnect you"

**Edge cases handled**:
- Detached worktree (git knows about it but directory gone)
- Session exists but worktree doesn't
- Worktree exists but session doesn't

---

### rig crew ls / rig crew list

List crew workspaces.

```bash
rig crew ls [name]
rig crew list [name]      # alias
```

**Arguments**:
- `name`: Optional filter by crew member name

**Examples**:
```bash
rig crew ls               # List all crew workspaces

# Output:
# === Crew Workspaces ===
#
# tracy
#   - notes (notes@tracy) [running]
#   - myapp (myapp@tracy) [stopped]
#
# alex
#   - notes (notes@alex) [running]

rig crew ls tracy         # List only tracy's workspaces
```

**Behavior**:
- Groups by crew member name
- Shows all repos for each member
- Shows session status (running/stopped)
- Shows session name

---

### rig crew status

Show active crew sessions.

```bash
rig crew status
```

**Output**:
```
=== Active Crew Sessions ===

notes@tracy               /Users/user/crew/notes/tracy       tracy/work      [running]
notes@alex                /Users/user/crew/notes/alex        alex/work       [running]
myapp@tracy         /Users/user/crew/myapp/tracy tracy/work      [running]
```

**Behavior**:
- Lists only running crew sessions
- Shows session name, path, and current branch
- Only shows sessions with existing workspaces

---

## Environment Variables

### RIGS_BASE

Base directory for main repositories.

```bash
export RIGS_BASE="$HOME/git"      # default
export RIGS_BASE="/custom/path"   # custom
```

### CREW_BASE

Base directory for crew workspaces.

```bash
export CREW_BASE="$HOME/crew"     # default
export CREW_BASE="/custom/path"   # custom
```

### RIG_USE_CC

Enable iTerm2 integration mode.

```bash
export RIG_USE_CC="false"         # default (native tmux)
export RIG_USE_CC="true"          # iTerm2 mode
```

**Differences**:
- Native mode: 2 windows (Claude Code, Terminal)
- iTerm2 mode: 1 window with 2 panes (Claude Code | Terminal)

### RIG_DEFAULT_BRANCH

Default branch for crew worktrees.

```bash
export RIG_DEFAULT_BRANCH="main"     # default
export RIG_DEFAULT_BRANCH="master"   # custom
export RIG_DEFAULT_BRANCH="develop"  # custom
```

**Behavior**:
- Tries RIG_DEFAULT_BRANCH first
- Falls back to "master" if not found
- Errors if neither exists

---

## Session Naming Convention

### Rig Sessions

Format: `<repo-name>`

Examples:
- `notes`
- `myapp`
- `myproject`

### Crew Sessions

Format: `<rig>@<name>`

Examples:
- `notes@tracy`
- `notes@alex`
- `myapp@tracy`

**Why @ separator?**
- Safe in tmux (no syntax conflicts)
- Clear semantic meaning (rig "has" crew member)
- Easy to parse: `${session%%@*}` = rig, `${session##*@}` = name
- No collision with regular rig sessions

---

## Directory Structure

```
~/git/                      # RIGS_BASE
  ├── notes/                # main repo
  ├── myapp/          # main repo
  └── myproject/            # main repo

~/crew/                     # CREW_BASE
  ├── notes/                # repo directory
  │   ├── tracy/            # crew member workspace (git worktree)
  │   └── alex/             # crew member workspace (git worktree)
  └── myapp/          # repo directory
      └── tracy/            # crew member workspace (git worktree)
```

**Key points**:
- Main repos in `~/git/<repo>`
- Crew workspaces in `~/crew/<repo>/<name>`
- Each crew workspace is a git worktree
- Each crew member gets a branch: `<name>/work`

---

## Rig Inference

The system can automatically infer which rig you're working with:

### 1. Explicit Flag (Highest Priority)

```bash
rig crew add tracy --rig=notes
```

### 2. Current Directory in RIGS_BASE

```bash
cd ~/git/notes
rig crew add tracy        # Infers: notes
```

### 3. Current Directory in CREW_BASE

```bash
cd ~/crew/notes/tracy
rig crew add alex         # Infers: notes (from path structure)
```

### 4. Active Tmux Session

#### From rig session:
```bash
# In tmux session "notes"
rig crew add tracy        # Infers: notes
```

#### From crew session:
```bash
# In tmux session "notes@tracy"
rig crew add alex         # Infers: notes (from session name)
```

### Error When No Inference Possible

```bash
cd /tmp
rig crew add tracy
# Error: could not infer rig. Use --rig=<repo> or run from within a repo
```

---

## Error Handling

### Crew Name Validation

Invalid crew names are rejected:

```bash
rig crew add tracy@dev    # Error: cannot contain @
rig crew add tracy/dev    # Error: cannot contain /
rig crew add .tracy       # Error: cannot start with .
rig crew add -tracy       # Error: cannot start with -
```

Valid: `tracy`, `tracy-dev`, `tracy_dev`, `tracy123`

### Repo Not Found

```bash
rig up nonexistent
# Error: repo not found: /Users/user/git/nonexistent
```

### Session Not Found

```bash
rig down nonexistent
# Error: rig not found: nonexistent
```

### Workspace Not Found

```bash
rig crew start tracy
# Error: crew workspace not found: /Users/user/crew/notes/tracy
# Use 'rig crew add tracy --rig=notes' first
```

---

## Tips & Tricks

### Quick Crew Creation

```bash
# From anywhere
rig crew add tracy --rig=notes

# Or cd first
cd ~/git/notes && rig crew add tracy
```

### List Everything

```bash
rig status        # Shows all rigs and crew
rig crew ls       # Shows all crew workspaces
rig crew status   # Shows only running crew
```

### Cleanup

```bash
# Remove single crew member
rig crew rm tracy

# Remove all crew sessions (keeps workspaces)
rig killall --crew-only

# Remove all crew workspaces manually
rm -rf ~/crew/notes/*
```

### Multiple Rigs

```bash
# Tracy can work on multiple repos
rig crew add tracy --rig=notes
rig crew add tracy --rig=myapp

# List shows both
rig crew ls tracy
# tracy
#   - notes (notes@tracy) [stopped]
#   - myapp (myapp@tracy) [stopped]
```

---

## Comparison with Bash Version

The Go implementation maintains **100% CLI compatibility** with the bash version:

- Same commands and flags
- Same output format
- Same directory structure
- Same session naming
- Same rig inference logic
- Same interactive prompts

**Differences**:
- Better error messages (structured errors)
- Faster execution (compiled binary)
- Better testability (unit tests)
- Cross-platform support (works on Windows with tmux)

---

## Help System

Every command has built-in help:

```bash
rig --help
rig up --help
rig crew --help
rig crew add --help
```

Use `-h` as shorthand:

```bash
rig -h
rig crew add -h
```
