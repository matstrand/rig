package work

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInferWorkFromBranch(t *testing.T) {
	tests := []struct {
		name       string
		branchName string
		expected   string
	}{
		{"Feature branch", "feat/build-frontend", "build-frontend"},
		{"Feature branch 2", "feat/add-auth", "add-auth"},
		{"Non-feature branch", "main", ""},
		{"Other prefix", "bugfix/something", ""},
		{"Empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InferWorkFromBranch(tt.branchName)
			if result != tt.expected {
				t.Errorf("InferWorkFromBranch(%q) = %q, want %q", tt.branchName, result, tt.expected)
			}
		})
	}
}

func TestGetWorkPath(t *testing.T) {
	result := GetWorkPath("/home/user/repo", "my-feature")
	expected := "/home/user/repo/work/my-feature"
	if result != expected {
		t.Errorf("GetWorkPath() = %q, want %q", result, expected)
	}
}

func TestGetFormulaPath(t *testing.T) {
	result := GetFormulaPath("/home/user/repo", "build")
	expected := "/home/user/repo/work/formula/build.md"
	if result != expected {
		t.Errorf("GetFormulaPath() = %q, want %q", result, expected)
	}
}

func TestParseProgress(t *testing.T) {
	// Create a temp file with progress content
	tmpDir := t.TempDir()
	progressFile := filepath.Join(tmpDir, "progress.md")

	content := `# Progress: Test Feature

## Status: In Progress
## Assigned to: polecat_emma

## Checklist
- [x] Spec review
- [x] Initial design
- [ ] Design review
- [ ] Implementation

## Notes
Some important notes here
About the work
`

	if err := os.WriteFile(progressFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Parse the file
	progress, err := ParseProgress(progressFile)
	if err != nil {
		t.Fatalf("ParseProgress() error = %v", err)
	}

	// Verify status
	if progress.Status != "In Progress" {
		t.Errorf("Status = %q, want %q", progress.Status, "In Progress")
	}

	// Verify assigned to
	if progress.AssignedTo != "polecat_emma" {
		t.Errorf("AssignedTo = %q, want %q", progress.AssignedTo, "polecat_emma")
	}

	// Verify tasks
	if len(progress.Tasks) != 4 {
		t.Errorf("len(Tasks) = %d, want %d", len(progress.Tasks), 4)
	}

	// Check first task
	if !progress.Tasks[0].Done {
		t.Errorf("Tasks[0].Done = false, want true")
	}
	if progress.Tasks[0].Description != "Spec review" {
		t.Errorf("Tasks[0].Description = %q, want %q", progress.Tasks[0].Description, "Spec review")
	}

	// Check third task (not done)
	if progress.Tasks[2].Done {
		t.Errorf("Tasks[2].Done = true, want false")
	}
	if progress.Tasks[2].Description != "Design review" {
		t.Errorf("Tasks[2].Description = %q, want %q", progress.Tasks[2].Description, "Design review")
	}

	// Verify notes
	if !contains(progress.Notes, "important notes") {
		t.Errorf("Notes missing expected content")
	}
}

func TestGetCurrentTask(t *testing.T) {
	tests := []struct {
		name     string
		progress *Progress
		expected string
	}{
		{
			name: "First task not done",
			progress: &Progress{
				Tasks: []Task{
					{Done: false, Description: "Task 1"},
					{Done: false, Description: "Task 2"},
				},
			},
			expected: "Task 1",
		},
		{
			name: "First done, second not",
			progress: &Progress{
				Tasks: []Task{
					{Done: true, Description: "Task 1"},
					{Done: false, Description: "Task 2"},
				},
			},
			expected: "Task 2",
		},
		{
			name: "All done",
			progress: &Progress{
				Tasks: []Task{
					{Done: true, Description: "Task 1"},
					{Done: true, Description: "Task 2"},
				},
			},
			expected: "",
		},
		{
			name:     "No tasks",
			progress: &Progress{Tasks: []Task{}},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.progress.GetCurrentTask()
			if result != tt.expected {
				t.Errorf("GetCurrentTask() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestGenerateHook(t *testing.T) {
	tmpDir := t.TempDir()
	workName := "test-feature"
	formulaName := "build"

	// Create work directory
	workPath := GetWorkPath(tmpDir, workName)
	if err := os.MkdirAll(workPath, 0755); err != nil {
		t.Fatalf("Failed to create work directory: %v", err)
	}

	// Create formula file
	formulaPath := GetFormulaPath(tmpDir, formulaName)
	if err := os.MkdirAll(filepath.Dir(formulaPath), 0755); err != nil {
		t.Fatalf("Failed to create formula directory: %v", err)
	}
	if err := os.WriteFile(formulaPath, []byte("# Test Formula"), 0644); err != nil {
		t.Fatalf("Failed to create formula file: %v", err)
	}

	// Generate hook
	err := GenerateHook(tmpDir, workName, formulaName)
	if err != nil {
		t.Fatalf("GenerateHook() error = %v", err)
	}

	// Verify hook file exists
	hookPath := filepath.Join(workPath, "hook.md")
	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		t.Error("Hook file was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("Failed to read hook file: %v", err)
	}

	contentStr := string(content)
	if !contains(contentStr, "test-feature") {
		t.Error("Hook content missing work name")
	}
	if !contains(contentStr, "work/formula/build.md") {
		t.Error("Hook content missing formula path")
	}
	if !contains(contentStr, "work/test-feature/spec.md") {
		t.Error("Hook content missing spec path")
	}
}

func TestGenerateHookMissingFormula(t *testing.T) {
	tmpDir := t.TempDir()
	workName := "test-feature"
	formulaName := "nonexistent"

	// Create work directory
	workPath := GetWorkPath(tmpDir, workName)
	if err := os.MkdirAll(workPath, 0755); err != nil {
		t.Fatalf("Failed to create work directory: %v", err)
	}

	// Try to generate hook with missing formula
	err := GenerateHook(tmpDir, workName, formulaName)
	if err == nil {
		t.Error("Expected error for missing formula, got nil")
	}
}

func TestListFormulas(t *testing.T) {
	tmpDir := t.TempDir()

	// Initially no formulas
	formulas, err := ListFormulas(tmpDir)
	if err != nil {
		t.Fatalf("ListFormulas() error = %v", err)
	}
	if len(formulas) != 0 {
		t.Errorf("Expected 0 formulas, got %d", len(formulas))
	}

	// Create some formulas
	formulaDir := filepath.Join(tmpDir, "work", "formula")
	if err := os.MkdirAll(formulaDir, 0755); err != nil {
		t.Fatalf("Failed to create formula directory: %v", err)
	}

	formulaNames := []string{"build", "hotfix", "custom"}
	for _, name := range formulaNames {
		path := filepath.Join(formulaDir, name+".md")
		if err := os.WriteFile(path, []byte("# "+name), 0644); err != nil {
			t.Fatalf("Failed to create formula: %v", err)
		}
	}

	// List formulas
	formulas, err = ListFormulas(tmpDir)
	if err != nil {
		t.Fatalf("ListFormulas() error = %v", err)
	}

	if len(formulas) != 3 {
		t.Errorf("Expected 3 formulas, got %d", len(formulas))
	}

	// Verify all formula names are present
	formulaMap := make(map[string]bool)
	for _, f := range formulas {
		formulaMap[f] = true
	}

	for _, expected := range formulaNames {
		if !formulaMap[expected] {
			t.Errorf("Expected formula %s not found", expected)
		}
	}
}
