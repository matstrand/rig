package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mstrand/rig/pkg/config"
	"github.com/mstrand/rig/pkg/crew"
	"github.com/mstrand/rig/pkg/git"
	"github.com/mstrand/rig/pkg/polecat"
	"github.com/mstrand/rig/pkg/tmux"
	"github.com/mstrand/rig/pkg/work"
	"github.com/spf13/cobra"
)

var cfg *config.Config

// condensePath replaces the home directory with ~ for shorter display
func condensePath(path string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if strings.HasPrefix(path, homeDir) {
		return "~" + strings.TrimPrefix(path, homeDir)
	}
	return path
}

func main() {
	cfg = config.Load()

	rootCmd := &cobra.Command{
		Use:   "rig",
		Short: "Manage tmux-based development environments",
		Long: `Rig - Manage tmux-based development environments

Examples:
    rig up myapp            Start rig for ~/git/myapp
    rig up                  Start rig (infers from current directory)
    rig status              Show all running rigs and crew
    rig down myapp          Shut down the myapp rig
    rig down                Shut down current rig (infers from context)`,
	}

	// Rig commands
	rootCmd.AddCommand(upCmd())
	rootCmd.AddCommand(downCmd())
	rootCmd.AddCommand(statusCmd())
	rootCmd.AddCommand(listCmd())
	rootCmd.AddCommand(switchCmd())
	rootCmd.AddCommand(atCmd())
	rootCmd.AddCommand(killallCmd())

	// Crew commands
	rootCmd.AddCommand(crewCmd())

	// Work commands
	rootCmd.AddCommand(workCmd())
	rootCmd.AddCommand(hookCmd())
	rootCmd.AddCommand(slingCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func upCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "up [name]",
		Short: "Bring up a rig (creates or switches)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var name string
			var err error

			if len(args) == 0 {
				// Infer rig from current context
				name, err = crew.InferRig(cfg, "")
				if err != nil {
					return err
				}
				fmt.Printf("Inferred rig: %s\n", name)
			} else {
				name = args[0]
			}

			repoPath := cfg.GetRepoPath(name)

			if !git.IsGitRepo(repoPath) {
				return fmt.Errorf("repo not found: %s", repoPath)
			}

			sessionName := name

			if tmux.SessionExists(sessionName) {
				fmt.Printf("Switching to existing rig: %s\n", name)
				return tmux.AttachSession(sessionName, cfg.UseCC)
			}

			fmt.Printf("Creating new rig: %s\n", name)
			fmt.Printf("Repo: %s\n", repoPath)

			if err := tmux.CreateRigSession(sessionName, repoPath, cfg.UseCC); err != nil {
				return fmt.Errorf("failed to create rig session: %w", err)
			}

			fmt.Printf("✓ Rig created: %s\n", name)
			return tmux.AttachSession(sessionName, cfg.UseCC)
		},
	}
}

func downCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "down [name]",
		Short: "Shut down a rig",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var name string
			var err error

			if len(args) == 0 {
				// Infer rig from current context
				name, err = crew.InferRig(cfg, "")
				if err != nil {
					return err
				}
				fmt.Printf("Inferred rig: %s\n", name)
			} else {
				name = args[0]
			}

			if !tmux.SessionExists(name) {
				return fmt.Errorf("rig not found: %s", name)
			}

			if err := tmux.KillSession(name); err != nil {
				return err
			}

			fmt.Printf("✓ Rig shut down: %s\n", name)
			return nil
		},
	}
}

func statusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status",
		Aliases: []string{"ls"},
		Short:   "Show all active rigs and crew",
		RunE: func(cmd *cobra.Command, args []string) error {
			sessions, err := tmux.ListSessions()
			if err != nil {
				return err
			}

			if len(sessions) == 0 {
				fmt.Println("No active rigs or crew")
				fmt.Println()
				fmt.Println("Start a rig with: rig up <name>")
				fmt.Println("Start crew with: rig crew add <name>")
				return nil
			}

			currentSession := tmux.GetCurrentSession()

			var rigSessions []string
			var crewSessions []string

			for _, session := range sessions {
				if strings.Contains(session, "@") {
					// Crew session
					parts := strings.Split(session, "@")
					rigPart, namePart := parts[0], parts[1]
					crewPath := cfg.GetCrewPath(rigPart, namePart)
					if _, err := os.Stat(crewPath); err == nil {
						crewSessions = append(crewSessions, session)
					}
				} else {
					// Rig session
					repoPath := cfg.GetRepoPath(session)
					if git.IsGitRepo(repoPath) {
						rigSessions = append(rigSessions, session)
					}
				}
			}

			// Display rig sessions
			fmt.Println("🏗️  Active Rigs")
			fmt.Println()

			if len(rigSessions) == 0 {
				fmt.Println("  No active rigs")
			} else {
				for _, session := range rigSessions {
					activeMarker := " "
					if session == currentSession {
						activeMarker = "✓"
					}
					repoPath := cfg.GetRepoPath(session)
					branch, err := git.GetCurrentBranch(repoPath)
					if err != nil {
						branch = "unknown"
					}

					// Condense path with ~
					displayPath := condensePath(repoPath)

					fmt.Printf("  %s %s\n", activeMarker, session)
					fmt.Printf("      %-50s 🌿 %s\n", displayPath, branch)
					fmt.Println()
				}
			}

			// Display crew sessions
			fmt.Println("👥 Crew")
			fmt.Println()

			if len(crewSessions) == 0 {
				fmt.Println("  No active crew")
			} else {
				for _, session := range crewSessions {
					activeMarker := " "
					if session == currentSession {
						activeMarker = "✓"
					}
					parts := strings.Split(session, "@")
					rigPart, namePart := parts[0], parts[1]
					crewPath := cfg.GetCrewPath(rigPart, namePart)

					emoji := "👤"
					if polecat.IsPolecat(namePart) {
						emoji = "🐱"
					}

					branch, err := git.GetCurrentBranch(crewPath)
					if err != nil {
						branch = "unknown"
					}

					// Condense path with ~
					displayPath := condensePath(crewPath)

					fmt.Printf("  %s %s %s\n", activeMarker, emoji, session)
					fmt.Printf("      %-50s 🌿 %s\n", displayPath, branch)
					fmt.Println()
				}
			}

			if len(rigSessions) == 0 && len(crewSessions) == 0 {
				fmt.Println()
				fmt.Println("Start a rig with: rig up <name>")
				fmt.Println("Start crew with: rig crew add <name>")
			}

			return nil
		},
	}
	return cmd
}

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available repos",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("🏗️  Available Repos")
			fmt.Println()

			entries, err := os.ReadDir(cfg.RigsBase)
			if err != nil {
				return fmt.Errorf("base directory does not exist: %s", cfg.RigsBase)
			}

			count := 0
			for _, entry := range entries {
				if entry.IsDir() {
					path := filepath.Join(cfg.RigsBase, entry.Name())
					if git.IsGitRepo(path) {
						status := ""
						if tmux.SessionExists(entry.Name()) {
							status = " [running]"
						}
						fmt.Printf("  %s%s\n", entry.Name(), status)
						count++
					}
				}
			}

			if count == 0 {
				fmt.Println("  No git repos found")
			}

			fmt.Println()
			fmt.Printf("Total: %d repos\n", count)
			return nil
		},
	}
}

func switchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "switch <name>",
		Short: "Switch to a rig or crew session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionName := args[0]

			if !tmux.SessionExists(sessionName) {
				return fmt.Errorf("session not found: %s", sessionName)
			}

			return tmux.AttachSession(sessionName, cfg.UseCC)
		},
	}
}

func atCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "at [name]",
		Short: "Attach to a tmux session (default session if no name provided)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				// No name provided, attach to default tmux session
				return tmux.AttachDefault(cfg.UseCC)
			}

			// Name provided, attach to specific session
			sessionName := args[0]
			if !tmux.SessionExists(sessionName) {
				return fmt.Errorf("session not found: %s", sessionName)
			}

			return tmux.AttachSession(sessionName, cfg.UseCC)
		},
	}
}

func killallCmd() *cobra.Command {
	var killCrew bool
	var crewOnly bool

	cmd := &cobra.Command{
		Use:   "killall",
		Short: "Shut down all rigs (add --crew to include crew)",
		RunE: func(cmd *cobra.Command, args []string) error {
			sessions, err := tmux.ListSessions()
			if err != nil {
				return err
			}

			if len(sessions) == 0 {
				fmt.Println("No active rigs or crew")
				return nil
			}

			killedCount := 0

			for _, session := range sessions {
				isCrew := strings.Contains(session, "@")
				isRig := false

				if !isCrew {
					repoPath := cfg.GetRepoPath(session)
					if git.IsGitRepo(repoPath) {
						isRig = true
					}
				} else {
					parts := strings.Split(session, "@")
					rigPart, namePart := parts[0], parts[1]
					crewPath := cfg.GetCrewPath(rigPart, namePart)
					if _, err := os.Stat(crewPath); err != nil {
						isCrew = false
					}
				}

				shouldKill := false
				if crewOnly {
					shouldKill = isCrew
				} else if killCrew {
					shouldKill = isRig || isCrew
				} else {
					shouldKill = isRig
				}

				if shouldKill {
					tmux.KillSession(session)
					fmt.Printf("  Killed: %s\n", session)
					killedCount++
				}
			}

			if killedCount == 0 {
				fmt.Println("No matching sessions to kill")
			} else {
				fmt.Printf("Killed %d session(s)\n", killedCount)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&killCrew, "crew", false, "Kill both rigs and crew")
	cmd.Flags().BoolVar(&crewOnly, "crew-only", false, "Kill only crew sessions")

	return cmd
}

func crewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "crew",
		Short: "Manage crew workspaces",
	}

	cmd.AddCommand(crewAddCmd())
	cmd.AddCommand(crewStartCmd())
	cmd.AddCommand(crewRemoveCmd())
	cmd.AddCommand(crewListCmd())
	cmd.AddCommand(crewStatusCmd())
	cmd.AddCommand(crewPruneCmd())

	return cmd
}

