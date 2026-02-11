package tmux

import "testing"

func TestNormalizeSessionName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with.period", "with_period"},
		{"multiple.dots.here", "multiple_dots_here"},
		{"no-change-needed", "no-change-needed"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeSessionName(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeSessionName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
