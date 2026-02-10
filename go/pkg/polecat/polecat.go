package polecat

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mstrand/rig/pkg/config"
)

var names = []string{
	"emma", "olivia", "ava", "sophia", "mia", "charlotte",
	"amelia", "harper", "evelyn", "abigail", "ella", "scarlett",
	"grace", "chloe", "lily", "zoe", "maya", "lucy",
	"isabella", "aria", "aurora", "violet", "nova", "hazel",
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// GenerateName generates a random polecat name not in the used list
func GenerateName(used []string) string {
	usedMap := make(map[string]bool)
	for _, name := range used {
		if IsPolecat(name) {
			// Extract the base name from polecat_<name>
			parts := strings.Split(name, "_")
			if len(parts) == 2 {
				usedMap[parts[1]] = true
			}
		}
	}

	// Find available names
	available := []string{}
	for _, name := range names {
		if !usedMap[name] {
			available = append(available, name)
		}
	}

	if len(available) == 0 {
		// All names used, pick random one anyway
		return fmt.Sprintf("polecat_%s", names[rand.Intn(len(names))])
	}

	return fmt.Sprintf("polecat_%s", available[rand.Intn(len(available))])
}

// IsPolecat checks if a name follows polecat naming convention
func IsPolecat(name string) bool {
	return strings.HasPrefix(name, "polecat_")
}

// List returns all polecat names from all rigs
func List(cfg *config.Config) ([]string, error) {
	polecats := []string{}

	// Check if crew base exists
	if _, err := os.Stat(cfg.CrewBase); os.IsNotExist(err) {
		return polecats, nil
	}

	// Scan all rigs
	rigDirs, err := os.ReadDir(cfg.CrewBase)
	if err != nil {
		return nil, fmt.Errorf("failed to read crew directory: %w", err)
	}

	for _, rigDir := range rigDirs {
		if !rigDir.IsDir() {
			continue
		}

		rigPath := filepath.Join(cfg.CrewBase, rigDir.Name())
		crewDirs, err := os.ReadDir(rigPath)
		if err != nil {
			continue
		}

		for _, crewDir := range crewDirs {
			if !crewDir.IsDir() {
				continue
			}

			name := crewDir.Name()
			if IsPolecat(name) {
				polecats = append(polecats, name)
			}
		}
	}

	return polecats, nil
}