func crewAddCmd() *cobra.Command {
	var rigName string

	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Create crew workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Infer rig if not provided
			if rigName == "" {
				var err error
				rigName, err = crew.InferRig(cfg, rigName)
				if err != nil {
					return err
				}
			}

			return crew.Add(cfg, name, rigName)
		},
	}

	cmd.Flags().StringVar(&rigName, "rig", "", "Explicit rig name")

	return cmd
}

func crewStartCmd() *cobra.Command {
	var rigName string

	cmd := &cobra.Command{
		Use:   "start <name>",
		Short: "Attach to crew workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Infer rig if not provided
			if rigName == "" {
				var err error
				rigName, err = crew.InferRig(cfg, rigName)
				if err != nil {
					return err
				}
			}

			return crew.Start(cfg, name, rigName)
		},
	}

	cmd.Flags().StringVar(&rigName, "rig", "", "Explicit rig name")

	return cmd
}

func crewRemoveCmd() *cobra.Command {
	var rigName string

	cmd := &cobra.Command{
		Use:     "remove <name>",
		Aliases: []string{"rm"},
		Short:   "Remove crew workspace",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Infer rig if not provided
			if rigName == "" {
				var err error
				rigName, err = crew.InferRig(cfg, rigName)
				if err != nil {
					return err
				}
			}

			return crew.Remove(cfg, name, rigName)
		},
	}

	cmd.Flags().StringVar(&rigName, "rig", "", "Explicit rig name")

	return cmd
}

func crewListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "ls [name]",
		Aliases: []string{"list"},
		Short:   "List crew workspaces",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filterName := ""
			if len(args) == 1 {
				filterName = args[0]
			}

			if _, err := os.Stat(cfg.CrewBase); os.IsNotExist(err) {
				fmt.Printf("No crew workspaces (directory doesn't exist: %s)\n", cfg.CrewBase)
				return nil
			}

			// Build map of rigs to their crew members
			type CrewMember struct {
				Name   string
				Branch string
				Status string
			}
			rigCrew := make(map[string][]CrewMember)

			repoDirs, err := os.ReadDir(cfg.CrewBase)
			if err != nil {
				return err
			}

			for _, repoDir := range repoDirs {
				if !repoDir.IsDir() {
					continue
				}

				rigName := repoDir.Name()
				repoPath := filepath.Join(cfg.CrewBase, rigName)

				workspaces, err := os.ReadDir(repoPath)
				if err != nil {
					continue
				}

				for _, workspace := range workspaces {
					if !workspace.IsDir() {
						continue
					}

					crewName := workspace.Name()

					// Filter by name if provided
					if filterName != "" && crewName != filterName {
						continue
					}

					crewPath := filepath.Join(repoPath, crewName)
					sessionName := cfg.GetCrewSessionName(rigName, crewName)

					// Get branch
					branch, err := git.GetCurrentBranch(crewPath)
					if err != nil {
						branch = "unknown"
					}

					// Get status
					status := "stopped"
					if tmux.SessionExists(sessionName) {
						status = "running"
					}

					rigCrew[rigName] = append(rigCrew[rigName], CrewMember{
						Name:   crewName,
						Branch: branch,
						Status: status,
					})
				}
			}

			if len(rigCrew) == 0 {
				if filterName != "" {
					fmt.Printf("No workspaces found for: %s\n", filterName)
				} else {
					fmt.Println("No crew workspaces found")
				}
				fmt.Println()
				fmt.Println("Create one with: rig crew add <name>")
				return nil
			}

			// Display by rig
			for rigName, crew := range rigCrew {
				fmt.Printf("🏗️  %s\n", rigName)

				for _, member := range crew {
					emoji := "👤"
					if polecat.IsPolecat(member.Name) {
						emoji = "🐱"
					}

					fmt.Printf("  %s %-18s %-26s [%s]\n", emoji, member.Name, member.Branch, member.Status)
				}
				fmt.Println()
			}

			return nil
		},
	}
}

func crewStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show active crew sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("👥 Active Crew Sessions")
			fmt.Println()

			sessions, err := tmux.ListSessions()
			if err != nil || len(sessions) == 0 {
				fmt.Println("  No active crew sessions")
				return nil
			}

			var crewSessions []string
			for _, session := range sessions {
				if strings.Contains(session, "@") {
					parts := strings.Split(session, "@")
					rigPart, namePart := parts[0], parts[1]
					crewPath := cfg.GetCrewPath(rigPart, namePart)

					if _, err := os.Stat(crewPath); err == nil {
						crewSessions = append(crewSessions, session)
					}
				}
			}

			if len(crewSessions) == 0 {
				fmt.Println("  No active crew sessions")
				return nil
			}

			for _, session := range crewSessions {
				parts := strings.Split(session, "@")
				rigPart, namePart := parts[0], parts[1]
				crewPath := cfg.GetCrewPath(rigPart, namePart)

				emoji := "👤"
				if polecat.IsPolecat(namePart) {
					emoji = "🐱"
				}

				branch, err := git.GetCurrentBranch(crewPath)
				if err != nil {
					branch = "unknown"
				}

				fmt.Printf("  %s %s\n", emoji, session)
				fmt.Printf("      %s\n", crewPath)
				fmt.Printf("      %s\n", branch)
				fmt.Println()
			}

			return nil
		},
	}
}

func crewPruneCmd() *cobra.Command {
	var polecatsOnly bool

	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Remove crew workspaces (with --polecats flag, removes only polecats)",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Find polecats across all rigs
			type PolecatInfo struct {
				Name    string
				RigName string
				Path    string
			}

			polecats := []PolecatInfo{}

			if _, err := os.Stat(cfg.CrewBase); os.IsNotExist(err) {
				fmt.Println("No crew workspaces found")
				return nil
			}

			rigDirs, err := os.ReadDir(cfg.CrewBase)
			if err != nil {
				return fmt.Errorf("failed to read crew directory: %w", err)
			}

			for _, rigDir := range rigDirs {
				if !rigDir.IsDir() {
					continue
				}

				rigName := rigDir.Name()
				rigPath := filepath.Join(cfg.CrewBase, rigName)

				crewDirs, err := os.ReadDir(rigPath)
				if err != nil {
					continue
				}

				for _, crewDir := range crewDirs {
					if !crewDir.IsDir() {
						continue
					}

					crewName := crewDir.Name()
					if polecat.IsPolecat(crewName) {
						polecats = append(polecats, PolecatInfo{
							Name:    crewName,
							RigName: rigName,
							Path:    filepath.Join(rigPath, crewName),
						})
					}
				}
			}

			if len(polecats) == 0 {
				fmt.Println("No polecats found")
				return nil
			}

			// Display found polecats
			fmt.Printf("Found %d polecat(s):\n", len(polecats))
			for _, p := range polecats {
				fmt.Printf("  - 🐱 %s (rig: %s)\n", p.Name, p.RigName)
			}
			fmt.Println()

			// Confirm removal
			fmt.Print("Remove these workspaces and worktrees? (y/N) ")
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" {
				fmt.Println("Cancelled")
				return nil
			}

			// Remove each polecat
			for _, p := range polecats {
				fmt.Printf("Removing 🐱 %s...\n", p.Name)

				// Get repo path
				repoPath := cfg.GetRepoPath(p.RigName)
				sessionName := cfg.GetCrewSessionName(p.RigName, p.Name)

				// Kill session if running
				if tmux.SessionExists(sessionName) {
					tmux.KillSession(sessionName)
					fmt.Printf("  ✓ Killed session: %s\n", sessionName)
				}

				// Remove worktree
				if _, err := os.Stat(p.Path); err == nil {
					git.RemoveWorktree(repoPath, p.Path)
					fmt.Printf("  ✓ Removed worktree: %s\n", p.Path)
				}

				// Prune stale worktree metadata
				git.PruneWorktrees(repoPath)

				// Remove empty rig directory if needed
				rigDir := filepath.Dir(p.Path)
				if entries, err := os.ReadDir(rigDir); err == nil && len(entries) == 0 {
					os.Remove(rigDir)
				}
			}

			fmt.Printf("\n✓ Removed %d polecat(s)\n", len(polecats))
			return nil
		},
	}

	cmd.Flags().BoolVar(&polecatsOnly, "polecats", false, "Remove only polecats (default behavior)")

	return cmd
}

func workCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "work",
		Short: "Manage feature work",
	}

	cmd.AddCommand(workCreateCmd())
	cmd.AddCommand(workStatusCmd())

	return cmd
}

func workCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new work directory with feature branch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workName := args[0]

			// Get current directory and find repo root
			pwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}

			repoPath, err := git.GetRepoRoot(pwd)
			if err != nil {
				return fmt.Errorf("not in a git repository: %w", err)
			}

			// Check if work directory already exists
			workPath := work.GetWorkPath(repoPath, workName)
			workExists := false
			if _, err := os.Stat(workPath); err == nil {
				workExists = true
				fmt.Printf("⚠️  Warning: work/%s/ already exists\n", workName)
			}

			// Check if feature branch already exists
			featureBranch := "feat/" + workName
			branchExists := git.BranchExists(repoPath, featureBranch)
			if branchExists {
				fmt.Printf("⚠️  Warning: Branch %s already exists\n", featureBranch)
			}

			// Create work directory and files
			if err := work.Create(repoPath, workName); err != nil {
				return fmt.Errorf("failed to create work directory: %w", err)
			}

			if workExists {
				fmt.Println("✓ Skipped existing files, created missing ones")
			} else {
				fmt.Printf("✓ Created work directory: work/%s/\n", workName)
			}

			// Get base branch
			baseBranch, err := git.GetBaseBranch(repoPath, cfg.DefaultBranch)
			if err != nil {
				return err
			}

			// Create feature branch if it doesn't exist
			if !branchExists {
				if err := git.CreateFeatureBranch(repoPath, featureBranch, baseBranch); err != nil {
					return fmt.Errorf("failed to create feature branch: %w", err)
				}
				fmt.Printf("✓ Created feature branch: %s\n", featureBranch)
			} else {
				// Checkout existing branch
				if err := git.CheckoutBranch(repoPath, featureBranch); err != nil {
					return fmt.Errorf("failed to checkout branch: %w", err)
				}
				fmt.Printf("✓ Using existing branch: %s\n", featureBranch)
			}

			// Check if formula was installed
			formulaPath := work.GetFormulaPath(repoPath, "build")
			if _, err := os.Stat(formulaPath); err == nil {
				fmt.Println("✓ Ensured formula directory exists: work/formula/")
			}

			// Create initial commit if work directory was newly created
			if !workExists {
				// Stage work directory
				addCmd := exec.Command("git", "add", "work/"+workName+"/", "work/formula/")
				addCmd.Dir = repoPath
				if err := addCmd.Run(); err != nil {
					fmt.Printf("⚠️  Warning: failed to stage files: %v\n", err)
				} else {
					// Create commit
					commitMsg := fmt.Sprintf("Initialize work: %s", workName)
					commitCmd := exec.Command("git", "commit", "-m", commitMsg)
					commitCmd.Dir = repoPath
					if err := commitCmd.Run(); err != nil {
						fmt.Printf("⚠️  Warning: failed to create initial commit: %v\n", err)
					} else {
						fmt.Printf("✓ Initial commit: \"%s\"\n", commitMsg)
					}
				}
			}

			fmt.Println()
			fmt.Println("Next steps:")
			fmt.Printf("  1. Edit work/%s/spec.md\n", workName)
			fmt.Printf("  2. When ready: rig sling work/%s\n", workName)
			fmt.Printf("\nYou are now on branch: %s\n", featureBranch)

			return nil
		},
	}
}

func workStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show all active work across all rigs",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("💼 Active Work")
			fmt.Println()

			// Check if crew base exists
			if _, err := os.Stat(cfg.CrewBase); os.IsNotExist(err) {
				fmt.Println("No crew workspaces found")
				return nil
			}

			// Scan all rigs
			rigDirs, err := os.ReadDir(cfg.CrewBase)
			if err != nil {
				return fmt.Errorf("failed to read crew directory: %w", err)
			}

			type WorkItem struct {
				WorkName    string
				Status      string
				AssignedTo  string
				Branch      string
				CurrentTask string
			}

			rigWork := make(map[string][]WorkItem)

			for _, rigDir := range rigDirs {
				if !rigDir.IsDir() {
					continue
				}

				rigName := rigDir.Name()
				rigPath := filepath.Join(cfg.CrewBase, rigName)

				// Scan crew members in this rig
				crewDirs, err := os.ReadDir(rigPath)
				if err != nil {
					continue
				}

				for _, crewDir := range crewDirs {
					if !crewDir.IsDir() {
						continue
					}

					crewName := crewDir.Name()
					crewPath := filepath.Join(rigPath, crewName)

					// Get current branch
					branch, err := git.GetCurrentBranch(crewPath)
					if err != nil {
						continue
					}

					// Check if it's a feature branch
					workName := work.InferWorkFromBranch(branch)
					if workName == "" {
						continue
					}

					// Try to read progress.md
					progressPath := filepath.Join(crewPath, "work", workName, "progress.md")
					progress, err := work.ParseProgress(progressPath)
					if err != nil {
						// If progress.md doesn't exist or can't be parsed, show basic info
						rigWork[rigName] = append(rigWork[rigName], WorkItem{
							WorkName:    workName,
							Status:      "Unknown",
							AssignedTo:  crewName,
							Branch:      branch,
							CurrentTask: "",
						})
						continue
					}

					// Add work item with full details
					rigWork[rigName] = append(rigWork[rigName], WorkItem{
						WorkName:    workName,
						Status:      progress.Status,
						AssignedTo:  crewName,
						Branch:      branch,
						CurrentTask: progress.GetCurrentTask(),
					})
				}
			}

			if len(rigWork) == 0 {
				fmt.Println("No active work found")
				fmt.Println()
				fmt.Println("Create work with: rig work create <name>")
				fmt.Println("Assign work with: rig sling work/<name>")
				return nil
			}

			// Display work grouped by rig
			for rigName, workItems := range rigWork {
				fmt.Printf("🏗️  %s\n", rigName)

				for _, item := range workItems {
					statusDisplay := item.Status
					if statusDisplay == "" {
						statusDisplay = "Unknown"
					}

					emoji := "👤"
					if polecat.IsPolecat(item.AssignedTo) {
						emoji = "🐱"
					}

					fmt.Printf("  %-20s [%-14s] %s %-18s %s\n",
						item.WorkName,
						statusDisplay,
						emoji,
						item.AssignedTo,
						item.Branch)

					if item.CurrentTask != "" {
						fmt.Printf("    → %s\n", item.CurrentTask)
					}
				}
				fmt.Println()
			}

			return nil
		},
	}
}

func hookCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hook",
		Short: "Display the hook file for current work",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get current directory and find repo root
			pwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}

			repoPath, err := git.GetRepoRoot(pwd)
			if err != nil {
				return fmt.Errorf("not in a git repository: %w", err)
			}

			// Get current branch
			branch, err := git.GetCurrentBranch(repoPath)
			if err != nil {
				return fmt.Errorf("failed to get current branch: %w", err)
			}

			// Infer work name from branch
			workName := work.InferWorkFromBranch(branch)
			if workName == "" {
				return fmt.Errorf("not on a feature branch (expected feat/<name>), current branch: %s", branch)
			}

			// Find hook file
			hookPath := filepath.Join(work.GetWorkPath(repoPath, workName), "hook.md")
			if _, err := os.Stat(hookPath); os.IsNotExist(err) {
				return fmt.Errorf("no hook found for work: %s\nRun 'rig sling work/%s' to create one", workName, workName)
			}

			// Read and display hook
			content, err := os.ReadFile(hookPath)
			if err != nil {
				return fmt.Errorf("failed to read hook: %w", err)
			}

			fmt.Printf("🪝 Hook: %s\n\n", workName)
			fmt.Print(string(content))

			return nil
		},
	}
}

func slingCmd() *cobra.Command {
	var toName string
	var formulaName string
	var self bool

	cmd := &cobra.Command{
		Use:   "sling <work-path>",
		Short: "Assign work to a crew member or polecat",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workPath := args[0]

			// Parse work path (e.g., "work/build-frontend")
			parts := strings.Split(workPath, "/")
			if len(parts) != 2 || parts[0] != "work" {
				return fmt.Errorf("work path must be in format 'work/<name>', got: %s", workPath)
			}
			workName := parts[1]

			// Get current directory and find repo root
			pwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}

			repoPath, err := git.GetRepoRoot(pwd)
			if err != nil {
				return fmt.Errorf("not in a git repository: %w", err)
			}

			// Infer rig name
			rigName := filepath.Base(repoPath)

			// Verify work directory exists
			fullWorkPath := work.GetWorkPath(repoPath, workName)
			if _, err := os.Stat(fullWorkPath); os.IsNotExist(err) {
				return fmt.Errorf("work directory not found: %s\nRun 'rig work create %s' first", workPath, workName)
			}

			// Feature branch name
			featureBranch := "feat/" + workName

			// Verify feature branch exists
			if !git.BranchExists(repoPath, featureBranch) {
				return fmt.Errorf("feature branch not found: %s\nRun 'rig work create %s' first", featureBranch, workName)
			}

			// Get current branch
			currentBranch, err := git.GetCurrentBranch(repoPath)
			if err != nil {
				return fmt.Errorf("failed to get current branch: %w", err)
			}

			// If we're not on the feature branch, switch to it first
			if currentBranch != featureBranch {
				fmt.Printf("Switching to %s...\n", featureBranch)
				if err := git.CheckoutBranch(repoPath, featureBranch); err != nil {
					return fmt.Errorf("failed to checkout feature branch: %w", err)
				}
			}

			// Default formula to "build"
			if formulaName == "" {
				formulaName = "build"
			}

			// Validate formula exists
			formulas, err := work.ListFormulas(repoPath)
			if err != nil {
				return fmt.Errorf("failed to list formulas: %w", err)
			}

			formulaExists := false
			for _, f := range formulas {
				if f == formulaName {
					formulaExists = true
					break
				}
			}

			if !formulaExists {
				if len(formulas) == 0 {
					return fmt.Errorf("formula not found: %s\nNo formulas available", formulaName)
				}
				return fmt.Errorf("formula not found: %s\nAvailable formulas: %s", formulaName, strings.Join(formulas, ", "))
			}

			// Generate hook (while on feature branch)
			if err := work.GenerateHook(repoPath, workName, formulaName); err != nil {
				return fmt.Errorf("failed to generate hook: %w", err)
			}

			fmt.Printf("✓ Created hook: work/%s/hook.md\n", workName)

			// Check for uncommitted changes in work directory (including hook.md)
			statusCmd := exec.Command("git", "status", "--porcelain", "work/"+workName+"/")
			statusCmd.Dir = repoPath
			statusOutput, err := statusCmd.Output()
			if err != nil {
				return fmt.Errorf("failed to check git status: %w", err)
			}

			hasUncommittedChanges := len(strings.TrimSpace(string(statusOutput))) > 0

			if hasUncommittedChanges {
				fmt.Println("⚠️  Uncommitted changes in work directory:")
				fmt.Println(string(statusOutput))
				fmt.Print("Commit these changes before slinging? (Y/n) ")
				var response string
				fmt.Scanln(&response)

				if strings.ToLower(response) == "n" {
					return fmt.Errorf("cancelled - please commit your changes before slinging")
				}

				// Commit the changes (including hook.md)
				addCmd := exec.Command("git", "add", "work/"+workName+"/")
				addCmd.Dir = repoPath
				if err := addCmd.Run(); err != nil {
					return fmt.Errorf("failed to stage changes: %w", err)
				}

				commitMsg := fmt.Sprintf("Update work files for %s", workName)
				commitCmd := exec.Command("git", "commit", "-m", commitMsg)
				commitCmd.Dir = repoPath
				if err := commitCmd.Run(); err != nil {
					return fmt.Errorf("failed to commit changes: %w", err)
				}

				fmt.Printf("✓ Committed changes: \"%s\"\n", commitMsg)
			}

			// Now switch to base branch (making feature branch available for worktree)
			baseBranch, err := git.GetBaseBranch(repoPath, cfg.DefaultBranch)
			if err != nil {
				return fmt.Errorf("failed to get base branch: %w", err)
			}

			fmt.Printf("Switching to %s...\n", baseBranch)
			if err := git.CheckoutBranch(repoPath, baseBranch); err != nil {
				return fmt.Errorf("failed to checkout base branch: %w", err)
			}

			// Handle --self flag
			if self {
				fmt.Println("✓ Hook ready in current workspace")
				fmt.Println()
				fmt.Println("To start working, run this command in your Claude Code session:")
				fmt.Println("  rig hook")
				fmt.Println()
				fmt.Println("This will display your work instructions. Follow them to begin.")
				return nil
			}

			// Handle --to flag (existing crew member)
			if toName != "" {
				crewPath := cfg.GetCrewPath(rigName, toName)

				// Check if crew workspace exists
				if _, err := os.Stat(crewPath); os.IsNotExist(err) {
					return fmt.Errorf("crew workspace not found: %s\nRun 'rig crew add %s --rig=%s' first", crewPath, toName, rigName)
				}

				// Check if on correct branch
				currentBranch, err := git.GetCurrentBranch(crewPath)
				if err == nil && currentBranch != featureBranch {
					fmt.Printf("⚠️  Warning: %s is on branch '%s', expected '%s'\n", toName, currentBranch, featureBranch)
					fmt.Print("Checkout feature branch? [Y/n] ")
					var response string
					fmt.Scanln(&response)
					if strings.ToLower(response) != "n" {
						if err := git.CheckoutBranch(crewPath, featureBranch); err != nil {
							return fmt.Errorf("failed to checkout branch: %w", err)
						}
						fmt.Printf("✓ Checked out branch: %s\n", featureBranch)
					}
				}

				fmt.Printf("✓ Workspace ready: %s\n", crewPath)
				fmt.Printf("✓ Branch: %s\n", featureBranch)
				fmt.Println()
				fmt.Printf("To start working, paste this command into %s's Claude Code session:\n", toName)
				fmt.Println("  rig hook")
				fmt.Println()
				fmt.Println("This will display the work instructions. Ask them to follow the instructions.")
				return nil
			}

			// Create polecat (default behavior)
			// Get list of existing crew members for name generation
			existingNames := []string{}
			crewBaseForRig := filepath.Join(cfg.CrewBase, rigName)
			if entries, err := os.ReadDir(crewBaseForRig); err == nil {
				for _, entry := range entries {
					if entry.IsDir() {
						existingNames = append(existingNames, entry.Name())
					}
				}
			}

			// Check if work is already assigned
			worktrees, err := git.ListWorktrees(repoPath)
			if err == nil {
				for _, wt := range worktrees {
					if wt.Branch == featureBranch {
						// Extract crew name from path
						relPath, err := filepath.Rel(cfg.CrewBase, wt.Path)
						if err == nil {
							pathParts := strings.Split(relPath, string(filepath.Separator))
							if len(pathParts) >= 2 {
								assignedName := pathParts[1]
								displayName := assignedName
								if polecat.IsPolecat(assignedName) {
									displayName = "🐱 " + assignedName
								}
								fmt.Printf("⚠️  Warning: work/%s is already assigned to %s\n", workName, displayName)
								fmt.Printf("   Workspace: %s\n", wt.Path)
								fmt.Println()
								fmt.Print("Reassign to new polecat? (y/N) ")
								var response string
								fmt.Scanln(&response)
								if strings.ToLower(response) != "y" {
									return fmt.Errorf("cancelled")
								}
							}
						}
					}
				}
			}

			// Generate polecat name
			polecatName := polecat.GenerateName(existingNames)

			fmt.Printf("✓ Created polecat: 🐱 %s\n", polecatName)

			// Create crew workspace for polecat
			crewPath := cfg.GetCrewPath(rigName, polecatName)
			sessionName := cfg.GetCrewSessionName(rigName, polecatName)

			// Create worktree
			if err := os.MkdirAll(filepath.Dir(crewPath), 0755); err != nil {
				return fmt.Errorf("failed to create crew directory: %w", err)
			}

			// Check if worktree for this branch already exists
			existingWorktree, _ := git.GetWorktreeForBranch(repoPath, featureBranch)
			if existingWorktree != "" && existingWorktree != crewPath {
				// Check if the existing worktree is the main repo
				existingResolved, _ := filepath.EvalSymlinks(existingWorktree)
				repoResolved, _ := filepath.EvalSymlinks(repoPath)

				if existingResolved == repoResolved {
					// The feature branch is still checked out in the main repo
					// This shouldn't happen since we already switched earlier, but handle it just in case
					baseBranch, err := git.GetBaseBranch(repoPath, cfg.DefaultBranch)
					if err != nil {
						return fmt.Errorf("failed to get base branch: %w", err)
					}

					fmt.Printf("Switching main repo to %s...\n", baseBranch)
					if err := git.CheckoutBranch(repoPath, baseBranch); err != nil {
						return fmt.Errorf("failed to checkout base branch in main repo: %w", err)
					}
				} else {
					// It's a crew worktree, remove it
					git.RemoveWorktree(repoPath, existingWorktree)
					git.PruneWorktrees(repoPath)
				}
			}

			// Create worktree from existing feature branch
			if err := git.CreateWorktreeFromExisting(repoPath, crewPath, featureBranch); err != nil {
				return fmt.Errorf("failed to create worktree: %w", err)
			}

			fmt.Printf("✓ Workspace: %s\n", crewPath)
			fmt.Printf("✓ Session: %s\n", sessionName)
			fmt.Printf("✓ Branch: %s\n", featureBranch)

			// Create tmux session
			if err := tmux.CreateCrewSession(sessionName, crewPath, rigName, polecatName, featureBranch, cfg.UseCC); err != nil {
				// Cleanup on failure
				git.RemoveWorktree(repoPath, crewPath)
				git.PruneWorktrees(repoPath)
				return fmt.Errorf("failed to create session: %w", err)
			}

			// Send initial command to Claude Code
			time := 2000 // milliseconds - wait for Claude Code to start
			sleepCmd := exec.Command("sleep", fmt.Sprintf("%.1f", float64(time)/1000.0))
			sleepCmd.Run()

			// Send the hook command to the first pane (Claude Code)
			target := sessionName + ":.1"

			// First send a clear instruction message
			instructionMsg := "# YOUR WORK ASSIGNMENT: Run the command 'rig hook' to see your instructions"
			sendCmd := exec.Command("tmux", "send-keys", "-t", target, instructionMsg)
			sendCmd.Run()

			// Send Enter to show the message
			sleepCmd = exec.Command("sleep", "0.1")
			sleepCmd.Run()
			sendCmd = exec.Command("tmux", "send-keys", "-t", target, "C-m")
			sendCmd.Run()

			// Small delay
			sleepCmd = exec.Command("sleep", "0.1")
			sleepCmd.Run()

			// Now send the actual rig hook command
			sendCmd = exec.Command("tmux", "send-keys", "-t", target, "rig hook")
			sendCmd.Run()

			// Small delay
			sleepCmd = exec.Command("sleep", "0.1")
			sleepCmd.Run()

			// Then send Enter to execute it
			sendCmd = exec.Command("tmux", "send-keys", "-t", target, "C-m")
			sendCmd.Run()

			fmt.Println()
			fmt.Println("Session started. Sent 'rig hook' command to Claude Code.")

			return nil
		},
	}

	cmd.Flags().StringVar(&toName, "to", "", "Assign to existing crew member")
	cmd.Flags().StringVar(&formulaName, "formula", "", "Formula to use (default: build)")
	cmd.Flags().BoolVar(&self, "self", false, "Work on it yourself in current session")

	return cmd
}
