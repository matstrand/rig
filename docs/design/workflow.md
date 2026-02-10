# Spec: Work-Based Development Workflow

## Overview

Enable structured, document-driven feature development with git-native state tracking and crew-based work assignment.

## Problem

Currently, `rig` manages tmux sessions and crew workspaces, but provides no structured way to:
- Define and track feature work across its lifecycle
- Assign work to crew members or ephemeral workers
- Maintain visibility into active work across feature branches
- Follow a consistent workflow from spec → design → implementation → review

## Goals

1. Provide a convention-based system for defining work in git-tracked documents
2. Enable work assignment to crew members and ephemeral polecats
3. Track progress through simple markdown checklists
4. Maintain full visibility into active work across all feature branches
5. Keep everything git-native (no external databases or state files outside the repo)

## Non-Goals

- Complex workflow engines or orchestration systems
- Integration with external project management tools
- Automated merging or CI/CD workflows
- Multi-repo or cross-repo work tracking

## User Experience

### Creating New Work
```bash
$ rig work create build-frontend

✓ Created work directory: work/build-frontend/
✓ Created feature branch: feat/build-frontend
✓ Scaffolded files: spec.md, design.md, breakdown.md, progress.md
✓ Ensured formulas directory exists: work/formula/
✓ Installed default formula: work/formula/build.md (if not exists)
✓ Initial commit: "Initialize work: build-frontend"

Next steps:
  1. Edit work/build-frontend/spec.md
  2. When ready: rig sling work/build-frontend
```

