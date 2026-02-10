package work

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Work represents a feature work item
type Work struct {
	Name        string // e.g., "build-frontend"
	Path        string // e.g., "work/build-frontend"
	FeatureName string // e.g., "feat/build-frontend"
}

// Progress represents the state of a work item
type Progress struct {
	Status     string
	AssignedTo string
	Tasks      []Task
	Notes      string
}

// Task represents a single task in the progress checklist
type Task struct {
	Done        bool
	Description string
}

// InferWorkFromBranch extracts work name from a feature branch name
// feat/build-frontend -> build-frontend
func InferWorkFromBranch(branchName string) string {
	if strings.HasPrefix(branchName, "feat/") {
		return strings.TrimPrefix(branchName, "feat/")
	}
	return ""
}

// GetWorkPath returns the full path to a work directory
func GetWorkPath(repoPath, workName string) string {
	return filepath.Join(repoPath, "work", workName)
}

// GetFormulaPath returns the path to a formula file
func GetFormulaPath(repoPath, formulaName string) string {
	return filepath.Join(repoPath, "work", "formula", formulaName+".md")
}

// Create creates a new work directory with scaffolded files
func Create(repoPath, workName string) error {
	workPath := GetWorkPath(repoPath, workName)
	formulaDir := filepath.Join(repoPath, "work", "formula")

	// Create work directory
	if err := os.MkdirAll(workPath, 0755); err != nil {
		return fmt.Errorf("failed to create work directory: %w", err)
	}

	// Create formula directory
	if err := os.MkdirAll(formulaDir, 0755); err != nil {
		return fmt.Errorf("failed to create formula directory: %w", err)
	}

	// Files to create (will skip if they exist)
	files := map[string]string{
		"spec.md":      getSpecTemplate(workName),
		"design.md":    getDesignTemplate(workName),
		"breakdown.md": getBreakdownTemplate(workName),
		"progress.md":  getProgressTemplate(workName),
	}

	createdFiles := []string{}
	skippedFiles := []string{}

	for filename, content := range files {
		filePath := filepath.Join(workPath, filename)
		if _, err := os.Stat(filePath); err == nil {
			skippedFiles = append(skippedFiles, filename)
			continue
		}

		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", filename, err)
		}
		createdFiles = append(createdFiles, filename)
	}

	// Install default formula if it doesn't exist
	if err := EnsureDefaultFormula(repoPath); err != nil {
		return fmt.Errorf("failed to install default formula: %w", err)
	}

	return nil
}

// EnsureDefaultFormula installs the default build formula if it doesn't exist
func EnsureDefaultFormula(repoPath string) error {
	formulaPath := GetFormulaPath(repoPath, "build")

	// Skip if already exists
	if _, err := os.Stat(formulaPath); err == nil {
		return nil
	}

	// Create formula directory if needed
	formulaDir := filepath.Dir(formulaPath)
	if err := os.MkdirAll(formulaDir, 0755); err != nil {
		return fmt.Errorf("failed to create formula directory: %w", err)
	}

	// Write default formula
	if err := os.WriteFile(formulaPath, []byte(getDefaultFormulaContent()), 0644); err != nil {
		return fmt.Errorf("failed to write formula: %w", err)
	}

	return nil
}

