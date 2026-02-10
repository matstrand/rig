# Testing Documentation

This document describes all the tests in the rig Go implementation and what behaviors they verify.

## Test Summary

```bash
go test ./pkg/...
```

All tests pass:
- `pkg/config`: 5 tests
- `pkg/crew`: 4 test suites (15 total test cases)
- `pkg/git`: 9 test suites (12 total test cases)

## Config Package Tests

**File**: `pkg/config/config_test.go`

### TestLoad

Tests configuration loading from environment variables.

**Subtests**:
- `default_values`: Verifies default values when env vars are not set
  - `RIGS_BASE` defaults to `~/git`
  - `CREW_BASE` defaults to `~/crew`
  - `RIG_USE_CC` defaults to `false`
  - `RIG_DEFAULT_BRANCH` defaults to `main`

- `custom_values`: Verifies custom env vars are used
  - Tests all four environment variables with custom values
  - Ensures configuration system properly reads from environment

### TestGetRepoPath

Tests repo path generation.

**Behavior**: `GetRepoPath("myrepo")` returns `$RIGS_BASE/myrepo`

### TestGetCrewPath

Tests crew workspace path generation.

**Behavior**: `GetCrewPath("myrepo", "tracy")` returns `$CREW_BASE/myrepo/tracy`

**Critical**: This test verifies the correct directory structure (`<repo>/<name>`, not `<name>/<repo>`)

### TestGetCrewSessionName

Tests tmux session name generation for crew.

**Behavior**: `GetCrewSessionName("notes", "tracy")` returns `"notes@tracy"`

**Critical**: Verifies the `<rig>@<name>` naming convention

### TestGetCrewBranchName

Tests branch name generation for crew members.

**Behavior**: `GetCrewBranchName("tracy")` returns `"tracy/work"`

## Git Package Tests

**File**: `pkg/git/git_test.go`

All tests use real git operations on temporary repositories.

### TestBranchExists

Tests branch existence checking.

**Behavior**:
- Returns `true` for existing branch (`main`)
- Returns `false` for non-existent branch

### TestGetBaseBranch

Tests base branch detection with fallbacks.

**Subtests**:
- `finds_main_branch`: Returns `"main"` when main exists
- `falls_back_to_master`: Returns `"master"` when main doesn't exist but master does
- `errors_when_no_base_branch`: Returns error when neither main nor master exist

**Critical**: This ensures proper fallback behavior for different repo configurations

### TestWorktreeOperations

Tests the complete worktree lifecycle.

**Subtests**:
- `WorktreeExists_returns_false_initially`: Verifies new worktree doesn't exist yet
- `CreateWorktree_creates_worktree`: Creates worktree and verifies:
  - Directory is created
  - Git knows about the worktree
  - Branch is created
- `RemoveWorktree_removes_worktree`: Removes worktree and verifies it's gone

**Critical**: These are integration tests that verify actual git worktree behavior

### TestGetCurrentBranch

Tests getting the current branch in a directory.

**Behavior**: Returns the current branch name (main or master)

### TestCheckoutBranch

Tests branch checkout functionality.

**Behavior**:
1. Creates feature branch
2. Checks out main
3. Checks out feature using function
4. Verifies we're on feature branch

**Critical**: Verifies branch switching works correctly

### TestGetRepoRoot

Tests getting repository root from subdirectory.

**Behavior**:
- From subdirectory, returns the repository root
- Handles symlink resolution (important for macOS `/private/var` vs `/var`)

### TestIsGitRepo

Tests git repository detection.

**Behavior**:
- Returns `true` for directories with `.git`
- Returns `false` for regular directories

### TestDeleteBranch

Tests branch deletion.

**Behavior**:
1. Creates branch
2. Checks out main
3. Deletes branch
4. Verifies branch no longer exists

## Crew Package Tests

**File**: `pkg/crew/crew_test.go`

### TestValidateCrewName

Tests crew name validation rules.

**Valid names**:
- `"tracy"` - simple name
- `"tracy123"` - with numbers
- `"tracy-dev"` - with hyphen
- `"tracy_dev"` - with underscore

**Invalid names**:
- `""` - empty name
- `"tracy/dev"` - contains slash
- `"tracy\\dev"` - contains backslash
- `"tracy:dev"` - contains colon
- `"tracy@dev"` - contains at sign (reserved for session naming)
- `".tracy"` - starts with dot
- `"-tracy"` - starts with dash
- Long name > 50 chars - too long

**Critical**: Ensures @ character is rejected (conflicts with session naming)

### TestInferRig

