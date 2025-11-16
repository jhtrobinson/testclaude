# Phase Integration Branch Management

## Overview

This document describes the process for managing feature development within a phase integration branch. The process uses multiple AI agents coordinated by an integration branch manager to implement, review, and refine features before merging to the main codebase.

## Roles

### Programmer (Human)
- Initiates the phase development
- Provides prompts to agents
- Monitors agent progress
- Makes final merge decisions
- Coordinates timing between agents

### Integration Branch Manager (AI Agent)
- Tracks all feature branches and their status
- Writes prompts for implementation and review agents
- Maintains todo list of feature progress
- Records branch IDs and relationships
- Documents the process

### Implementation Agents (AI Agents)
- Create feature branches from integration branch
- Implement specific features based on spec and test cases
- Reference documentation for requirements
- Test their implementations
- Respond to reviewer feedback
- Fix issues identified in reviews

### Reviewer Agents (AI Agents)
- Create review branches from implementation branches
- Analyze code quality and correctness
- Check implementation against spec and test cases
- Document issues in review files
- Re-review after fixes are applied
- Approve or request further changes

## Branch Structure

```
main
  └── phase-N-integration
        ├── feature/phase-N-feature-A
        │     ├── claude/impl-feature-A-{uuid}     (implementation)
        │     │     └── claude/review-feature-A-{uuid}  (review)
        │     └── (merged result)
        ├── feature/phase-N-feature-B
        │     ├── claude/impl-feature-B-{uuid}     (implementation)
        │     │     └── claude/review-feature-B-{uuid}  (review)
        │     └── (merged result)
        └── ... (additional features)
```

### Branch Naming Convention
- Integration branch: `phase-N-integration`
- Feature branches: `feature/phase-N-<feature-name>`
- Implementation branches: `claude/<feature-description>-<uuid>`
- Review branches: `claude/review-<feature-description>-<uuid>`

## Process Flow

### 1. Setup Phase
**Prerequisite**: `phase-N-integration` branch exists with:
- Updated spec in docs/parkr.spec.md
- Test cases in docs/TEST-phase-N.md
- Any other requirements documentation

Integration branch manager:
- Creates `phase-memo.txt` in project root to track all branch information persistently
- Identifies features to implement from the requirements
- **Creates all feature branches** from the integration branch (e.g., `feature/phase-N-feature-A`)
- **Pushes feature branches to origin** so implementation agents can access them
- Creates todo list tracking each feature's progress
- Documents branch structure and feature assignments
- Updates phase-memo.txt with each new branch ID as agents are assigned

### 2. Implementation Phase
For each feature:
1. **Programmer provides implementation prompt** to an agent
2. **Implementation agent**:
   - Creates branch from the feature branch (NOT directly from integration branch)
   - Reads spec (docs/parkr.spec.md) and test cases (docs/TEST-phase-N.md)
   - Implements the feature
   - Tests the implementation
   - Commits work
3. **Integration branch manager records branch ID**

### 3. Review Phase
For each feature:
1. **Integration branch manager writes review prompt**
2. **Programmer assigns reviewer agent**
3. **Reviewer agent**:
   - Creates branch from implementation branch
   - Reviews code against spec and test cases
   - Documents findings in `docs/review-<feature>.md`
   - Commits review document
4. **Integration branch manager records review branch ID**

### 4. Fix Phase
For each feature with review issues:
1. **Integration branch manager writes fix prompt**
2. **Programmer prompts implementation agent**
3. **Implementation agent**:
   - Merges review branch into implementation branch
   - Reads review document
   - Addresses issues by fixing code
   - Writes response in `docs/review-<feature>-response.md`
   - Commits fixes and response

### 5. Re-Review Phase
For each feature:
1. **Integration branch manager writes re-review prompt**
2. **Programmer prompts reviewer agent**
3. **Reviewer agent**:
   - Merges implementation branch to get fixes
   - Reads implementer's response
   - Reviews the fixes
   - Either approves (adds "APPROVED" to review doc) or requests further changes
   - Commits decision

### 6. Iteration (if needed)
If reviewer requests further changes:
- Repeat Fix Phase and Re-Review Phase
- Continue until approval

### 7. Feature Branch Merge Phase
For each feature:
1. **Integration branch manager checks approval status**
   - Fetches latest reviewer branch from origin
   - Checks if review document contains "APPROVED"
2. **If approved, integration branch manager merges**:
   - Checks out the feature branch (based on integration branch)
   - Merges the reviewer branch into the feature branch
   - The reviewer branch contains both implementation and review docs
   - If conflicts occur: escalate to programmer (this should not normally happen)
   - Commits the merge
