package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func createTestRepo(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "testrepo")

	// Initialize git repo
	if err := os.Mkdir(repoPath, 0755); err != nil {
		t.Fatalf("Failed to create repo dir: %v", err)
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git
	exec.Command("git", "config", "user.name", "Test User").Run()
	exec.Command("git", "config", "user.email", "test@example.com").Run()

	// Create initial commit on main branch
	cmd = exec.Command("git", "checkout", "-b", "main")
	cmd.Dir = repoPath
	cmd.Run()

	testFile := filepath.Join(repoPath, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = repoPath
	cmd.Run()

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	return repoPath
}

func TestBranchExists(t *testing.T) {
	repoPath := createTestRepo(t)

	if !BranchExists(repoPath, "main") {
		t.Error("Expected main branch to exist")
	}

	if BranchExists(repoPath, "nonexistent") {
		t.Error("Expected nonexistent branch to not exist")
	}
}

func TestGetBaseBranch(t *testing.T) {
	repoPath := createTestRepo(t)

	t.Run("finds main branch", func(t *testing.T) {
		branch, err := GetBaseBranch(repoPath, "main")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if branch != "main" {
			t.Errorf("Expected main, got %s", branch)
		}
	})

	t.Run("falls back to master", func(t *testing.T) {
		// Create master branch
		cmd := exec.Command("git", "checkout", "-b", "master")
		cmd.Dir = repoPath
		cmd.Run()

		branch, err := GetBaseBranch(repoPath, "nonexistent")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if branch != "master" {
			t.Errorf("Expected master, got %s", branch)
		}
	})

	t.Run("errors when no base branch", func(t *testing.T) {
		// Create new repo with no main or master
		tmpDir := t.TempDir()
		newRepo := filepath.Join(tmpDir, "newrepo")
		os.Mkdir(newRepo, 0755)
		cmd := exec.Command("git", "init")
		cmd.Dir = newRepo
		cmd.Run()

		cmd = exec.Command("git", "checkout", "-b", "develop")
		cmd.Dir = newRepo
		cmd.Run()

		_, err := GetBaseBranch(newRepo, "main")
		if err == nil {
			t.Error("Expected error when no base branch exists")
		}
	})
}

func TestWorktreeOperations(t *testing.T) {
	repoPath := createTestRepo(t)
	worktreePath := filepath.Join(t.TempDir(), "worktree")

	t.Run("WorktreeExists returns false initially", func(t *testing.T) {
		if WorktreeExists(repoPath, worktreePath) {
			t.Error("Expected worktree to not exist")
		}
	})

	t.Run("CreateWorktree creates worktree", func(t *testing.T) {
		err := CreateWorktree(repoPath, worktreePath, "test/work", "main")
		if err != nil {
			t.Fatalf("Failed to create worktree: %v", err)
		}

		if !WorktreeExists(repoPath, worktreePath) {
			t.Error("Expected worktree to exist after creation")
		}

		if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
			t.Error("Expected worktree directory to exist")
		}
	})

	t.Run("RemoveWorktree removes worktree", func(t *testing.T) {
		err := RemoveWorktree(repoPath, worktreePath)
		if err != nil {
			t.Fatalf("Failed to remove worktree: %v", err)
		}

		if WorktreeExists(repoPath, worktreePath) {
			t.Error("Expected worktree to not exist after removal")
		}
	})
}

func TestGetCurrentBranch(t *testing.T) {
	repoPath := createTestRepo(t)

	branch, err := GetCurrentBranch(repoPath)
	if err != nil {
		t.Fatalf("Failed to get current branch: %v", err)
	}

	if branch != "main" && branch != "master" {
		t.Errorf("Expected main or master, got %s", branch)
	}
}

func TestCheckoutBranch(t *testing.T) {
	repoPath := createTestRepo(t)

	// Create a new branch
	cmd := exec.Command("git", "checkout", "-b", "feature")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create feature branch: %v", err)
	}

	// Go back to main
	cmd = exec.Command("git", "checkout", "main")
	cmd.Dir = repoPath
	cmd.Run()

	// Test checkout
	err := CheckoutBranch(repoPath, "feature")
	if err != nil {
		t.Fatalf("Failed to checkout branch: %v", err)
	}

	branch, _ := GetCurrentBranch(repoPath)
	if branch != "feature" {
		t.Errorf("Expected to be on feature branch, got %s", branch)
	}
}

