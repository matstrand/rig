package crew

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mstrand/rig/pkg/config"
	"github.com/mstrand/rig/pkg/git"
	"github.com/mstrand/rig/pkg/tmux"
)

// ValidateCrewName validates a crew member name
func ValidateCrewName(name string) error {
	if name == "" {
		return fmt.Errorf("crew name cannot be empty")
	}

	// Must not contain special characters
	if matched, _ := regexp.MatchString(`[/\\:@]`, name); matched {
		return fmt.Errorf("crew name cannot contain special characters (/, \\, :, @): %s", name)
	}

	// Must not start with . or -
	if matched, _ := regexp.MatchString(`^[.-]`, name); matched {
		return fmt.Errorf("crew name cannot start with . or -: %s", name)
	}

	// Length limit
	if len(name) > 50 {
		return fmt.Errorf("crew name too long (max 50 chars): %s", name)
	}

	return nil
}

// InferRig infers the rig name from current directory or tmux session
func InferRig(cfg *config.Config, explicitRig string) (string, error) {
	// If explicitly provided, use it
	if explicitRig != "" {
		return explicitRig, nil
	}

	// Check if pwd is under RIGS_BASE
	pwd, err := os.Getwd()
	if err == nil {
		pwdAbs, _ := filepath.Abs(pwd)
		if strings.HasPrefix(pwdAbs, cfg.RigsBase+string(filepath.Separator)) {
			root, err := git.GetRepoRoot(pwdAbs)
			if err == nil {
				return filepath.Base(root), nil
			}
		}

		// Check if pwd is under CREW_BASE
		if strings.HasPrefix(pwdAbs, cfg.CrewBase+string(filepath.Separator)) {
			root, err := git.GetRepoRoot(pwdAbs)
			if err == nil {
				// For crew workspaces, the structure is ~/crew/<rig>/<name>
				// We need to extract the rig name (parent of the worktree)
				relPath, err := filepath.Rel(cfg.CrewBase, root)
				if err == nil {
					// Split the relative path and get the first component (rig name)
					parts := strings.Split(relPath, string(filepath.Separator))
					if len(parts) > 0 {
						return parts[0], nil
					}
				}
			}
		}
	}

	// Check active tmux session
	sessionName := tmux.GetCurrentSession()
	if sessionName != "" {
		// If it's a crew session (format: <rig>@<name>), extract rig
		if strings.Contains(sessionName, "@") {
			parts := strings.Split(sessionName, "@")
			return parts[0], nil
		}

		// If it's a regular rig session, use it directly
		repoPath := cfg.GetRepoPath(sessionName)
		if git.IsGitRepo(repoPath) {
			return sessionName, nil
		}
	}

	return "", fmt.Errorf("could not infer rig. Use --rig=<repo> or run from within a repo in %s or %s", cfg.RigsBase, cfg.CrewBase)
}

// Add creates a new crew workspace
func Add(cfg *config.Config, name, rigName string) error {
	if err := ValidateCrewName(name); err != nil {
		return err
	}

	// Get repo path and validate it exists
	repoPath := cfg.GetRepoPath(rigName)
	if !git.IsGitRepo(repoPath) {
		return fmt.Errorf("repo not found: %s", repoPath)
	}

	crewPath := cfg.GetCrewPath(rigName, name)
	sessionName := cfg.GetCrewSessionName(rigName, name)
	branchName := cfg.GetCrewBranchName(name)

	// Get base branch
	baseBranch, err := git.GetBaseBranch(repoPath, cfg.DefaultBranch)
	if err != nil {
		return err
	}

	// Check if worktree already exists (idempotency)
	if _, err := os.Stat(crewPath); err == nil {
		if tmux.SessionExists(sessionName) {
			fmt.Printf("Crew workspace already exists and session is running\n")
			fmt.Printf("Attaching to existing session: %s\n", sessionName)
			return tmux.AttachSession(sessionName, cfg.UseCC)
		}

		fmt.Printf("Crew workspace exists but session is not running\n")
		fmt.Printf("Recreating session...\n")

		if err := tmux.CreateCrewSession(sessionName, crewPath, rigName, name, branchName, cfg.UseCC); err != nil {
			return fmt.Errorf("failed to recreate session: %w", err)
		}

		fmt.Printf("✓ Session recreated: %s\n", sessionName)
		return tmux.AttachSession(sessionName, cfg.UseCC)
	}

	// Create crew directory
	if err := os.MkdirAll(filepath.Dir(crewPath), 0755); err != nil {
		return fmt.Errorf("failed to create crew directory: %w", err)
	}

	fmt.Printf("Creating crew workspace for %s on %s\n", name, rigName)
	fmt.Printf("  Repo: %s\n", repoPath)
	fmt.Printf("  Workspace: %s\n", crewPath)
	fmt.Printf("  Branch: %s (from %s)\n", branchName, baseBranch)

	// Check if branch already exists
	useExistingBranch := false
	if git.BranchExists(repoPath, branchName) {
		fmt.Printf("Branch %s already exists\n", branchName)
		fmt.Print("Use existing branch? [Y/n] ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) == "n" {
			return fmt.Errorf("cancelled. Delete the branch first or use a different crew name")
		}
		useExistingBranch = true
	}

	// Create worktree
	if useExistingBranch {
		if err := git.CreateWorktreeFromExisting(repoPath, crewPath, branchName); err != nil {
			return err
		}
	} else {
		if err := git.CreateWorktree(repoPath, crewPath, branchName, baseBranch); err != nil {
			// Cleanup on failure
			cleanupWorktree(repoPath, crewPath, branchName)
			return err
		}
	}

	fmt.Printf("✓ Crew workspace created: %s\n", crewPath)

	// Create tmux session
	if err := tmux.CreateCrewSession(sessionName, crewPath, rigName, name, branchName, cfg.UseCC); err != nil {
		fmt.Printf("Session creation failed, cleaning up worktree...\n")
		cleanupWorktree(repoPath, crewPath, branchName)
		return fmt.Errorf("failed to create session: %w", err)
	}

	fmt.Printf("✓ Session created: %s\n", sessionName)

	// Attach to session
	return tmux.AttachSession(sessionName, cfg.UseCC)
}