**Behavior:**
- Warns if `work/build-frontend/` already exists but continues
- Creates only missing files (doesn't overwrite existing ones)
- Ensures `work/formula/` directory exists
- Installs formula files that don't exist yet (starting with `build.md`)
- Never overwrites existing formulas

### Work Directory Structure

All feature work lives in `work/<feature-name>/`:
```
work/
├── formula/
│   ├── build.md         # Default build workflow (copied from rig/build.md reference)
│   └── hotfix.md        # Optional: fast-track workflow for urgent fixes
│
└── build-frontend/
    ├── spec.md          # What we're building (requirements, context)
    ├── design.md        # How we'll build it (architecture, approach)
    ├── breakdown.md     # Implementation steps (tasks, subtasks)
    ├── progress.md      # Current status and checklist
    └── hook.md          # Startup instructions for worker (created by rig sling)
```

**Multiple works per repo:**
- You can have many feature directories under `work/`
- Only one is active per feature branch
- Active work matches the branch name: `feat/build-frontend` → `work/build-frontend/`

### Progress Tracking

`progress.md` contains a simple checklist that defines the workflow:
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
- Status field updated manually or by agents
- Checklist can evolve as work progresses
- All changes committed to git on the feature branch at each intermediate phase
- The agent carries on automatically between phases
- Final checklist items remind user to push and cleanup crew workspace

This checklist must be updated as work progresses. Tasks can be added or checked off.

### Hook File

When work is slung to a worker, `rig sling` creates a `hook.md` file in the work directory that provides startup instructions:
```markdown
# Hook: build-frontend

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

**Hook behavior:**
- Created once when work is slung
- Static - does not change after creation
- Simple instructions: review spec and follow the formula
- Used for initial context and recovery after restarts
- Agent checks hook on startup via `rig hook` command
- Persists on feature branch in git

### Slinging Work (Assignment)
```bash
# Assign to named crew member
$ rig sling work/build-frontend --to=tracy

✓ Created hook: work/build-frontend/hook.md
✓ Workspace ready: ~/crew/myapp/tracy
✓ Branch checked out: feat/build-frontend

To start working, paste this into tracy's Claude Code session:
  Check your hook: rig hook

# Assign to ephemeral polecat (auto-named, uses default formula)
$ rig sling work/build-frontend

✓ Created polecat: polecat_emma
✓ Created hook: work/build-frontend/hook.md
✓ Workspace: ~/crew/myapp/polecat_emma
✓ Session: myapp-polecat_emma
✓ Branch: feat/build-frontend

Session started. Claude Code will check hook automatically.

# Use a specific formula instead of default
$ rig sling work/build-frontend --formula=hotfix

✓ Created polecat: polecat_maya
✓ Created hook: work/build-frontend/hook.md (using work/formula/hotfix.md)
✓ Workspace: ~/crew/myapp/polecat_maya
✓ Session: myapp-polecat_maya
✓ Branch: feat/build-frontend

Session started. Claude Code will check hook automatically.

# Work on it yourself in current session
$ rig sling work/build-frontend --self

✓ Created hook: work/build-frontend/hook.md
✓ Hook ready in current workspace

To start working, paste this into your Claude Code session:
  Check your hook: rig hook
```

**What happens during sling:**

1. Create `hook.md` file with:
   - Path to formula (`work/formula/<formula_name>.md`, via `--formula=<name>`, default: build)
   - Path to spec file (`work/<feature-name>/spec.md`)
   - Links to other context files (design.md, breakdown.md, progress.md)
   - Instructions to commit intermediate progress
2. If assigning to new polecat:
   - Create crew workspace
   - Checkout feature branch via git worktree
   - Start tmux session with Claude Code
   - Send initial message: `"Check your hook: rig hook and follow instructions there"`
3. If assigning to existing crew or `--self`:
   - Create/update hook.md
   - Provide copy-paste instruction for user to give to Claude Code
   - Do NOT send keys to existing session

**Re-slinging behavior:**
- If work is already assigned (worktree exists), warn and ask for confirmation
- Show current assignee and ask if user wants to reassign

### The `rig hook` Command
```bash
# Display current hook (reads from work directory in current branch)
$ rig hook

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

The `rig hook` command simply reads and displays the `hook.md` file. Agents are prompted to run this on startup to get their instructions.

### Viewing Work Status
```bash
$ rig work status

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

**How it works:**
- Scans all rigs in `~/crew/`
- For each rig, scans all crew/polecat workspaces
- Reads the git branch and checks for `work/<feature-name>/progress.md` from each workspace
- Feature name derived from branch: `feat/build-frontend` → `work/build-frontend/`
- Parses status, current task, and assignee from the workspace directory name
- Displays consolidated view across all active work
- No agent/polecat needed - just filesystem and git operations

### Managing Polecats

Polecats are ephemeral workers for one-off tasks. They live in crew workspaces with auto-generated names.
```bash
# List all crew including polecats
$ rig crew ls

Crew (myapp):
  tracy                  feat/optimize-queries    [crew]
  sam                    feat/add-logging        [crew]
  polecat_emma          feat/build-frontend      [polecat]
  polecat_sofia         feat/fix-auth-bug        [polecat]

# Attach to review polecat's work
$ rig crew attach polecat_emma

# Remove when done (also removes git worktree)
$ rig crew remove polecat_emma

✓ Removed crew workspace: ~/crew/myapp/polecat_emma
✓ Removed git worktree

# Bulk cleanup
$ rig crew prune --polecats

Found polecats: polecat_emma, polecat_sofia
Remove these workspaces and worktrees? (y/N)
```

**Polecat naming:**
- Random female names from a predefined pool
- Format: `polecat_<name>`
- Names recycled after removal
- Visually distinguished in `rig crew ls` output

**Cleanup behavior:**
- `rig crew remove` deletes both the crew workspace directory and the git worktree
- Progress.md includes reminder to cleanup when work is complete
- Polecats do not auto-cleanup when work finishes

## Conventions

### Directory Naming
- Work lives in: `work/<feature-name>/`
- Feature branches: `feat/<feature-name>`
- Names should be kebab-case (e.g., `build-frontend`)

### Crew Workspaces
- Base directory: `~/crew/<repo-name>/`
- Named crew: `~/crew/<repo>/<name>/`
- Polecats: `~/crew/<repo>/polecat_<name>/`
- Each workspace is a git worktree of the feature branch

### State Management
- All state tracked in `progress.md` on feature branch
- Hook instructions tracked in `hook.md` on feature branch
- No hidden files or external databases
- Git is the source of truth
- Assignee inferred from git worktree ownership

### Formula System

Formulas define reusable workflows. They live in `work/formula/` and can be referenced by `rig sling`:
```bash
# Use default formula (work/formula/build.md)
$ rig sling work/build-frontend

# Use specific formula
$ rig sling work/build-frontend --formula=hotfix
# Uses work/formula/hotfix.md
```

**Formula reference:**
- Default formula (`build.md`) is based on `rig/build.md` in the rig repository
- Formulas follow design principles:
  - Spec review
  - Design
  - Design review
  - Implementation breakdown
  - Implementation
  - Code review
  - Fixes
  - Push/merge

**Formula requirements:**
The formula must emphasize that:
- Intermediate progress should be committed at each step
- The checklist in progress.md must be kept up to date
- Each phase should leave the work in a consistent state

## Error Handling & Edge Cases

### Existing Work Directory
```bash
$ rig work create build-frontend

⚠️  Warning: work/build-frontend/ already exists
✓ Skipping existing files
✓ Created missing files: breakdown.md
✓ Ensured formulas directory exists
✓ Formula work/formula/build.md already exists (not overwriting)
```

### Existing Feature Branch
```bash
$ rig work create build-frontend

⚠️  Warning: Branch feat/build-frontend already exists
✓ Using existing branch
✓ Created work directory: work/build-frontend/
```

### Re-slinging Assigned Work
```bash
$ rig sling work/build-frontend

⚠️  Warning: work/build-frontend is already assigned to polecat_emma
   Workspace: ~/crew/myapp/polecat_emma

Reassign to new polecat? (y/N)
```

### Missing Progress File
```bash
$ rig work status

⚠️  Warning: No progress.md found in ~/crew/myapp/polecat_emma
   Skipping polecat_emma
```

### Formula Not Found
```bash
$ rig sling work/build-frontend --formula=custom

❌ Error: Formula not found: work/formula/custom.md
   Available formulas: build
```

## Open Questions & Decisions

### 1. Should `progress.md` explicitly track assignee, or infer from git worktree?
**Decision:** Infer from git worktree. The worktree owner is the assignee.

### 2. How should context be delivered to agents? Tmux send-keys? File-based handoff?
**Decision:** Hybrid approach:
- `rig sling` writes `hook.md` file with full context
- For new polecats: tmux send-keys with `"Check your hook: rig hook"`
- For existing crew/self: provide copy-paste instruction for user

### 3. Should there be `rig work done <name>` to mark work complete and archive the branch?
**Decision:** Yes, this is a good idea for future implementation.

### 4. Should `rig sling` support auto-assignment based on crew availability/workload?
**Decision:** No. `rig sling` always creates a new polecat unless `--to=<name>` is specified.

### 5. What happens if you sling work that's already assigned? Reassign? Warn?
**Decision:** Warn the user that work is already assigned. Ask for confirmation before reassigning.

### 6. What happens when a polecat completes work?
**Decision:** Nothing automatic. User manually runs `rig crew remove <polecat>` after reviewing work. The final progress.md checklist includes a reminder to cleanup.

### 7. How are formulas created and managed?
**Decision:** Start with `build.md` based on reference in `rig/build.md`. Formulas are extensible - users can add their own. `rig work create` installs missing formulas but never overwrites existing ones.

## Success Criteria

- User can create new work with a single command
- Work state is always visible in git (no hidden state)
- Work can be assigned to polecats seamlessly
- `rig work status` provides clear overview of all active work across all rigs
- Polecats can be easily created, reviewed, and cleaned up
- Hook system provides clear, persistent instructions for workers
- No dependency on external tools beyond git and tmux
- Fully unit tested
- CLI interfaces are nicely formatted with appropriate unicode and emoji (🪝 for hook, ✓ for success, ⚠️ for warnings, 📦 for rigs, etc.)
- Graceful handling of edge cases (existing files, branches, assignments)
- Formula system is extensible without code changes

## Future Enhancements (Out of Scope)

- Work templates for different types of features
- Automatic progress.md updates based on git activity
- Integration with PR/merge workflows
- Cross-rig work coordination
- Work metrics and velocity tracking
- `rig work done` command to mark work complete
- Formula marketplace or sharing system