Tests rig name inference from various contexts.

**Subtests**:

- `explicit_rig`: Explicit `--rig` flag takes precedence

- `from_rigs_directory`: Infers from `~/git/<repo>/`
  - Creates test repo in RIGS_BASE
  - Changes to repo directory
  - Infers rig name from git root

- `from_crew_directory`: Infers from `~/crew/<repo>/<name>/`
  - Creates test repo
  - Creates crew worktree
  - Changes to crew directory
  - **Critical**: Extracts rig name from path structure (not worktree name)

- `no_inference_possible`: Errors when not in rig/crew directory and no explicit flag

**Critical**: The crew directory test verifies we correctly parse the `~/crew/<repo>/<name>` structure

### TestCrewWorkflow

Integration test for complete crew workspace lifecycle.

**Note**: Skipped if tmux not available

**Subtests**:

- `add_crew_workspace`: Tests workspace creation
  - Validates crew name
  - Gets base branch
  - Creates worktree
  - Verifies directory exists
  - Verifies worktree registered in git
  - Verifies branch exists

- `verify_branch_is_correct`: Tests branch detection
  - Checks current branch in crew workspace
  - Verifies it matches expected branch name (`<name>/work`)

- `remove_crew_workspace`: Tests cleanup
  - Removes worktree
  - Prunes worktree metadata
  - Deletes branch
  - Verifies everything is cleaned up

**Critical**: This is an end-to-end integration test that verifies the entire workflow

### TestCrewPathStructure

Tests crew path structure correctness.

**Behavior**:
- `GetCrewPath("myrepo", "tracy")` returns `$CREW_BASE/myrepo/tracy`
- Parent directory is `$CREW_BASE/myrepo`
- Base name is `tracy`

**Critical**: Confirms the `~/crew/<repo>/<name>` structure is correct

## Test Execution

### Running All Tests

```bash
go test ./pkg/...
```

### Running with Verbose Output

```bash
go test -v ./pkg/...
```

### Running Specific Test

```bash
# Run single test
go test -v ./pkg/config -run TestLoad

# Run single subtest
go test -v ./pkg/crew -run TestInferRig/from_crew_directory
```

### Running with Coverage

```bash
go test -cover ./pkg/...
```

## Test Philosophy

### Real Behavior, Not Mocks

These tests use **real git operations** and **real file system operations**, not mocks:

- Git tests create actual repositories and worktrees
- Crew tests create real worktrees and verify git state
- Tests use `t.TempDir()` for automatic cleanup

### Why This Approach?

1. **Confidence**: Tests verify actual behavior, not mock expectations
2. **Regression prevention**: Real git behavior changes would be caught
3. **Integration validation**: Tests verify components work together
4. **Platform verification**: Tests catch platform-specific issues (like macOS symlinks)

### Trade-offs

- **Slower**: Real git operations take time (~3 seconds for git package)
- **Dependencies**: Requires git to be installed
- **Isolation**: Tests must be independent (using temp directories)

The slower execution is worth it for the confidence that the system actually works with real git.

## Platform-Specific Considerations

### macOS Symlinks

macOS uses `/private/var` which is symlinked to `/var`. Tests use `filepath.EvalSymlinks()` to handle this:

```go
expectedRoot, _ := filepath.EvalSymlinks(repoPath)
actualRoot, _ := filepath.EvalSymlinks(root)
```

This ensures path comparisons work across platforms.

### Temp Directory Cleanup

Tests use `t.TempDir()` which automatically cleans up after each test, preventing:
- Disk space issues
- Test pollution
- Permission problems

## Continuous Integration

These tests are suitable for CI/CD because:

1. **No manual intervention**: All tests are automated
2. **Clean state**: Each test creates its own temp directories
3. **Minimal dependencies**: Only requires Go and git
4. **Fast enough**: Complete suite runs in ~3 seconds

## Future Test Additions

Areas that could use more testing:

1. **Tmux operations**: Currently not tested (would require tmux running)
2. **Interactive prompts**: Crew operations that prompt for input
3. **Error scenarios**: More edge cases and error conditions
4. **Concurrent operations**: Multiple crew members on same repo
5. **Session attachment**: Full end-to-end with tmux

## Test Maintenance

When adding new features:

1. **Add tests first** (TDD approach)
2. **Test actual behavior** (not implementation details)
3. **Use real operations** (avoid mocks where possible)
4. **Clean up resources** (use t.TempDir, defer cleanup)
5. **Handle platform differences** (symlinks, paths)
6. **Document what you're testing** (clear test names, comments)