// Start attaches to an existing crew workspace
func Start(cfg *config.Config, name, rigName string) error {
	if err := ValidateCrewName(name); err != nil {
		return err
	}

	crewPath := cfg.GetCrewPath(rigName, name)
	sessionName := cfg.GetCrewSessionName(rigName, name)
	branchName := cfg.GetCrewBranchName(name)

	// Check if worktree exists
	if _, err := os.Stat(crewPath); os.IsNotExist(err) {
		return fmt.Errorf("crew workspace not found: %s\nUse 'rig crew add %s --rig=%s' first", crewPath, name, rigName)
	}

	// Verify we're on the expected branch
	currentBranch, err := git.GetCurrentBranch(crewPath)
	if err == nil && currentBranch != "" && currentBranch != branchName {
		fmt.Printf("Workspace is on branch '%s', expected '%s'\n", currentBranch, branchName)
		fmt.Printf("Switch to %s? [Y/n] ", branchName)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "n" {
			if err := git.CheckoutBranch(crewPath, branchName); err != nil {
				return fmt.Errorf("failed to switch to branch %s: %w", branchName, err)
			}
			fmt.Printf("✓ Switched to branch %s\n", branchName)
		}
	}

	// Check if session exists
	if !tmux.SessionExists(sessionName) {
		fmt.Printf("Session doesn't exist, recreating...\n")
		if err := tmux.CreateCrewSession(sessionName, crewPath, rigName, name, branchName, cfg.UseCC); err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}
		fmt.Printf("✓ Session created: %s\n", sessionName)
	}

	// Attach to session
	return tmux.AttachSession(sessionName, cfg.UseCC)
}

// Remove removes a crew workspace
func Remove(cfg *config.Config, name, rigName string) error {
	if err := ValidateCrewName(name); err != nil {
		return err
	}

	repoPath := cfg.GetRepoPath(rigName)
	if !git.IsGitRepo(repoPath) {
		return fmt.Errorf("repo not found: %s", repoPath)
	}

	crewPath := cfg.GetCrewPath(rigName, name)
	sessionName := cfg.GetCrewSessionName(rigName, name)
	branchName := cfg.GetCrewBranchName(name)

	// Check if worktree directory exists
	worktreeDirExists := false
	if _, err := os.Stat(crewPath); err == nil {
		worktreeDirExists = true
	}

	// Check if git thinks worktree exists
	worktreeInGit := git.WorktreeExists(repoPath, crewPath)

	// Handle detached state
	if worktreeInGit && !worktreeDirExists {
		fmt.Printf("Worktree is in detached state (git knows about it but directory is gone)\n")
		fmt.Printf("Cleaning up git worktree metadata...\n")
		git.RemoveWorktree(repoPath, crewPath)
		git.PruneWorktrees(repoPath)
		worktreeInGit = false
	}

	// Neither directory nor git reference exists
	if !worktreeDirExists && !worktreeInGit {
		// Maybe just the session exists?
		if tmux.SessionExists(sessionName) {
			fmt.Printf("Only session exists (no worktree), killing it...\n")
			tmux.KillSession(sessionName)
			fmt.Printf("✓ Session killed: %s\n", sessionName)
			return nil
		}
		return fmt.Errorf("crew workspace not found: %s", crewPath)
	}

	// Warn if user is currently in this session
	if tmux.SessionExists(sessionName) && tmux.GetCurrentSession() == sessionName {
		fmt.Printf("You are currently in session '%s' - removing it will disconnect you\n", sessionName)
	}

	// Ask about branch deletion BEFORE killing session
	deleteBranch := false
	if git.BranchExists(repoPath, branchName) {
		fmt.Printf("Delete branch %s? [Y/n] ", branchName)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "n" {
			deleteBranch = true
		}
	}

	// Kill tmux session if running
	if tmux.SessionExists(sessionName) {
		fmt.Printf("Killing session: %s\n", sessionName)
		tmux.KillSession(sessionName)
	}

	// Remove git worktree
	if worktreeDirExists {
		fmt.Printf("Removing worktree: %s\n", crewPath)
		git.RemoveWorktree(repoPath, crewPath)
	}

	// Prune stale worktree metadata
	git.PruneWorktrees(repoPath)

	// Delete branch if user confirmed
	if deleteBranch {
		git.DeleteBranch(repoPath, branchName)
		fmt.Printf("✓ Branch deleted: %s\n", branchName)
	}

	// Remove empty repo directory
	repoDir := filepath.Dir(crewPath)
	if entries, err := os.ReadDir(repoDir); err == nil && len(entries) == 0 {
		os.Remove(repoDir)
		fmt.Printf("Removed empty directory: %s\n", repoDir)
	}

	fmt.Printf("✓ Crew workspace removed: %s on %s\n", name, rigName)
	return nil
}

func cleanupWorktree(repoPath, crewPath, branchName string) {
	git.RemoveWorktree(repoPath, crewPath)
	git.PruneWorktrees(repoPath)
	git.DeleteBranch(repoPath, branchName)
}
