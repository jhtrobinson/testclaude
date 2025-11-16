# Parkr Code Review - Refactoring Plan

## Overview
This document outlines code smells, bugs, deficiencies, and antipatterns identified in the Parkr codebase, along with recommendations for refactoring.

---

## 🐛 BUGS (Critical - Fix Immediately)

### 1. Nil Pointer Dereference in Rsync Functions
**Location:** `core/rsync.go:11,29`
```go
if src[len(src)-1] != '/' {  // PANIC if src is empty string
```
**Impact:** Will panic if empty string passed
**Fix:** Add length check before accessing index

### 2. Double Dereference Anti-pattern
**Location:** `cli/park.go:58`, `cli/rm.go:53`
```go
if newestInfo != nil && *newestInfo != nil {
    mtime := (*newestInfo).ModTime()
```
**Impact:** Confusing semantics, potential nil dereference
**Fix:** Change `GetNewestMtime` return type (see Antipatterns #1)

### 3. Ignored Errors on UserHomeDir
**Location:** `core/state.go:40,150`
```go
homeDir, _ := os.UserHomeDir()  // Error silently ignored
```
**Impact:** Silent failures, broken paths if home dir unavailable
**Fix:** Return error or panic with clear message

### 4. Empty Directory Returns Nil Without Error
**Location:** `core/archive.go:54-76`
**Impact:** If directory has no files, returns nil FileInfo without indication
**Fix:** Return explicit error for empty directories

---

## 🦨 CODE SMELLS (Medium Priority)

### 1. Manual Argument Parsing
**Location:** `main.go`
**Issue:** Hand-rolled argument parsing instead of using cobra/urfave/cli
**Impact:** Difficult to extend, inconsistent behavior, no auto-completion
**Recommendation:** Migrate to cobra as spec originally intended

### 2. Repeated StateManager Instantiation
**Location:** All CLI commands
```go
sm := core.NewStateManager()  // Created fresh each time
```
**Issue:** No dependency injection
**Recommendation:** Pass StateManager as parameter or use singleton pattern

### 3. Magic Strings Scattered Throughout
**Locations:**
- `state.go:113-118`: `"primary"`, `"/Volumes/Extra/..."`
- `state.go:153-158`: `"pycharm"`, `"rstudio"`, `"code"`
- `main.go:82-83`: Exit code numbers

**Recommendation:** Define constants in a central location

### 4. No Test Coverage
**Issue:** Zero test files exist
**Impact:** Regressions likely, refactoring risky
**Recommendation:** Add unit tests before refactoring

### 5. Hardcoded Platform-Specific Paths
**Location:** `core/state.go:113-118`
```go
"code": "/Volumes/Extra/project-archive/code",
```
**Issue:** macOS-specific, won't work elsewhere
**Recommendation:** Use environment variables or config file

### 6. Inconsistent Error Handling
**Issue:** Some functions wrap errors with context, others don't
**Recommendation:** Establish error wrapping convention using `fmt.Errorf("context: %w", err)`

---

## 🏗️ ANTIPATTERNS (High Priority)

### 1. Returning Pointer to Interface
**Location:** `core/archive.go:54`
```go
func GetNewestMtime(dirPath string) (*os.FileInfo, error)
```
**Issue:** `os.FileInfo` is already an interface; pointer to interface is almost never needed
**Fix:** Return `time.Time, error` directly:
```go
func GetNewestMtime(dirPath string) (time.Time, error)
```

### 2. Temporal Coupling
**Issue:** Must call `Load()` before `Save()`, not enforced by type system
**Impact:** Easy to save without loading, causing data loss
**Recommendation:** Consider transaction pattern or make Save require loaded state

### 3. No Separation of Concerns in CLI Layer
**Location:** All CLI commands (e.g., `cli/grab.go`)
**Issue:** CLI functions perform:
- Business logic
- File system operations
- State management
- User output

**Recommendation:**
- Core layer: pure business logic
- CLI layer: input/output only
- Inject dependencies

### 4. God Object Potential
**Location:** `core/state.go` - `State` struct
**Issue:** Mixes configuration (masters, default_master) with runtime data (projects)
**Recommendation:** Separate `Config` from `ProjectRegistry`

### 5. Silent Failures in RsyncWithProgress
**Location:** `core/rsync.go:32-33`
```go
cmd.Stdout = nil // Will be displayed directly
cmd.Stderr = nil
```
**Issue:** Sets output to nil but comment says "displayed directly" - actually silences output
**Fix:** Properly connect to os.Stdout/os.Stderr or remove misleading comment

---

## ⚠️ DEFICIENCIES (Address During Refactoring)

### Security & Validation

1. **No Input Validation**
   - Project names could contain path traversal (`../../../etc`)
   - No length limits on strings
   - No validation of category names
   - **Risk:** Directory traversal attacks, file system corruption

2. **No Concurrency Safety**
   - State file has no locking mechanism
   - **Risk:** Data corruption with concurrent access

### Missing Infrastructure

3. **No Logging Framework**
   - Only `fmt.Printf` for output
   - No log levels, no structured logging
   - **Recommendation:** Add zerolog or zap

4. **No Graceful Shutdown**
   - Long rsync operations can't be cancelled
   - No signal handling
   - **Recommendation:** Use context.Context with cancellation

5. **No Configuration File**
   - Paths are hardcoded
   - **Recommendation:** Support `~/.parkr/config.yaml`

### Implementation Gaps

6. **Missing --force Flag in Grab**
   - Referenced in error message (`grab.go:42`) but not implemented
   - **Fix:** Add flag parsing and implementation

7. **Platform Assumptions**
   - Assumes `rsync` exists on system
   - macOS-specific paths (`/Volumes/`)
   - No Windows support
   - **Recommendation:** Check for rsync, provide fallback or clear error

8. **Missing Features from Spec**
   - No `--all` flag support for batch operations
   - No `--json` output option
   - No progress indicators for rsync
   - No hash verification (Phase 2 stub only)

9. **Incomplete Error Recovery**
   - No idempotency guarantees
   - State may be inconsistent after failures
   - **Recommendation:** Implement transaction-like semantics

10. **Resource Leak Potential**
    - Partial cleanup on failures (grab.go:60 attempts this)
    - No defer patterns for cleanup
    - **Recommendation:** Use defer consistently

---

## Refactoring Priority Order

### Phase 1: Critical Fixes (1-2 days)
1. Fix nil pointer dereferences in rsync.go
2. Fix GetNewestMtime return type
3. Add input validation (path traversal protection)
4. Handle UserHomeDir errors properly

### Phase 2: Architecture Improvements (3-5 days)
1. Migrate to cobra CLI framework
2. Separate business logic from CLI
3. Add proper error handling strategy
4. Implement dependency injection

### Phase 3: Testing & Safety (3-5 days)
1. Add unit tests for core package
2. Add integration tests for CLI
3. Implement state file locking
4. Add input sanitization

### Phase 4: Feature Completion (5-7 days)
1. Implement --force flag properly
2. Add --json output support
3. Add --all batch operations
4. Implement hash verification (Phase 2 from TODO)

### Phase 5: Production Readiness (3-5 days)
1. Add structured logging
2. Add configuration file support
3. Implement graceful shutdown
4. Add platform detection and rsync checks

---

## Code Locations Quick Reference

| File | Lines | Issues |
|------|-------|--------|
| `main.go` | 99 | Manual arg parsing, magic numbers |
| `core/state.go` | 161 | Ignored errors, hardcoded paths, god object |
| `core/rsync.go` | 41 | Nil pointer risk, silent failures |
| `core/archive.go` | 117 | Bad return type, nil return |
| `cli/grab.go` | 82 | Missing --force, no separation of concerns |
| `cli/park.go` | 73 | Double dereference, temporal coupling |
| `cli/rm.go` | 86 | Double dereference, complex conditionals |
| `cli/list.go` | 68 | Synchronous size calculation (slow) |
| `cli/init.go` | 24 | Cleanest file, minor issues |

---

## 🔁 DUPLICATED CODE (Extract Common Patterns)

### 1. StateManager Load Pattern
**Location:** `cli/grab.go:14-18`, `cli/park.go:13-17`, `cli/rm.go:13-17`, `cli/list.go:13-17`
```go
sm := core.NewStateManager()
state, err := sm.Load()
if err != nil {
    return err
}
```
**Fix:** Create middleware/decorator or base command struct

### 2. Project Grabbed Validation
**Location:** `cli/park.go:20-22`, `cli/rm.go:19-21`
```go
project, exists := state.Projects[projectName]
if !exists || !project.IsGrabbed {
    return fmt.Errorf("project '%s' is not currently grabbed", projectName)
}
```
**Fix:** Extract to `state.GetGrabbedProject(name)` helper

### 3. Local Path Existence Check
**Location:** `cli/park.go:26-28`, `cli/rm.go:25` (similar pattern)
```go
if _, err := os.Stat(project.LocalPath); os.IsNotExist(err) {
    return fmt.Errorf("local path does not exist: %s", project.LocalPath)
}
```
**Fix:** Extract to `project.VerifyLocalPathExists()` method

### 4. Rsync Trailing Slash Logic
**Location:** `core/rsync.go:11-13`, `core/rsync.go:29-31`
```go
if src[len(src)-1] != '/' {
    src = src + "/"
}
```
**Fix:** Extract to `ensureTrailingSlash(path string)` helper

### 5. GetNewestMtime Double Dereference
**Location:** `cli/park.go:58`, `cli/rm.go:53`
```go
if newestInfo != nil && *newestInfo != nil {
    mtime := (*newestInfo).ModTime()
```
**Fix:** Fix return type to return `time.Time` directly (eliminates duplication)

---

## 🏛️ SOLID PRINCIPLE VIOLATIONS

### S - Single Responsibility Principle ❌

**`cli/grab.go` has 5 responsibilities:**
1. Load state from disk
2. Validate user input and project existence
3. Discover archive projects (scanning filesystem)
4. Perform file system operations (rsync, mkdir)
5. Update and save state

**`core/state.go` mixes concerns:**
1. Data structures (Project, State structs)
2. File I/O operations (Load, Save)
3. Business logic (GetArchivePath)
4. Default configuration (CreateDefault)

**Recommendation:** Split into:
- `core/models.go` - Data structures only
- `core/persistence.go` - File I/O
- `core/operations.go` - Business logic
- `core/config.go` - Configuration management

### O - Open/Closed Principle ❌

**`main.go` switch statement:**
```go
switch command {
case "init": ...
case "list", "ls": ...
case "grab", "checkout": ...
// Must modify this file for every new command
}
```
**Issue:** Closed for extension - adding new commands requires modifying main.go
**Fix:** Use command registry pattern or cobra's built-in command system

**`core/state.go:149-160` hardcoded mappings:**
```go
switch category {
case "pycharm": return filepath.Join(homeDir, "PycharmProjects")
case "rstudio": return filepath.Join(homeDir, "RStudioProjects")
default: return filepath.Join(homeDir, "code")
}
```
**Issue:** New categories require code changes
**Fix:** Make category→path mapping configurable

### L - Liskov Substitution Principle ⚠️

**Not applicable** - No inheritance or polymorphism used anywhere. This itself is a design smell; polymorphism would enable better testing and extensibility.

### I - Interface Segregation Principle ❌

**No interfaces defined at all.** The codebase uses only concrete types:
- No `StateRepository` interface for persistence
- No `FileSystem` interface for OS operations
- No `SyncEngine` interface for rsync operations
- No `ProjectValidator` interface for validation logic

**Impact:**
- Cannot mock dependencies for unit testing
- Cannot swap implementations (e.g., different sync backends)
- Tight coupling between all components

**Recommendation:** Define focused interfaces:
```go
type StateLoader interface {
    Load() (*State, error)
}

type StateSaver interface {
    Save(*State) error
}

type Syncer interface {
    Sync(src, dst string) error
}

type ProjectValidator interface {
    ValidateGrabbed(projectName string) (*Project, error)
}
```

### D - Dependency Inversion Principle ❌

**High-level modules depend directly on low-level modules:**

```go
// cli/grab.go - High-level policy depends on concrete implementation
func GrabCmd(projectName string) error {
    sm := core.NewStateManager()  // Direct instantiation
    // ...
    core.DiscoverArchiveProjects(state)  // Direct function call
    core.Rsync(archiveProject.Path, localPath)  // Direct dependency
}
```

```go
// core/rsync.go - Depends directly on exec.Command
cmd := exec.Command("rsync", "-av", "--delete", src, dst)
```

**Issues:**
- Cannot inject test doubles
- Cannot replace rsync with alternative sync mechanism
- Cannot test CLI commands without real filesystem
- Cannot test state operations without real disk I/O

**Recommendation:** Invert dependencies:
```go
// Define abstractions
type Syncer interface {
    Sync(src, dst string) error
}

type StateRepository interface {
    Load() (*State, error)
    Save(*State) error
}

// High-level depends on abstraction
type GrabService struct {
    stateRepo StateRepository
    syncer    Syncer
    fs        FileSystem
}

func (s *GrabService) Grab(projectName string) error {
    state, err := s.stateRepo.Load()
    // ...
}

// Wire up in main.go with concrete implementations
```

---

## Recommended Tools

- **CLI Framework:** github.com/spf13/cobra
- **Logging:** github.com/rs/zerolog
- **Testing:** Built-in testing + github.com/stretchr/testify
- **Validation:** Custom validators or github.com/go-playground/validator
- **Config:** github.com/spf13/viper (pairs with cobra)
- **Mocking:** github.com/stretchr/testify/mock or github.com/golang/mock