// ParseProgress reads and parses a progress.md file
func ParseProgress(path string) (*Progress, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open progress file: %w", err)
	}
	defer file.Close()

	progress := &Progress{
		Tasks: []Task{},
	}

	scanner := bufio.NewScanner(file)
	inChecklist := false
	inNotes := false
	notesLines := []string{}

	// Regex patterns
	statusRe := regexp.MustCompile(`(?i)^##\s*Status:\s*(.+)$`)
	assignedRe := regexp.MustCompile(`(?i)^##\s*Assigned to:\s*(.*)$`)
	checklistRe := regexp.MustCompile(`(?i)^##\s*Checklist\s*$`)
	notesRe := regexp.MustCompile(`(?i)^##\s*Notes\s*$`)
	taskRe := regexp.MustCompile(`^-\s*\[([ xX])\]\s*(.+)$`)

	for scanner.Scan() {
		line := scanner.Text()

		// Check for section headers
		if match := statusRe.FindStringSubmatch(line); match != nil {
			progress.Status = strings.TrimSpace(match[1])
			continue
		}

		if match := assignedRe.FindStringSubmatch(line); match != nil {
			progress.AssignedTo = strings.TrimSpace(match[1])
			continue
		}

		if checklistRe.MatchString(line) {
			inChecklist = true
			inNotes = false
			continue
		}

		if notesRe.MatchString(line) {
			inNotes = true
			inChecklist = false
			continue
		}

		// Parse tasks in checklist section
		if inChecklist {
			if match := taskRe.FindStringSubmatch(line); match != nil {
				done := strings.ToLower(match[1]) == "x"
				desc := strings.TrimSpace(match[2])
				progress.Tasks = append(progress.Tasks, Task{
					Done:        done,
					Description: desc,
				})
			}
		}

		// Collect notes
		if inNotes && strings.TrimSpace(line) != "" {
			notesLines = append(notesLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading progress file: %w", err)
	}

	progress.Notes = strings.Join(notesLines, "\n")
	return progress, nil
}

// GetCurrentTask returns the first unchecked task, or empty string if all done
func (p *Progress) GetCurrentTask() string {
	for _, task := range p.Tasks {
		if !task.Done {
			return task.Description
		}
	}
	return ""
}

// GenerateHook creates a hook.md file for a work item
func GenerateHook(repoPath, workName, formulaName string) error {
	workPath := GetWorkPath(repoPath, workName)
	hookPath := filepath.Join(workPath, "hook.md")
	formulaPath := GetFormulaPath(repoPath, formulaName)

	// Validate formula exists
	if _, err := os.Stat(formulaPath); os.IsNotExist(err) {
		return fmt.Errorf("formula not found: %s", formulaPath)
	}

	// Generate hook content
	content := fmt.Sprintf(`# Hook: %s

## Your Assignment

You are working on: **%s**

## Instructions

1. **Read the workflow formula**: Open and read work/formula/%s.md
   - This defines the phases you'll follow

2. **Read the spec**: Open and read work/%s/spec.md
   - This describes what you're building

3. **Follow the formula**: Execute each phase in order
   - Update work/%s/progress.md as you complete tasks
   - Commit your progress after each phase
   - Each commit should follow the pattern described in the formula

## Context Files

- Formula: work/formula/%s.md
- Spec: work/%s/spec.md
- Design: work/%s/design.md
- Breakdown: work/%s/breakdown.md
- Progress: work/%s/progress.md

## Important Notes

- Commit intermediate progress at each phase (don't wait until the end)
- Keep progress.md updated with your current status
- Follow the quality gates defined in the formula
- Ask questions if requirements are unclear

Ready? Start by reading the formula and spec files above.
`, workName, workName, formulaName, workName, workName, formulaName, workName, workName, workName, workName)

	// Write hook file
	if err := os.WriteFile(hookPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write hook file: %w", err)
	}

	return nil
}

// ListFormulas returns all available formula names
func ListFormulas(repoPath string) ([]string, error) {
	formulaDir := filepath.Join(repoPath, "work", "formula")

	// Check if formula directory exists
	if _, err := os.Stat(formulaDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	entries, err := os.ReadDir(formulaDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read formula directory: %w", err)
	}

	formulas := []string{}
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".md" {
			name := strings.TrimSuffix(entry.Name(), ".md")
			formulas = append(formulas, name)
		}
	}

	return formulas, nil
}

// Templates

func getSpecTemplate(workName string) string {
	return fmt.Sprintf(`# Spec: %s

## Overview

[Brief description of what this work aims to accomplish]

## Problem

[What problem are we solving?]

## Goals

[What are the specific objectives?]

## Non-Goals

[What are we explicitly not doing?]

## User Experience

[How will users interact with this feature?]

## Success Criteria

[How do we know when this is complete?]
`, strings.Title(strings.ReplaceAll(workName, "-", " ")))
}

func getDesignTemplate(workName string) string {
	return fmt.Sprintf(`# Design: %s

## Architecture

[High-level architectural approach]

## Components

[Key components and their responsibilities]

## Implementation Details

[Detailed technical approach]

## Risk Areas

[What could go wrong? How will we mitigate?]
`, strings.Title(strings.ReplaceAll(workName, "-", " ")))
}

func getBreakdownTemplate(workName string) string {
	return fmt.Sprintf(`# Implementation Breakdown: %s

## Tasks

1. [Task 1]
2. [Task 2]
3. [Task 3]

[Add as many tasks as needed with clear done criteria]
`, strings.Title(strings.ReplaceAll(workName, "-", " ")))
}

func getProgressTemplate(workName string) string {
	return fmt.Sprintf(`# Progress: %s

## Status: Not Started
## Assigned to:

## Checklist
- [ ] Spec review
- [ ] Initial design
- [ ] Design review
- [ ] Implementation breakdown
- [ ] Implementation
- [ ] Code review
- [ ] Testing
- [ ] Push feature branch
- [ ] Cleanup crew workspace

## Notes
`, strings.Title(strings.ReplaceAll(workName, "-", " ")))
}

func getDefaultFormulaContent() string {
	return `
# Feature Implementation Formula

Autonomous end-to-end feature implementation with built-in quality gates.
Takes a spec, designs the approach, implements the solution, validates
with tests, and commits to local git repo.

## Process

### Phase 1: Spec Review (Read-Only)
1. Read the spec thoroughly
2. Identify what exists vs what's new
3. List dependencies on other systems/modules
4. Flag critical gaps:
   - Missing acceptance criteria
   - Unclear requirements
   - Ambiguous edge cases

**Gate:** If critical gaps exist, create ` + "`CLARIFICATIONS.md`" + ` and STOP. Otherwise continue.

### Phase 2: Design
1. Survey existing codebase for patterns to follow
2. Identify files to create/modify
3. Design module structure and interfaces
4. Plan test strategy (unit, integration, e2e)
5. Update ` + "`design.md`" + ` with:
   - Files to change
   - New abstractions needed
   - Testing approach
   - Risk areas
6. **Commit progress:** ` + "`git commit -am \"docs: complete design phase\"`" + `

**Gate:** Review design. If major concerns, revise. Otherwise continue.

### Phase 3: Implementation Planning
1. Break design into tasks in ` + "`breakdown.md`" + `. Each task should:
   - Be completable in one session
   - Have clear done criteria
   - Be independently testable
   - Produce a commit
2. Update progress.md checklist with specific tasks
3. **Commit progress:** ` + "`git commit -am \"docs: create implementation breakdown\"`" + `

### Phase 4: Implementation
For each task:
1. Mark task as in progress in ` + "`progress.md`" + `
2. Work on task until done criteria met
3. Run relevant tests
4. **Commit with message:** ` + "`feat: [task description]`" + `
5. Mark task complete in ` + "`progress.md`" + `

**Gate:** After each task, verify tests pass. If fail, fix before next task.

### Phase 5: Review
1. Read all changed code
2. Check against spec acceptance criteria
3. Verify test coverage
4. Look for:
   - Performance issues
   - Security concerns
   - Error handling gaps
   - Documentation needs
5. Create review notes in ` + "`progress.md`" + `
6. **Commit progress:** ` + "`git commit -am \"docs: complete code review\"`" + `

**Gate:** If major issues, fix and re-review. Otherwise continue.

### Phase 6: Final Steps
1. Run full test suite
2. Update any necessary documentation
3. Final verification against spec
4. Update ` + "`progress.md`" + ` status to "Ready for Merge"
5. **Final commit:** ` + "`git commit -am \"docs: mark work ready for merge\"`" + `

## Important Notes

- **Commit intermediate progress at each phase** - This ensures work is always recoverable
- **Keep progress.md updated** - This is your state tracking mechanism
- **Each phase should leave work in a consistent state** - Anyone should be able to pick up from any phase
- **When complete, remind user to:**
  - Push feature branch: ` + "`git push -u origin feat/<feature-name>`" + `
  - Cleanup crew workspace: ` + "`rig crew remove <worker-name>`" + `
  - Create pull request if needed

## Outputs
- Updated ` + "`design.md`" + ` - Design document
- Updated ` + "`breakdown.md`" + ` - Implementation tasks
- Updated ` + "`progress.md`" + ` - Progress tracking with status
- Feature implementation with test coverage
- Git commits following conventional commits
`
}