3. **Updates TODO list** with merge status
4. **Pushes feature branches to origin**
   - After all approved features are merged, push feature branches to remote
   - This is required before merge agent can access them
   - Merge agents do not have write access to integration branch

### 8. Pre-Merge Branch Creation
1. **Integration branch manager creates pre-merge branch**
   - Creates `phase-N-integration-premerge` branch from `phase-N-integration`
   - **Pushes branch to origin immediately** so merge agent can access it
   - This branch receives all merged features for testing
   - Named clearly to distinguish from agent branches
   - Provides clean staging area for validation

### 9. Integration Branch Merge Phase
For each feature:
1. **Integration branch manager writes merge prompt**
2. **Programmer prompts merge agent**
3. **Integration branch manager records merge agent branch ID** in phase-memo.txt
4. **Merge agent**:
   - Creates TODO list to track progress
   - Processes one feature at a time
   - Merges feature branch into pre-merge branch (`phase-N-integration-premerge`)
   - If conflicts occur: reads spec for conflicting features, resolves ensuring both work
   - If resolution not straightforward: stops and reports to programmer
   - Reports status and waits for programmer before next feature
5. **Repeat** until all features merged

### 10. Pre-Merge Validation
**Programmer responsibility** - must be done by the programmer, not agents:
- Verify all features compile/build correctly
- Run all phase test cases (preferably in isolated environment)
- Fix any integration issues
- Final review before merging to main

### 11. Final Integration Branch Update
**Integration branch manager responsibility**:
1. Merge the pre-merge branch back into the integration branch
   - This preserves all work done on the integration branch (docs updates, tracking changes, etc.)
   - `git checkout phase-N-integration && git merge phase-N-integration-premerge`
2. Push the updated integration branch to origin
3. The integration branch now contains all features plus any integration work

## Branch Tracking

The integration branch manager must maintain a persistent record of all branch information in `phase-memo.txt` at the project root. This file should be updated continuously as agents are assigned and branches are created.

**Purpose**: Allows the conversation to be resumed if the chat session is interrupted or restarted.

**Contents**:
- Phase number and integration branch name
- Pre-merge branch name (when created)
- For each feature:
  - Feature branch name
  - Implementation agent branch ID
  - Review agent branch ID
  - Current status (implementing, reviewing, fixing, approved, merged)
  - Any notes about issues or conflicts

**Example format**:
```
Phase: 4
Integration Branch: phase-4-integration
Pre-merge Branch: phase-4-integration-premerge

Feature: add-move
  Feature Branch: feature/phase-4-add-move
  Implementation: claude/add-move-option-01HceGD8kh5ZTiNVZWsN3Nx2
  Review: claude/review-add-move-option-01FuiG3F5EsshZJtKbVtajrh
  Status: approved
  Notes: None

Feature: grab-force
  Feature Branch: feature/phase-4-grab-force
  Implementation: claude/grab-force-option-01T9c8yBkhqTtrjq8xcLACGc
  Review: claude/review-grab-force-option-01FUxC2SBGh1qox6qku5HoCP
  Status: merged
  Notes: None
```

**Important**: This file should be added to `.gitignore` as it contains session-specific information.

## Documentation Structure

Each feature generates the following documents:
- `docs/review-<feature>.md` - Reviewer's findings and issues
- `docs/review-<feature>-response.md` - Implementer's response to review

## Agent Prompt Templates

### Implementation Agent Prompt

```
Create branch feature/phase-N-<feature-name> from phase-N-integration.

Implement the <feature description>.

Reference:
- docs/parkr.spec.md (<relevant section>)
- docs/TEST-phase-N.md (<relevant test cases>)

Test your implementation and commit your work.
```

### Review Agent Prompt

```
Create branch from claude/<implementation-branch-id>.

Review the <feature description> implementation.

Reference:
- docs/parkr.spec.md (<relevant section>)
- docs/TEST-phase-N.md (<relevant test cases>)

Write your review findings to docs/review-<feature>.md

Commit your review document.
```

### Fix Agent Prompt (to Implementation Agent)

```
Merge branch claude/<review-branch-id> into your implementation branch.

Read the review at docs/review-<feature>.md

Address the review issues by fixing your implementation as appropriate.

Write your response to the reviewer at docs/review-<feature>-response.md

Commit your fixes and response.
```

### Re-Review Agent Prompt (to Reviewer Agent)