func TestGetRepoRoot(t *testing.T) {
	repoPath := createTestRepo(t)

	// Create subdirectory
	subDir := filepath.Join(repoPath, "subdir")
	os.Mkdir(subDir, 0755)

	root, err := GetRepoRoot(subDir)
	if err != nil {
		t.Fatalf("Failed to get repo root: %v", err)
	}

	// Resolve symlinks for comparison (macOS uses /private/var instead of /var)
	expectedRoot, _ := filepath.EvalSymlinks(repoPath)
	actualRoot, _ := filepath.EvalSymlinks(root)

	if actualRoot != expectedRoot {
		t.Errorf("Expected repo root %s, got %s", expectedRoot, actualRoot)
	}
}

func TestIsGitRepo(t *testing.T) {
	repoPath := createTestRepo(t)

	if !IsGitRepo(repoPath) {
		t.Error("Expected directory to be a git repo")
	}

	nonRepo := t.TempDir()
	if IsGitRepo(nonRepo) {
		t.Error("Expected directory to not be a git repo")
	}
}

func TestDeleteBranch(t *testing.T) {
	repoPath := createTestRepo(t)

	// Create a branch
	cmd := exec.Command("git", "checkout", "-b", "todelete")
	cmd.Dir = repoPath
	cmd.Run()

	// Go back to main
	cmd = exec.Command("git", "checkout", "main")
	cmd.Dir = repoPath
	cmd.Run()

	// Delete the branch
	err := DeleteBranch(repoPath, "todelete")
	if err != nil {
		t.Fatalf("Failed to delete branch: %v", err)
	}

	if BranchExists(repoPath, "todelete") {
		t.Error("Expected branch to be deleted")
	}
}

func TestCreateFeatureBranch(t *testing.T) {
	repoPath := createTestRepo(t)

	err := CreateFeatureBranch(repoPath, "feat/new-feature", "main")
	if err != nil {
		t.Fatalf("Failed to create feature branch: %v", err)
	}

	if !BranchExists(repoPath, "feat/new-feature") {
		t.Error("Expected feature branch to exist")
	}

	branch, _ := GetCurrentBranch(repoPath)
	if branch != "feat/new-feature" {
		t.Errorf("Expected to be on feat/new-feature branch, got %s", branch)
	}
}

func TestListWorktrees(t *testing.T) {
	repoPath := createTestRepo(t)

	// Initially should have one worktree (main repo)
	worktrees, err := ListWorktrees(repoPath)
	if err != nil {
		t.Fatalf("Failed to list worktrees: %v", err)
	}

	if len(worktrees) != 1 {
		t.Errorf("Expected 1 worktree, got %d", len(worktrees))
	}

	// Create a new worktree
	worktreePath := filepath.Join(t.TempDir(), "wt1")
	err = CreateWorktree(repoPath, worktreePath, "test/branch", "main")
	if err != nil {
		t.Fatalf("Failed to create worktree: %v", err)
	}

	// List again
	worktrees, err = ListWorktrees(repoPath)
	if err != nil {
		t.Fatalf("Failed to list worktrees: %v", err)
	}

	if len(worktrees) != 2 {
		t.Errorf("Expected 2 worktrees, got %d", len(worktrees))
	}

	// Check the new worktree
	found := false
	for _, wt := range worktrees {
		if wt.Branch == "test/branch" {
			found = true
			// Resolve symlinks for comparison (macOS)
			expectedPath, _ := filepath.EvalSymlinks(worktreePath)
			actualPath, _ := filepath.EvalSymlinks(wt.Path)
			if actualPath != expectedPath {
				t.Errorf("Expected worktree path %s, got %s", expectedPath, actualPath)
			}
		}
	}

	if !found {
		t.Error("Expected to find test/branch worktree")
	}
}

func TestGetWorktreeForBranch(t *testing.T) {
	repoPath := createTestRepo(t)

	// Create a worktree
	worktreePath := filepath.Join(t.TempDir(), "wt_test")
	err := CreateWorktree(repoPath, worktreePath, "test/feature", "main")
	if err != nil {
		t.Fatalf("Failed to create worktree: %v", err)
	}

	// Find it
	path, err := GetWorktreeForBranch(repoPath, "test/feature")
	if err != nil {
		t.Fatalf("Failed to get worktree for branch: %v", err)
	}

	expectedPath, _ := filepath.EvalSymlinks(worktreePath)
	actualPath, _ := filepath.EvalSymlinks(path)
	if actualPath != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, actualPath)
	}

	// Try non-existent branch
	_, err = GetWorktreeForBranch(repoPath, "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent branch")
	}
}
