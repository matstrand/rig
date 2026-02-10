# Design: Workflow System

## Overview

Add work-based development workflow to `rig` with git-native state tracking, formula-based workflows, and ephemeral worker (polecat) support.

## Architecture

### New Package: `pkg/work`

**Purpose:** Manage work directories, formulas, progress tracking, and hook generation.

**Key Types:**
```go
type Work struct {
    Name        string  // e.g., "build-frontend"
    Path        string  // e.g., "work/build-frontend"
    FeatureName string  // e.g., "feat/build-frontend"
}

type Progress struct {
    Status      string   // e.g., "In Design"
    AssignedTo  string   // e.g., "polecat_emma" (inferred from worktree)
    Tasks       []Task
    Notes       string
}

type Task struct {
    Done        bool
    Description string
}
```

**Key Functions:**
- `Create(repoPath, workName string) error` - Create work directory structure
- `ParseProgress(path string) (*Progress, error)` - Parse progress.md
- `GenerateHook(work *Work, formula string) error` - Generate hook.md
- `ListWorks(repoPath string) ([]*Work, error)` - List work directories
- `InferWorkFromBranch(branchName string) string` - Extract work name from branch

### New Package: `pkg/polecat`

**Purpose:** Manage ephemeral worker naming and lifecycle.

**Key Functions:**
- `GenerateName() string` - Generate random polecat name
- `IsPolecat(name string) bool` - Check if name is polecat format
- `List(cfg *config.Config) ([]string, error)` - List all polecats

**Name Pool:**
```go
var names = []string{
    "emma", "olivia", "ava", "sophia", "mia", "charlotte",
    "amelia", "harper", "evelyn", "abigail", "ella", "scarlett",
    "grace", "chloe", "lily", "zoe", "maya", "lucy",
}
```

### Extended Package: `pkg/git`

**New Functions:**
- `CreateFeatureBranch(repoPath, branchName, baseBranch string) error`
- `ListWorktrees(repoPath string) ([]Worktree, error)`
- `GetWorktreeForBranch(repoPath, branchName string) (string, error)`

### Extended Package: `pkg/config`

**New Config Methods:**
- `GetWorkDir(repoPath, workName string) string` - Return work/<name>
- `GetFormulaPath(repoPath, formula string) string` - Return work/formula/<name>.md

### New Commands in `main.go`

1. **`rig work create <name>`**
   - Create work directory with spec.md, design.md, breakdown.md, progress.md
   - Create feature branch feat/<name>
   - Ensure work/formula/ directory exists
   - Install default formula if missing
   - Create initial commit

2. **`rig work status`**
   - Scan all rigs in ~/crew/
   - For each worktree, read branch and parse progress.md
   - Display formatted status with rig, work, status, assignee, current task

3. **`rig sling <work-path> [flags]`**
   - Flags: `--to=<name>`, `--formula=<name>`, `--self`
   - Generate hook.md
   - Create polecat or assign to existing crew
   - For polecat: create workspace, start session, send initial message
   - For crew/self: just create hook and provide instructions

4. **`rig hook`**
   - Read and display hook.md from current work directory
   - Infer work name from current branch

5. **`rig crew prune [--polecats]`**
   - List polecats
   - Prompt for confirmation
   - Remove workspaces and worktrees

## Files to Create/Modify

### New Files
- `rig/go/pkg/work/work.go` - Core work management
- `rig/go/pkg/work/work_test.go` - Unit tests
- `rig/go/pkg/polecat/polecat.go` - Polecat naming
- `rig/go/pkg/polecat/polecat_test.go` - Unit tests
- `rig/go/templates/spec.md` - Default spec template
- `rig/go/templates/design.md` - Default design template
- `rig/go/templates/breakdown.md` - Default breakdown template
- `rig/go/templates/progress.md` - Default progress template
- `rig/go/templates/formula_build.md` - Default build formula

### Modified Files
- `rig/go/cmd/rig/main.go` - Add new commands
- `rig/go/pkg/git/git.go` - Add feature branch and worktree listing
- `rig/go/pkg/config/config.go` - Add work directory helpers
- `rig/go/pkg/crew/crew.go` - Extend for polecat support

## Testing Strategy

### Unit Tests
- `work_test.go`: Test parsing progress.md, generating hooks, listing works
- `polecat_test.go`: Test name generation, validation
- `git_test.go`: Test new git functions

### Integration Tests
- Create work, verify file structure
- Sling to polecat, verify workspace creation
- Parse progress from real file
- Hook generation and display

### Manual Testing
- Full workflow: create → sling → work → status → cleanup
- Edge cases: existing work, reassignment, missing formulas
- Cross-rig status display

## Risk Areas

### 1. Git Worktree Management
**Risk:** Worktrees can get into inconsistent states (detached, deleted directories)
**Mitigation:** Reuse existing crew.Remove patterns with cleanup and pruning

### 2. Progress.md Parsing
**Risk:** Markdown parsing fragility, different checkbox formats
**Mitigation:** Simple regex-based parser, tolerate variations, fail gracefully

### 3. Polecat Name Collisions
**Risk:** Generated name already exists
**Mitigation:** Check existing names before generation, retry with different name

### 4. Concurrent Modifications
**Risk:** Multiple agents modifying progress.md simultaneously
**Mitigation:** Document as out of scope, rely on git merge conflict resolution

### 5. Formula Not Found
**Risk:** User specifies non-existent formula
**Mitigation:** Validate formula exists before slinging, list available formulas in error

## Implementation Order

1. **Phase 1:** Core work package (create, list, parse progress)
2. **Phase 2:** Polecat package (name generation)
3. **Phase 3:** `rig work create` command
4. **Phase 4:** `rig hook` command (simple file read)
5. **Phase 5:** `rig sling` command (complex, many moving parts)
6. **Phase 6:** `rig work status` command (requires parsing across repos)
7. **Phase 7:** `rig crew prune` command

## Open Questions

1. Should progress.md be validated on parse (strict schema) or tolerate variations?
   **Decision:** Tolerate variations, extract what we can, warn on parse errors

2. Should hook.md be regenerated if formula changes?
   **Decision:** No, hook is static once created. Users can manually edit or delete and re-sling.

3. Should we validate work directory structure on sling?
   **Decision:** Yes, warn if spec.md missing but don't block.

4. Should polecats auto-cleanup on completion?
   **Decision:** No, manual cleanup with `rig crew remove`. Final checklist reminds user.
