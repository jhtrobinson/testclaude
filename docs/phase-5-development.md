# Phase 5: Space Management - Development Plan

## Overview

Phase 5 adds automated cleanup features to help users free up disk space by identifying and removing local project copies that are safely backed up in the archive.

## Feature Branches

### Branch 1: `feature/size-parsing`

**Scope:** Core utility for parsing size strings

- Parse human-readable sizes: `10G`, `500M`, `2T`, `1.5GB`
- Convert to bytes and back to human-readable format
- Unit tests for parsing logic

**Dependencies:** None

**Files to create/modify:**
- `core/size.go` - Size parsing and formatting functions
- `core/size_test.go` - Unit tests

---

### Branch 2: `feature/report-command`

**Scope:** Disk usage analysis and reporting

- `parkr report` - Show disk usage for local projects
- `parkr report --candidates` - Show safe-to-delete projects
- `parkr report --recompute-hashes` - Thorough safety check
- `--sort <field>` option (size|modified|name)
- `--json` output option
- Compute sizes on-demand for local projects

**Dependencies:** None (uses existing hash verification from Phase 2)

**Files to create/modify:**
- `cli/report.go` - Report command implementation
- `core/report.go` - Report logic (size computation, candidate detection)
- `core/report_test.go` - Unit tests

---

### Branch 3: `feature/prune-command`

**Scope:** Full pruning functionality (dry-run + execution)

- `parkr prune <size>` - Dry-run mode (default)
- `parkr prune --exec` - Actually delete projects
- Select candidates sorted by last modified (oldest first)
- Verification logic respecting `no_hash_mode`
- `--no-hash` and `--force` flags
- Progress reporting during deletion

**Dependencies:**
- Size parsing (Branch 1)
- Report logic (Branch 2)

**Files to create/modify:**
- `cli/prune.go` - Prune command implementation
- `core/prune.go` - Pruning logic
- `core/prune_test.go` - Unit tests

---

### Branch 4: `feature/prune-interactive`

**Scope:** Interactive project selection UI

- `parkr prune --interactive` - User selects projects
- Terminal UI for toggling selections
- Keyboard controls (space to toggle, a for all, enter to confirm)
- Running total of selected space

**Dependencies:** Prune command (Branch 3)

**Files to create/modify:**
- `cli/prune.go` - Add interactive mode
- `core/interactive.go` - Terminal UI logic (or use a library like bubbletea)

---

## Development Order

```
Branch 1 (size-parsing) ──┐
                          ├──> Branch 3 (prune-command) ──> Branch 4 (prune-interactive)
Branch 2 (report-command) ┘
```

1. **Parallel:** Branches 1 & 2 can be developed simultaneously
2. **Sequential:** Branch 3 requires both 1 & 2 to be complete
3. **Sequential:** Branch 4 requires Branch 3 to be complete

## Integration

After all feature branches are complete and merged to main:
1. Create `phase-5-integration` branch
2. Add comprehensive integration tests
3. Update documentation
4. Create TEST-phase-5.md test script