```
Merge branch claude/<implementation-branch-id> into your review branch.

Read the implementer's response at docs/review-<feature>-response.md

Review the fixes made to the implementation.

If further changes are required, update docs/review-<feature>.md with new issues.

If approved, add "APPROVED" to the top of docs/review-<feature>.md

Commit your decision.
```

### Merge Agent Prompt

```
You are the merge agent for Phase N. Your job is to merge feature branches into the integration branch.

Create a TODO list to track your progress with one item per feature.

Feature branches to merge into phase-N-integration:
- feature/phase-N-<feature-A>
- feature/phase-N-<feature-B>
- feature/phase-N-<feature-C>
...

For each feature (one at a time):

1. Mark the feature as in_progress in your TODO list
2. Checkout phase-N-integration
3. Merge the feature branch into phase-N-integration
4. If merge conflicts occur:
   - Read docs/parkr.spec.md for the conflicting features
   - Resolve conflicts ensuring both features work correctly
   - If resolution is not straightforward, stop and report the conflict details to programmer
5. Commit the merge
6. Mark as completed in TODO
7. Report to programmer and wait for instructions

Process one feature, report back, then wait for programmer before proceeding to the next.
```

## Example: Phase 4 Features

The following shows a concrete example of managing Phase 4 features:

### Features to Implement

| Feature | Description |
|---------|-------------|
| add-move | --move option for add command |
| grab-force | --force option for grab/checkout |
| grab-to | --to option for grab/checkout |
| remove | remove command (state only) |
| remove-archive | --archive flag for remove |

### Example Branch Tracking

| Feature | Implementation Branch | Review Branch |
|---------|----------------------|---------------|
| add-move | claude/add-move-option-01HceGD8kh5ZTiNVZWsN3Nx2 | claude/review-add-move-option-01FuiG3F5EsshZJtKbVtajrh |
| grab-force | claude/grab-force-option-01T9c8yBkhqTtrjq8xcLACGc | claude/review-grab-force-option-01FUxC2SBGh1qox6qku5HoCP |
| grab-to | claude/add-grab-to-option-0166556DCMtA1jNJLBBpd2r8 | claude/review-grab-to-option-01GisbPTbbbC17D8gq5EYkxp |
| remove | claude/parkr-remove-command-01Vahic1RscXVEHivDVg9sH6 | claude/review-remove-command-01DQNjQyWyMUK1zLxZ8eo15Q |
| remove-archive | claude/add-remove-archive-flag-01YKyPbSeq6M11PtbjskA6Dx | claude/review-remove-archive-flag-01LGjsc1fNLgMUrFjmrQJPzR |

### Example Implementation Prompt

```
Create branch feature/phase-4-grab-force from phase-4-integration.

Implement the --force option for the parkr grab/checkout command.

Reference:
- docs/parkr.spec.md (checkout command section)
- docs/TEST-phase-4.md (--force test cases)

Test your implementation and commit your work.
```

### Example Review Prompt

```
Create branch from claude/grab-force-option-01T9c8yBkhqTtrjq8xcLACGc.

Review the --force option implementation for the parkr grab/checkout command.

Reference:
- docs/parkr.spec.md (checkout command section)
- docs/TEST-phase-4.md (--force test cases)

Write your review findings to docs/review-grab-force-option.md

Commit your review document.
```

### Example Fix Prompt

```
Merge branch claude/review-grab-force-option-01FUxC2SBGh1qox6qku5HoCP into your implementation branch.

Read the review at docs/review-grab-force-option.md

Address the review issues by fixing your implementation as appropriate.

Write your response to the reviewer at docs/review-grab-force-option-response.md

Commit your fixes and response.
```

### Example Re-Review Prompt

```
Merge branch claude/grab-force-option-01T9c8yBkhqTtrjq8xcLACGc into your review branch.

Read the implementer's response at docs/review-grab-force-option-response.md

Review the fixes made to the implementation.

If further changes are required, update docs/review-grab-force-option.md with new issues.

If approved, add "APPROVED" to the top of docs/review-grab-force-option.md

Commit your decision.
```

## Benefits of This Approach

1. **Parallel development** - Multiple features can be implemented simultaneously
2. **Independent review** - Reviewers have fresh perspective on code
3. **Documented feedback** - All review comments and responses are preserved
4. **Clear accountability** - Each agent has specific responsibilities
5. **Traceability** - Branch IDs link all related work together
6. **Quality assurance** - Multiple rounds of review ensure correctness

## Reference Documents

- `docs/parkr.spec.md` - Main specification for all commands
- `docs/TEST-phase-N.md` - Test cases for phase features
- `docs/TODO.md` - Overall project progress
