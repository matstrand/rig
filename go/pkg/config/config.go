package config

import (
	"os"
	"path/filepath"
)

// Config holds all configuration for rig
type Config struct {
	RigsBase         string
	CrewBase         string
	UseCC            bool
	DefaultBranch    string
}

// Load reads configuration from environment variables
func Load() *Config {
	home := os.Getenv("HOME")

	rigsBase := os.Getenv("RIGS_BASE")
	if rigsBase == "" {
		rigsBase = filepath.Join(home, "git")
	}

	crewBase := os.Getenv("CREW_BASE")
	if crewBase == "" {
		crewBase = filepath.Join(home, "crew")
	}

	useCC := os.Getenv("RIG_USE_CC") == "true"

	defaultBranch := os.Getenv("RIG_DEFAULT_BRANCH")
	if defaultBranch == "" {
		defaultBranch = "main"
	}

	return &Config{
		RigsBase:      rigsBase,
		CrewBase:      crewBase,
		UseCC:         useCC,
		DefaultBranch: defaultBranch,
	}
}

// GetRepoPath returns the full path to a repo
func (c *Config) GetRepoPath(name string) string {
	return filepath.Join(c.RigsBase, name)
}

// GetCrewPath returns the path to a crew workspace
func (c *Config) GetCrewPath(rig, name string) string {
	return filepath.Join(c.CrewBase, rig, name)
}

// GetCrewSessionName returns the tmux session name for a crew member
func (c *Config) GetCrewSessionName(rig, name string) string {
	return rig + "@" + name
}

// GetCrewBranchName returns the branch name for a crew member
func (c *Config) GetCrewBranchName(name string) string {
	return name + "/work"
}
