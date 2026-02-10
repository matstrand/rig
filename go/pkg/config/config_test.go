package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	// Save original env vars
	origRigsBase := os.Getenv("RIGS_BASE")
	origCrewBase := os.Getenv("CREW_BASE")
	origUseCC := os.Getenv("RIG_USE_CC")
	origDefaultBranch := os.Getenv("RIG_DEFAULT_BRANCH")

	defer func() {
		os.Setenv("RIGS_BASE", origRigsBase)
		os.Setenv("CREW_BASE", origCrewBase)
		os.Setenv("RIG_USE_CC", origUseCC)
		os.Setenv("RIG_DEFAULT_BRANCH", origDefaultBranch)
	}()

	t.Run("default values", func(t *testing.T) {
		os.Unsetenv("RIGS_BASE")
		os.Unsetenv("CREW_BASE")
		os.Unsetenv("RIG_USE_CC")
		os.Unsetenv("RIG_DEFAULT_BRANCH")

		cfg := Load()

		home := os.Getenv("HOME")
		expectedRigsBase := filepath.Join(home, "git")
		expectedCrewBase := filepath.Join(home, "crew")

		if cfg.RigsBase != expectedRigsBase {
			t.Errorf("Expected RigsBase=%s, got %s", expectedRigsBase, cfg.RigsBase)
		}
		if cfg.CrewBase != expectedCrewBase {
			t.Errorf("Expected CrewBase=%s, got %s", expectedCrewBase, cfg.CrewBase)
		}
		if cfg.UseCC != false {
			t.Errorf("Expected UseCC=false, got %v", cfg.UseCC)
		}
		if cfg.DefaultBranch != "main" {
			t.Errorf("Expected DefaultBranch=main, got %s", cfg.DefaultBranch)
		}
	})

	t.Run("custom values", func(t *testing.T) {
		os.Setenv("RIGS_BASE", "/custom/rigs")
		os.Setenv("CREW_BASE", "/custom/crew")
		os.Setenv("RIG_USE_CC", "true")
		os.Setenv("RIG_DEFAULT_BRANCH", "develop")

		cfg := Load()

		if cfg.RigsBase != "/custom/rigs" {
			t.Errorf("Expected RigsBase=/custom/rigs, got %s", cfg.RigsBase)
		}
		if cfg.CrewBase != "/custom/crew" {
			t.Errorf("Expected CrewBase=/custom/crew, got %s", cfg.CrewBase)
		}
		if cfg.UseCC != true {
			t.Errorf("Expected UseCC=true, got %v", cfg.UseCC)
		}
		if cfg.DefaultBranch != "develop" {
			t.Errorf("Expected DefaultBranch=develop, got %s", cfg.DefaultBranch)
		}
	})
}

func TestGetRepoPath(t *testing.T) {
	cfg := &Config{RigsBase: "/test/git"}

	path := cfg.GetRepoPath("myrepo")
	expected := "/test/git/myrepo"

	if path != expected {
		t.Errorf("Expected %s, got %s", expected, path)
	}
}

func TestGetCrewPath(t *testing.T) {
	cfg := &Config{CrewBase: "/test/crew"}

	path := cfg.GetCrewPath("myrepo", "tracy")
	expected := "/test/crew/myrepo/tracy"

	if path != expected {
		t.Errorf("Expected %s, got %s", expected, path)
	}
}

func TestGetCrewSessionName(t *testing.T) {
	cfg := &Config{}

	sessionName := cfg.GetCrewSessionName("notes", "tracy")
	expected := "notes@tracy"

	if sessionName != expected {
		t.Errorf("Expected %s, got %s", expected, sessionName)
	}
}

func TestGetCrewBranchName(t *testing.T) {
	cfg := &Config{}

	branchName := cfg.GetCrewBranchName("tracy")
	expected := "tracy/work"

	if branchName != expected {
		t.Errorf("Expected %s, got %s", expected, branchName)
	}
}
