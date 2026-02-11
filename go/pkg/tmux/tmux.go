package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// NormalizeSessionName converts a session name to be tmux-compatible.
// Tmux automatically converts periods to underscores in session names,
// so we normalize them to prevent mismatches.
func NormalizeSessionName(name string) string {
	return strings.ReplaceAll(name, ".", "_")
}

// SessionExists checks if a tmux session exists
func SessionExists(name string) bool {
	name = NormalizeSessionName(name)
	cmd := exec.Command("tmux", "has-session", "-t", name)
	return cmd.Run() == nil
}

// ListSessions returns all active tmux sessions
func ListSessions() ([]string, error) {
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")
	output, err := cmd.Output()
	if err != nil {
		// No sessions exist
		return []string{}, nil
	}

	sessions := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(sessions) == 1 && sessions[0] == "" {
		return []string{}, nil
	}
	return sessions, nil
}

// KillSession kills a tmux session
func KillSession(name string) error {
	name = NormalizeSessionName(name)
	cmd := exec.Command("tmux", "kill-session", "-t", name)
	return cmd.Run()
}

// AttachSession attaches to a tmux session
func AttachSession(name string, useCC bool) error {
	name = NormalizeSessionName(name)
	inTmux := os.Getenv("TMUX") != ""

	if inTmux {
		// Already in tmux, switch client
		cmd := exec.Command("tmux", "switch-client", "-t", name)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// Not in tmux, attach
	args := []string{"attach-session", "-t", name}
	if useCC {
		args = append([]string{"-CC"}, args...)
	}
	cmd := exec.Command("tmux", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// AttachDefault attaches to the default tmux session (most recent or first)
func AttachDefault(useCC bool) error {
	inTmux := os.Getenv("TMUX") != ""

	if inTmux {
		return fmt.Errorf("already in a tmux session")
	}

	// Not in tmux, attach without specifying session
	args := []string{"attach-session"}
	if useCC {
		args = append([]string{"-CC"}, args...)
	}
	cmd := exec.Command("tmux", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// CreateRigSession creates a tmux session for a rig
func CreateRigSession(name, repoPath string, useCC bool) error {
	name = NormalizeSessionName(name)
	if useCC {
		return createRigSessionCC(name, repoPath)
	}
	return createRigSessionNative(name, repoPath)
}

func createRigSessionNative(name, repoPath string) error {
	// Create session with first window (Claude Code)
	cmd := exec.Command("tmux", "new-session", "-d", "-s", name, "-n", "Claude Code", "-c", repoPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Start Claude Code in first window
	sendKeys(name+":1", "cd "+repoPath)
	time.Sleep(100 * time.Millisecond)
	sendKeys(name+":1", "claude")

	// Create second window (Terminal)
	cmd = exec.Command("tmux", "new-window", "-t", name, "-n", "Terminal", "-c", repoPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create terminal window: %w", err)
	}

	// Add helpful header in terminal window
	sendKeys(name+":2", "cd "+repoPath)
	sendKeys(name+":2", fmt.Sprintf("echo '# %s terminal'", name))
	sendKeys(name+":2", "git status")

	// Select first window
	cmd = exec.Command("tmux", "select-window", "-t", name+":1")
	return cmd.Run()
}

func createRigSessionCC(name, repoPath string) error {
	// Create session with single window (add emoji to window name for iTerm2)
	windowName := "🏗️  " + name
	cmd := exec.Command("tmux", "new-session", "-d", "-s", name, "-n", windowName, "-c", repoPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Disable automatic renaming
	cmd = exec.Command("tmux", "set-window-option", "-t", name, "automatic-rename", "off")
	if err := cmd.Run(); err != nil {
		return err
	}

	// Split window vertically
	cmd = exec.Command("tmux", "split-window", "-h", "-t", name, "-c", repoPath)
	if err := cmd.Run(); err != nil {
		return err
	}

	// Set pane titles
	exec.Command("tmux", "select-pane", "-t", name+":.1", "-T", "Claude Code").Run()
	exec.Command("tmux", "select-pane", "-t", name+":.2", "-T", "Terminal").Run()

	// Resize panes (70/30 split)
	exec.Command("tmux", "resize-pane", "-t", name+":.1", "-x", "70%").Run()

	// Select Claude Code pane
	exec.Command("tmux", "select-pane", "-t", name+":.1").Run()

	// Start Claude Code
	sendKeys(name+":.1", "cd "+repoPath)
	time.Sleep(100 * time.Millisecond)
	sendKeys(name+":.1", "claude")

	// Terminal pane
	sendKeys(name+":.2", "cd "+repoPath)
	sendKeys(name+":.2", fmt.Sprintf("echo '# %s terminal'", name))
	sendKeys(name+":.2", "git status")

	return nil
}

// CreateCrewSession creates a tmux session for a crew member
func CreateCrewSession(sessionName, crewPath, rigName, memberName, branchName string, useCC bool) error {
	sessionName = NormalizeSessionName(sessionName)
	if useCC {
		return createCrewSessionCC(sessionName, crewPath, rigName, memberName, branchName)
	}
	return createCrewSessionNative(sessionName, crewPath, rigName, memberName, branchName)
}

func createCrewSessionNative(sessionName, crewPath, rigName, memberName, branchName string) error {
	// Create session with first window
	cmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName, "-n", "Claude Code", "-c", crewPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create crew session: %w", err)
	}

	// Start Claude Code
	sendKeys(sessionName+":1", "cd "+crewPath)
	time.Sleep(100 * time.Millisecond)
	sendKeys(sessionName+":1", "claude")

	// Create second window
	cmd = exec.Command("tmux", "new-window", "-t", sessionName, "-n", "Terminal", "-c", crewPath)
	if err := cmd.Run(); err != nil {
		return err
	}

	sendKeys(sessionName+":2", "cd "+crewPath)
	sendKeys(sessionName+":2", fmt.Sprintf("echo '# %s on %s (branch: %s)'", memberName, rigName, branchName))
	sendKeys(sessionName+":2", "git status")

	// Select first window
	cmd = exec.Command("tmux", "select-window", "-t", sessionName+":1")
	return cmd.Run()
}

func createCrewSessionCC(sessionName, crewPath, rigName, memberName, branchName string) error {
	// Determine emoji based on crew type
	emoji := "👤"
	if strings.HasPrefix(memberName, "polecat_") {
		emoji = "🐱"
	}
	windowName := emoji + " " + sessionName

	cmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName, "-n", windowName, "-c", crewPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create crew session: %w", err)
	}

	exec.Command("tmux", "set-window-option", "-t", sessionName, "automatic-rename", "off").Run()

	cmd = exec.Command("tmux", "split-window", "-h", "-t", sessionName, "-c", crewPath)
	if err := cmd.Run(); err != nil {
		return err
	}

	exec.Command("tmux", "select-pane", "-t", sessionName+":.1", "-T", "Claude Code").Run()
	exec.Command("tmux", "select-pane", "-t", sessionName+":.2", "-T", "Terminal").Run()
	exec.Command("tmux", "resize-pane", "-t", sessionName+":.1", "-x", "70%").Run()
	exec.Command("tmux", "select-pane", "-t", sessionName+":.1").Run()

	sendKeys(sessionName+":.1", "cd "+crewPath)
	time.Sleep(100 * time.Millisecond)
	sendKeys(sessionName+":.1", "claude")

	sendKeys(sessionName+":.2", "cd "+crewPath)
	sendKeys(sessionName+":.2", fmt.Sprintf("echo '# %s on %s (branch: %s)'", memberName, rigName, branchName))
	sendKeys(sessionName+":.2", "git status")

	return nil
}

func sendKeys(target, keys string) {
	exec.Command("tmux", "send-keys", "-t", target, keys, "C-m").Run()
}

// GetCurrentSession returns the current tmux session name, or empty string if not in tmux
func GetCurrentSession() string {
	if os.Getenv("TMUX") == "" {
		return ""
	}
	cmd := exec.Command("tmux", "display-message", "-p", "#S")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
