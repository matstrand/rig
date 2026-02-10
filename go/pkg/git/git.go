package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// BranchExists checks if a git branch exists
func BranchExists(repoPath, branchName string) bool {
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branchName)
	cmd.Dir = repoPath
	return cmd.Run() == nil
}

// GetBaseBranch returns the base branch to use, inferring from origin/HEAD if possible
func GetBaseBranch(repoPath, defaultBranch string) (string, error) {
	// First, try to infer from the remote's default branch
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err == nil {
		// Output will be something like "refs/remotes/origin/main"
		ref := strings.TrimSpace(string(output))
		branch := strings.TrimPrefix(ref, "refs/remotes/origin/")
		if branch != "" && BranchExists(repoPath, branch) {
			return branch, nil
		}
	}

	// Fallback: check if the configured default branch exists
	if BranchExists(repoPath, defaultBranch) {
		return defaultBranch, nil
	}

	// Last resort: try common default branch names
	for _, branch := range []string{"main", "master", "develop"} {
		if BranchExists(repoPath, branch) {
			return branch, nil
		}
	}

	return "", fmt.Errorf("could not find base branch (tried: origin/HEAD, %s, main, master, develop)", defaultBranch)
}

// WorktreeExists checks if a worktree exists at the given path
func WorktreeExists(repoPath, worktreePath string) bool {
	cmd := exec.Command("git", "worktree", "list")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), worktreePath)
}

// CreateWorktree creates a new git worktree
func CreateWorktree(repoPath, worktreePath, branchName, baseBranch string) error {
	cmd := exec.Command("git", "worktree", "add", worktreePath, "-b", branchName, baseBranch)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create worktree: %w\n%s", err, string(output))
	}
	return nil
}

// CreateWorktreeFromExisting creates a worktree from an existing branch
func CreateWorktreeFromExisting(repoPath, worktreePath, branchName string) error {
	cmd := exec.Command("git", "worktree", "add", worktreePath, branchName)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create worktree from existing branch: %w\n%s", err, string(output))
	}
	return nil
}

// RemoveWorktree removes a git worktree
func RemoveWorktree(repoPath, worktreePath string) error {
	cmd := exec.Command("git", "worktree", "remove", worktreePath, "--force")
	cmd.Dir = repoPath
	return cmd.Run()
}

// PruneWorktrees prunes stale worktree metadata
func PruneWorktrees(repoPath string) error {
	cmd := exec.Command("git", "worktree", "prune")
	cmd.Dir = repoPath
	return cmd.Run()
}

// DeleteBranch deletes a git branch
func DeleteBranch(repoPath, branchName string) error {
	cmd := exec.Command("git", "branch", "-D", branchName)
	cmd.Dir = repoPath
	return cmd.Run()
}

// GetCurrentBranch returns the current branch in a git directory
func GetCurrentBranch(path string) (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// CheckoutBranch checks out a branch
func CheckoutBranch(path, branchName string) error {
	cmd := exec.Command("git", "checkout", branchName)
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to checkout branch: %w\n%s", err, string(output))
	}
	return nil
}

// GetRepoRoot returns the root of the git repository
func GetRepoRoot(path string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// IsGitRepo checks if a directory is a git repository
func IsGitRepo(path string) bool {
	gitPath := filepath.Join(path, ".git")
	_, err := os.Stat(gitPath)
	return err == nil
}

// CreateFeatureBranch creates a new feature branch from a base branch
func CreateFeatureBranch(repoPath, branchName, baseBranch string) error {
	cmd := exec.Command("git", "checkout", "-b", branchName, baseBranch)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create feature branch: %w\n%s", err, string(output))
	}
	return nil
}

// Worktree represents a git worktree
type Worktree struct {
	Path   string
	Branch string
}

// ListWorktrees returns all worktrees for a repository
func ListWorktrees(repoPath string) ([]Worktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	worktrees := []Worktree{}
	lines := strings.Split(string(output), "\n")

	var currentPath string
	for _, line := range lines {
		if strings.HasPrefix(line, "worktree ") {
			currentPath = strings.TrimPrefix(line, "worktree ")
		} else if strings.HasPrefix(line, "branch ") {
			branch := strings.TrimPrefix(line, "branch ")
			branch = strings.TrimPrefix(branch, "refs/heads/")
			if currentPath != "" {
				worktrees = append(worktrees, Worktree{
					Path:   currentPath,
					Branch: branch,
				})
				currentPath = ""
			}
		}
	}

	return worktrees, nil
}

// GetWorktreeForBranch returns the worktree path for a given branch
func GetWorktreeForBranch(repoPath, branchName string) (string, error) {
	worktrees, err := ListWorktrees(repoPath)
	if err != nil {
		return "", err
	}

	for _, wt := range worktrees {
		if wt.Branch == branchName {
			return wt.Path, nil
		}
	}

	return "", fmt.Errorf("no worktree found for branch: %s", branchName)
}
