
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

**Gate:** If critical gaps exist, create `CLARIFICATIONS.md` and STOP. Otherwise continue.

### Phase 2: Design
1. Survey existing codebase for patterns to follow
2. Identify files to create/modify
3. Design module structure and interfaces
4. Plan test strategy (unit, integration, e2e)
5. Update `design.md` with:
   - Files to change
   - New abstractions needed
   - Testing approach
   - Risk areas
6. **Commit progress:** `git commit -am "docs: complete design phase"`

**Gate:** Review design. If major concerns, revise. Otherwise continue.

### Phase 3: Implementation Planning
1. Break design into tasks in `breakdown.md`. Each task should:
   - Be completable in one session
   - Have clear done criteria
   - Be independently testable
   - Produce a commit
2. Update progress.md checklist with specific tasks
3. **Commit progress:** `git commit -am "docs: create implementation breakdown"`

### Phase 4: Implementation
For each task:
1. Mark task as in progress in `progress.md`
2. Work on task until done criteria met
3. Run relevant tests
4. **Commit with message:** `feat: [task description]`
5. Mark task complete in `progress.md`

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
5. Create review notes in `progress.md`
6. **Commit progress:** `git commit -am "docs: complete code review"`

**Gate:** If major issues, fix and re-review. Otherwise continue.

### Phase 6: Final Steps
1. Run full test suite
2. Update any necessary documentation
3. Final verification against spec
4. Update `progress.md` status to "Ready for Merge"
5. **Final commit:** `git commit -am "docs: mark work ready for merge"`

## Important Notes

- **Commit intermediate progress at each phase** - This ensures work is always recoverable
- **Keep progress.md updated** - This is your state tracking mechanism
- **Each phase should leave work in a consistent state** - Anyone should be able to pick up from any phase
- **When complete, remind user to:**
  - Push feature branch: `git push -u origin feat/<feature-name>`
  - Cleanup crew workspace: `rig crew remove <worker-name>`
  - Create pull request if needed

## Outputs
- Updated `design.md` - Design document
- Updated `breakdown.md` - Implementation tasks
- Updated `progress.md` - Progress tracking with status
- Feature implementation with test coverage
- Git commits following conventional commits
