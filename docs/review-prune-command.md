# Review: Prune Command Implementation

**Branch:** `claude/implement-prune-command-01A6FtNY74y4sGeHGbZZmcSe`
**Reviewer:** Claude
**Date:** 2025-11-16

## Summary

The prune command implementation is **well-structured and largely complete**. It successfully implements the core functionality for safely pruning old projects to free disk space. The implementation correctly prioritizes safety with dry-run defaults, proper verification, and edge case handling.

## Requirements Checklist

### ✅ Fully Implemented

| Requirement | Status | Notes |
|-------------|--------|-------|
| Default is dry-run | ✅ PASS | `Execute: false` in main.go:prune case |
| parkr prune \<size> shows what would be deleted | ✅ PASS | `outputDryRun()` provides clear preview |
| --exec actually deletes projects | ✅ PASS | `ExecutePrune()` performs deletions |
| Selects oldest candidates first | ✅ PASS | Uses `GenerateReport()` which sorts by `LastModified` |
| Stops when target size reached | ✅ PASS | Loop exits when `totalSelected >= targetBytes` |
| Respects no_hash_mode settings | ✅ PASS | `verifyBeforeDeletion()` checks `project.NoHashMode` |
| --no-hash uses mtime verification | ✅ PASS | Properly passed through and respected |
| --force skips verification | ✅ PASS | Includes warning: "Data may be lost!" |
| Uses size parsing from core/size.go | ✅ PASS | `core.ParseSize()` called in CLI |
| Uses report logic from core/report.go | ✅ PASS | `core.GenerateReport()` used for candidate selection |
| Handles insufficient space | ✅ PASS | `InsufficientSpace` flag with appropriate warnings |
| Handles no candidates | ✅ PASS | `NoCandidates` flag with informative message |
| Handles all dirty projects | ✅ PASS | Test case `TestSelectPruneCandidates_NoCandidates_AllDirty` |

### ⚠️ Missing/Incomplete Features

| Feature | Status | Details |
|---------|--------|---------|
| --interactive flag | ❌ NOT IMPLEMENTED | Mentioned in TEST-phase-5.md but not in code |

### Code Quality Analysis

#### Strengths

1. **Safety First Design** (`code/core/prune.go:128-157`)
   - Re-verifies projects immediately before deletion
   - Catches modifications that occurred between selection and execution
   - Saves state after each successful deletion (atomic operations)

2. **Clear Separation of Concerns**
   - CLI layer (`code/cli/prune.go`) handles user interaction
   - Core layer (`code/core/prune.go`) handles business logic
   - Uses existing infrastructure (size parsing, report generation, hash computation)

3. **Comprehensive Output**
   - Dry-run shows numbered list with sizes
   - Execution mode shows real-time progress with checkmarks/crosses
   - Clear warnings for force mode and insufficient space

4. **Good Test Coverage**
   - Tests for empty state, dirty projects, oldest-first selection
   - Tests for target stopping, insufficient space, force mode
   - Tests for actual deletion and state updates
   - Tests for verification logic

#### Potential Issues

1. **Duplicated Verification Logic** (`code/core/prune.go:139-198`)
   ```go
   if !opts.Force {
       // ... 40 lines of verification and deletion ...
   } else {
       // ... 25 lines of nearly identical code ...
   }
   ```
   The force/non-force branches duplicate most of the deletion logic. Consider refactoring to reduce duplication.

2. **No Progress Callback Error Handling** (`code/core/prune.go:117`)
   ```go
   progressFn func(project ProjectReport, success bool, freed int64)
   ```
   The progress function could panic without recovery. Consider wrapping in defer/recover.

3. **State Manager Injection Pattern** (`code/core/prune.go:112`)
   ```go
   var newStateManagerFn = func() *StateManager {
       return NewStateManager()
   }
   ```
   This global function for testing is acceptable but could be made cleaner with dependency injection through function parameters.

4. **Inconsistent Error Handling in ExecutePrune**
   - When state save fails, the directory is already deleted but not recorded
   - This could lead to orphaned state entries
   - Consider: transaction-like behavior or better error recovery

5. **Missing Total Count in Dry-Run** (`code/cli/prune.go:80-90`)
   ```go
   fmt.Printf("Total to free: %s (target: %s)\n",
       core.FormatSize(result.TotalSelected),
       core.FormatSize(result.TargetBytes))
   ```
   Would be helpful to show "N projects" count.

### Test Coverage Analysis

**Covered Scenarios:**
- ✅ Empty state handling
- ✅ All dirty projects (no candidates)
- ✅ Oldest-first selection
- ✅ Stop at target size
- ✅ Insufficient space detection
- ✅ Force mode includes dirty projects
- ✅ Actual deletion execution
- ✅ Skip modified projects during execution
- ✅ Never-parked project safety check
- ✅ Safe deletion with mtime verification
- ✅ Unsafe deletion with modified files

**Missing Test Scenarios:**
- ❌ Hash-based verification path (only mtime tests)
- ❌ Error during state save after deletion
- ❌ Progress callback functionality
- ❌ Multiple projects with same modification time
- ❌ Very large number of candidates (performance)
- ❌ CLI integration tests for prune command
- ❌ FormatSize output in dry-run

### Security Considerations

✅ **Good:**
- Symlinks are not followed during hash computation
- Re-verification before actual deletion prevents TOCTOU races
- Force mode requires explicit flag with clear warning
- No shell injection vectors

⚠️ **Consider:**
- Race condition window between candidate selection and deletion (mitigated by re-verification)
- No locking mechanism for concurrent prune operations

### Recommendations

#### High Priority

1. **Add --interactive flag** (if required by spec)
   ```go
   // Allow user to select which projects to prune interactively
   case "--interactive":
       pruneOpts.Interactive = true
   ```

2. **Refactor duplicated deletion logic**
   ```go
   // Extract common deletion logic into helper function
   func deleteSingleProject(project *Project, result *PruneResult, verify bool) error {
       // Common deletion code here
   }
   ```

3. **Add hash-based verification tests**
   ```go
   func TestVerifyBeforeDeletion_WithHash(t *testing.T) {
       // Test the hash verification path
   }
   ```

#### Medium Priority

4. **Improve state save error handling**
   ```go
   // If state save fails, try to restore the directory or mark as inconsistent
   if err := sm.Save(state); err != nil {
       // Attempt rollback or log for manual recovery
   }
   ```

5. **Add project count to dry-run output**
   ```go
   fmt.Printf("Total to free: %s from %d projects (target: %s)\n",
       core.FormatSize(result.TotalSelected),
       len(result.SelectedProjects),
       core.FormatSize(result.TargetBytes))
   ```

6. **Add CLI integration tests**
   ```go
   func TestPruneCmd_DryRunOutput(t *testing.T) {
       // Capture stdout and verify dry-run output format
   }
   ```

#### Low Priority

7. **Consider adding confirmation prompt for large deletions**
8. **Add logging for audit trail**
9. **Consider parallel deletion for performance (with careful state management)**

## Conclusion

**Overall Grade: B+**

The prune command implementation is solid and production-ready for most use cases. The safety-first approach with dry-run defaults, re-verification before deletion, and proper edge case handling demonstrates good software engineering practices.

**Strengths:**
- Excellent safety mechanisms
- Clean code organization
- Good integration with existing codebase
- Comprehensive error reporting to users

**Areas for Improvement:**
- Missing --interactive flag (if required)
- Some code duplication that could be refactored
- A few edge cases in tests not covered
- Minor output formatting improvements

The implementation successfully meets the core requirements and can be safely used. The missing --interactive flag should be clarified whether it's a requirement or future enhancement.
