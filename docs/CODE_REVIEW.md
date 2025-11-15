# Code Review: parkr

**Date:** 2025-11-15
**Reviewer:** Claude Code

## Summary

This document captures issues found during a comprehensive code review of the parkr codebase, including code duplication, anti-patterns, bugs, and design issues.

---

## 1. Code Duplication

### State Loading Pattern
**Files:** `cli/grab.go:14-18`, `cli/park.go:13-17`, `cli/list.go:13-17`, `cli/rm.go:12-16`

```go
sm := core.NewStateManager()
state, err := sm.Load()
if err != nil {
    return err
}
```

**Recommendation:** Consider extracting to a helper function or using middleware pattern.

### Rsync Trailing Slash Logic
**Files:** `core/rsync.go:11-13` and `core/rsync.go:27-29`

```go
if src[len(src)-1] != '/' {
    src = src + "/"
}
```

**Recommendation:** Extract to a helper function like `ensureTrailingSlash()`.

### "Not Grabbed" Check
**Files:** `cli/park.go:20-22` and `cli/rm.go:19-21`

```go
project, exists := state.Projects[projectName]
if !exists || !project.IsGrabbed {
    return fmt.Errorf("project '%s' is not currently grabbed", projectName)
}
```

**Recommendation:** Add a `State.GetGrabbedProject()` method that returns the project or an appropriate error.

---

## 2. Anti-Patterns & Bugs

### Critical: Ignored Errors

**Location:** `core/state.go:40` and `core/state.go:150`

```go
homeDir, _ := os.UserHomeDir()  // Error silently ignored
```

**Impact:** This will panic or produce incorrect paths if home directory can't be determined.

**Recommendation:** Handle the error properly:
```go
homeDir, err := os.UserHomeDir()
if err != nil {
    return nil, fmt.Errorf("failed to get home directory: %w", err)
}
```

### Critical: Potential Panic

**Location:** `core/rsync.go:11` and `core/rsync.go:27`

```go
if src[len(src)-1] != '/' {  // Panics if src is empty string
```

**Impact:** No validation that `src` is non-empty will cause a panic.

**Recommendation:** Add input validation:
```go
if src == "" {
    return fmt.Errorf("source path cannot be empty")
}
```

### Bug: Confusing Nil Check

**Location:** `cli/park.go:58` and `cli/rm.go:53`

```go
if newestInfo != nil && *newestInfo != nil {
```

**Issue:** `GetNewestMtime` returns `*os.FileInfo`, but if no files exist in the directory, `newest` is never assigned. The function returns `&newest` which is a pointer to a nil interface - this is very confusing and error-prone.

**Recommendation:** Refactor `GetNewestMtime` to return `(os.FileInfo, bool, error)` or `(os.FileInfo, error)` where a missing file is indicated differently.

### Hardcoded Magic Paths

**Location:** `core/state.go:114-118`

```go
"code":    "/Volumes/Extra/project-archive/code",  // macOS-specific hardcoded path
```

**Impact:** These hardcoded paths are macOS-specific and won't work on other platforms.

**Recommendation:** Use environment variables or configuration file for these paths.

---

## 3. Design Issues

### No CLI Argument Parsing Library

**Location:** `main.go`

The code manually parses arguments. Consider using `flag`, `cobra`, or `urfave/cli` for:
- Better help messages
- Automatic validation
- Subcommand support
- Flag handling
- Shell completion

### Performance Issue in ListCmd

**Location:** `cli/list.go:57`

```go
size, err := core.GetDirSize(ap.Path)  // Walks entire tree for EVERY project
```

**Impact:** This scans every file in every project directory sequentially - extremely slow for large archives with many projects.

**Recommendation:**
- Make size calculation optional (add `--size` flag)
- Cache sizes
- Use concurrent workers to calculate sizes in parallel

### No Context Support

Long-running operations (rsync, directory walking) don't accept `context.Context` for cancellation.

**Impact:**
- Can't cancel long-running operations
- No timeout support
- Poor user experience for large projects

**Recommendation:** Add context support to all long-running functions.

### No Input Validation

**Issues:**
- No check that project names are valid filesystem names
- No maximum path length validation
- No sanitization of user input
- No check for path traversal attacks (e.g., `../` in project names)

**Recommendation:** Add comprehensive input validation layer.

---

## 4. Missing Functionality

### No Tests
No `*_test.go` files anywhere in the codebase.

**Recommendation:** Add unit tests for:
- State management
- Archive discovery
- Size formatting
- Path construction

### No Concurrent Operation Safety
State file has no locking mechanism.

**Impact:** Multiple concurrent parkr operations could corrupt the state file.

**Recommendation:** Implement file locking using `flock` or similar.

### Hash Verification Not Implemented
Phase 2 comments mention hash verification but it's not implemented.

**Location:** `cli/rm.go:63-65`

### No Logging
Only `fmt.Printf` for output - no structured logging.

**Recommendation:** Add structured logging with log levels for debugging.

---

## 5. Minor Issues

### Misleading Comment in RsyncWithProgress

**Location:** `core/rsync.go:32-33`

```go
cmd.Stdout = nil // Will be displayed directly
cmd.Stderr = nil
```

Setting to `nil` means output goes to parent process stdout, not that it's explicitly nil. The comment is misleading.

### Inconsistent Error Wrapping
Some errors use `%w` for wrapping, others don't. This should be consistent.

### No .gitignore
The `parkr` binary is in the repository root but there's no `.gitignore` file.

### No Documentation
- No README.md explaining how to use the tool
- No godoc comments on exported functions (some present, but incomplete)
- No architecture documentation

---

## Priority Recommendations

1. **High:** Fix ignored errors in `os.UserHomeDir()` calls
2. **High:** Add empty string validation in rsync functions
3. **High:** Add basic input validation for project names
4. **Medium:** Add unit tests for core functionality
5. **Medium:** Implement file locking for state file
6. **Medium:** Extract duplicated code to helper functions
7. **Low:** Consider using a CLI framework
8. **Low:** Add structured logging
