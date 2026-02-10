package crew

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/mstrand/rig/pkg/config"
	"github.com/mstrand/rig/pkg/git"
)

func TestValidateCrewName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		shouldErr bool
	}{
		{"valid name", "tracy", false},
		{"valid with numbers", "tracy123", false},
		{"valid with hyphen", "tracy-dev", false},
		{"valid with underscore", "tracy_dev", false},
		{"empty name", "", true},
		{"with slash", "tracy/dev", true},
		{"with backslash", "tracy\\dev", true},
		{"with colon", "tracy:dev", true},
		{"with at sign", "tracy@dev", true},
		{"starts with dot", ".tracy", true},
		{"starts with dash", "-tracy", true},
		{"too long", "this-is-a-very-long-crew-name-that-exceeds-the-fifty-character-limit", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCrewName(tt.input)
			if tt.shouldErr && err == nil {
				t.Errorf("Expected error for %s, got nil", tt.input)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Expected no error for %s, got %v", tt.input, err)
			}
		})
	}
}

func createTestGitRepo(t *testing.T, basePath, name string) string {
	t.Helper()

	repoPath := filepath.Join(basePath, name)
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("Failed to create repo dir: %v", err)
	}

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = repoPath
	cmd.Run()

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = repoPath
	cmd.Run()

	// Create initial commit on main branch
	cmd = exec.Command("git", "checkout", "-b", "main")
	cmd.Dir = repoPath
	cmd.Run()

	testFile := filepath.Join(repoPath, "README.md")
	if err := os.WriteFile(testFile, []byte("# Test"), 0644); err != nil {
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

func setupTestConfig(t *testing.T) *config.Config {
	t.Helper()

	tmpDir := t.TempDir()

	rigsBase := filepath.Join(tmpDir, "git")
	crewBase := filepath.Join(tmpDir, "crew")

	os.MkdirAll(rigsBase, 0755)
	os.MkdirAll(crewBase, 0755)

	return &config.Config{
		RigsBase:      rigsBase,
		CrewBase:      crewBase,
		UseCC:         false,
		DefaultBranch: "main",
	}
}

func TestInferRig(t *testing.T) {
	cfg := setupTestConfig(t)

	t.Run("explicit rig", func(t *testing.T) {
		rig, err := InferRig(cfg, "explicit-rig")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if rig != "explicit-rig" {
			t.Errorf("Expected explicit-rig, got %s", rig)
		}
	})

	t.Run("from rigs directory", func(t *testing.T) {
		// Create a test repo
		repoPath := createTestGitRepo(t, cfg.RigsBase, "testrepo")

		// Resolve symlinks for both paths (macOS issue)
		resolvedRigsBase, _ := filepath.EvalSymlinks(cfg.RigsBase)
		resolvedRepoPath, _ := filepath.EvalSymlinks(repoPath)

		// Update config with resolved path
		testCfg := *cfg
		testCfg.RigsBase = resolvedRigsBase

		// Change to repo directory
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		os.Chdir(resolvedRepoPath)

		rig, err := InferRig(&testCfg, "")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if rig != "testrepo" {
			t.Errorf("Expected testrepo, got %s", rig)
		}
	})

	t.Run("from crew directory", func(t *testing.T) {
		// Create a test repo
		createTestGitRepo(t, cfg.RigsBase, "testrepo2")

		// Resolve symlinks
		resolvedRigsBase, _ := filepath.EvalSymlinks(cfg.RigsBase)
		resolvedCrewBase, _ := filepath.EvalSymlinks(cfg.CrewBase)

		// Update config with resolved paths
		testCfg := *cfg
		testCfg.RigsBase = resolvedRigsBase
		testCfg.CrewBase = resolvedCrewBase

		// Create a crew worktree
		crewPath := testCfg.GetCrewPath("testrepo2", "tracy")
		os.MkdirAll(crewPath, 0755)

		repoPath := testCfg.GetRepoPath("testrepo2")
		git.CreateWorktree(repoPath, crewPath, "tracy/work", "main")

		// Resolve crew path
		resolvedCrewPath, _ := filepath.EvalSymlinks(crewPath)

		// Change to crew directory
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		os.Chdir(resolvedCrewPath)

		rig, err := InferRig(&testCfg, "")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if rig != "testrepo2" {
			t.Errorf("Expected testrepo2, got %s", rig)
		}
	})

	t.Run("no inference possible", func(t *testing.T) {
		// Change to temp directory outside rigs/crew
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		os.Chdir(tmpDir)

		_, err := InferRig(cfg, "")
		if err == nil {
			t.Error("Expected error when inference is not possible")
		}
	})
}

func TestCrewWorkflow(t *testing.T) {
	// Skip if tmux is not available
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("tmux not available, skipping integration test")
	}

	cfg := setupTestConfig(t)

	// Create a test repo
	repoName := "testrig"
	createTestGitRepo(t, cfg.RigsBase, repoName)

	crewName := "testcrew"
	crewPath := cfg.GetCrewPath(repoName, crewName)
	sessionName := cfg.GetCrewSessionName(repoName, crewName)
	branchName := cfg.GetCrewBranchName(crewName)

	t.Run("add crew workspace", func(t *testing.T) {
		// Note: We can't fully test Add() because it tries to attach to tmux
		// Instead, we'll test the individual components

		// Validate name
		if err := ValidateCrewName(crewName); err != nil {
			t.Fatalf("Failed to validate crew name: %v", err)
		}

		// Create worktree
		repoPath := cfg.GetRepoPath(repoName)
		baseBranch, err := git.GetBaseBranch(repoPath, cfg.DefaultBranch)
		if err != nil {
			t.Fatalf("Failed to get base branch: %v", err)
		}

		os.MkdirAll(filepath.Dir(crewPath), 0755)

		err = git.CreateWorktree(repoPath, crewPath, branchName, baseBranch)
		if err != nil {
			t.Fatalf("Failed to create worktree: %v", err)
		}

		// Verify worktree exists
		if _, err := os.Stat(crewPath); os.IsNotExist(err) {
			t.Error("Expected crew path to exist")
		}

		if !git.WorktreeExists(repoPath, crewPath) {
			t.Error("Expected worktree to exist in git")
		}

		if !git.BranchExists(repoPath, branchName) {
			t.Error("Expected branch to exist")
		}
	})

	t.Run("verify branch is correct", func(t *testing.T) {
		currentBranch, err := git.GetCurrentBranch(crewPath)
		if err != nil {
			t.Fatalf("Failed to get current branch: %v", err)
		}

		if currentBranch != branchName {
			t.Errorf("Expected branch %s, got %s", branchName, currentBranch)
		}
	})

	t.Run("remove crew workspace", func(t *testing.T) {
		// We'll test the removal logic directly
		repoPath := cfg.GetRepoPath(repoName)

		// Remove worktree
		err := git.RemoveWorktree(repoPath, crewPath)
		if err != nil {
			t.Fatalf("Failed to remove worktree: %v", err)
		}

		// Prune
		git.PruneWorktrees(repoPath)

		// Delete branch
		err = git.DeleteBranch(repoPath, branchName)
		if err != nil {
			t.Fatalf("Failed to delete branch: %v", err)
		}

		// Verify cleanup
		if _, err := os.Stat(crewPath); !os.IsNotExist(err) {
			t.Error("Expected crew path to not exist")
		}

		if git.BranchExists(repoPath, branchName) {
			t.Error("Expected branch to be deleted")
		}
	})

	// Cleanup session if it exists
	defer func() {
		exec.Command("tmux", "kill-session", "-t", sessionName).Run()
	}()
}

func TestCrewPathStructure(t *testing.T) {
	cfg := setupTestConfig(t)

	// Test that crew paths follow the ~/crew/<repo>/<name> structure
	crewPath := cfg.GetCrewPath("myrepo", "tracy")
	expected := filepath.Join(cfg.CrewBase, "myrepo", "tracy")

	if crewPath != expected {
		t.Errorf("Expected crew path %s, got %s", expected, crewPath)
	}

	// Verify structure components
	dir := filepath.Dir(crewPath)
	expectedDir := filepath.Join(cfg.CrewBase, "myrepo")

	if dir != expectedDir {
		t.Errorf("Expected parent directory %s, got %s", expectedDir, dir)
	}

	base := filepath.Base(crewPath)
	if base != "tracy" {
		t.Errorf("Expected base name tracy, got %s", base)
	}
}
