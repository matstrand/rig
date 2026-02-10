package polecat

import (
	"strings"
	"testing"
)

func TestIsPolecat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid polecat", "polecat_emma", true},
		{"Valid polecat 2", "polecat_olivia", true},
		{"Not a polecat", "tracy", false},
		{"Prefix only", "polecat_", true},
		{"Empty string", "", false},
		{"Different prefix", "crew_emma", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPolecat(tt.input)
			if result != tt.expected {
				t.Errorf("IsPolecat(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateName(t *testing.T) {
	tests := []struct {
		name string
		used []string
	}{
		{"No names used", []string{}},
		{"Some names used", []string{"polecat_emma", "polecat_olivia"}},
		{"Non-polecat names ignored", []string{"tracy", "sam"}},
		{"Mixed names", []string{"polecat_emma", "tracy", "polecat_ava"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generated := GenerateName(tt.used)

			// Check format
			if !strings.HasPrefix(generated, "polecat_") {
				t.Errorf("GenerateName() = %q, should start with 'polecat_'", generated)
			}

			// Extract name part
			parts := strings.Split(generated, "_")
			if len(parts) != 2 {
				t.Errorf("GenerateName() = %q, should have format 'polecat_<name>'", generated)
				return
			}

			baseName := parts[1]

			// Check it's from the pool
			found := false
			for _, name := range names {
				if name == baseName {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("GenerateName() = %q, name %q not in pool", generated, baseName)
			}

			// If names are available, should not pick a used one
			usedMap := make(map[string]bool)
			for _, used := range tt.used {
				if IsPolecat(used) {
					usedParts := strings.Split(used, "_")
					if len(usedParts) == 2 {
						usedMap[usedParts[1]] = true
					}
				}
			}

			// Only check if there are available names
			if len(usedMap) < len(names) && usedMap[baseName] {
				t.Errorf("GenerateName() = %q, should not pick used name when others available", generated)
			}
		})
	}
}

func TestGenerateNameExhaustion(t *testing.T) {
	// Use all names
	used := []string{}
	for _, name := range names {
		used = append(used, "polecat_"+name)
	}

	// Should still generate a name (reuse)
	generated := GenerateName(used)
	if !IsPolecat(generated) {
		t.Errorf("GenerateName() with all names used should still return valid polecat name, got %q", generated)
	}
}
