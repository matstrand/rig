
# Feature Implementation Skill

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
0. Create a feature branch base on the base branch (see --branch and --base params above)
1. Survey existing codebase for patterns to follow
2. Identify files to create/modify
3. Design module structure and interfaces
4. Plan test strategy (unit, integration, e2e)
5. Infer a short feature name from the spec
5. Create `design/<feature_name>.md` with:
   - Files to change
   - New abstractions needed
   - Testing approach
   - Risk areas

**Gate:** Review design. If major concerns, revise. Otherwise continue.

### Phase 3: Implementation Planning
Break design into tasks. Each task should:
- Be completable in one session
- Have clear done criteria
- Be independently testable
- Produce a commit

Create a checklist of the tasks.

```

### Phase 4: Implementation
For each task:
1. Work on task until done criteria met
2. Run relevant tests
3. Commit with message: `feat: [task description]`
4. Update checklist

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

**Gate:** If major issues, fix and re-review. Otherwise continue.

### Phase 6: Final Review & Commit
1. Run full test suite
2. Update documentation
3. Final code review
4. Commit with summary

## Outputs
- `design/DESIGN.md` - Design document
- Updated spec with completed checklist
- Feature implementation with test coverage
- Git commits following conventional commits
