# Phase 2 Implementation Deficiencies

This document outlines the known deficiencies and edge cases in the Phase 2 Safety Verification implementation that should be addressed in future updates.

## Critical Issues

### 1. Symlink Handling

**Location:** `code/core/hash.go:22`

**Problem:** The `filepath.Walk` function follows symbolic links, which can cause:
- Infinite loops when circular symlinks exist
- Hashing files outside the project directory
- Inconsistent hashes if symlink targets change
- Security implications (symlinks pointing to sensitive files)

**Current Behavior:**
```go
err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
    // Follows symlinks automatically
})
```

**Recommended Fix:**
- Use `filepath.WalkDir` with explicit symlink detection via `info.Type()&os.ModeSymlink`
- Either skip symlinks entirely or hash the symlink target path (not content)
- Document chosen behavior clearly

### 2. Empty Directory Behavior

**Location:** `code/core/hash.go:18-55`

**Problem:** Empty directories produce a valid SHA256 hash (hash of empty input), meaning:
- Two completely empty projects will have identical hashes
- A project that becomes empty may not be detected as changed
- Could mask data loss scenarios

**Current Behavior:**
```go
var fileHashes [][]byte
// If no files found, returns SHA256 of empty byte slice
projectHasher := sha256.New()
for _, fh := range fileHashes {
    projectHasher.Write(fh)
}
return hex.EncodeToString(projectHasher.Sum(nil)), nil
```

**Recommended Fix:**
- Return an error or warning for empty directories
- Or use a sentinel value to distinguish empty projects
- Or include a marker that indicates "empty project"

---

## Recommended Improvements

### 3. Special File Handling

**Location:** `code/core/hash.go:63-64`

**Problem:** No validation of file type before attempting to read. Could fail or behave unexpectedly with:
- Device files (`/dev/*`)
- Unix sockets
- Named pipes (FIFOs)
- Block devices

**Recommended Fix:**
```go
if !info.Mode().IsRegular() {
    // Skip non-regular files or handle appropriately
    return nil
}
```

### 4. Hash Mismatch Error Message

**Location:** `code/cli/park.go:85-87`

**Problem:** Generic error message doesn't help debugging:
```go
if localHash != archiveHash {
    return fmt.Errorf("hash mismatch after sync - this should not happen")
}
```

**Recommended Fix:**
Provide more context:
- File was modified during rsync
- Rsync failed silently
- Permission issues preventing complete sync
- Disk I/O errors

### 5. Missing Unit Tests

**Problem:** No automated tests for hashing functionality.

**Should Test:**
- Normal project with various file types
- Empty directory behavior
- Symlink handling
- Large files (memory efficiency)
- Permission denied scenarios
- Files with special characters in names
- Unicode filenames
- Very deep directory structures

### 6. File Permission Errors

**Location:** `code/core/hash.go:63`

**Problem:** If a file cannot be opened (permission denied), the entire hash computation fails.

**Current Behavior:** Returns error immediately.

**Consideration:** Should unreadable files be:
- Fatal errors (current behavior)
- Skipped with warning
- Included in hash by path only

### 7. Large File Memory Usage

**Location:** `code/core/hash.go:70`

**Current Implementation:** Uses `io.Copy` which is memory-efficient.

**Note:** This is actually implemented correctly, but should be documented as a design decision.

---

## Documentation Gaps

1. **No documentation of hash algorithm choice** - Why SHA256 vs other algorithms
2. **No documentation of hash format** - Path + content concatenation approach
3. **No versioning of hash scheme** - Future changes could break compatibility
4. **No documentation of performance characteristics** - Time/space complexity

---

## Future Considerations

1. **Parallel Hashing** - For large projects, hash files in parallel
2. **Incremental Hashing** - Only rehash changed files based on mtime
3. **Hash Caching** - Store per-file hashes for faster verification
4. **Gitignore Support** - Should `.git/` or other patterns be excluded from hashing?
5. **Binary File Detection** - Different handling for text vs binary files

---

## Priority

**Must Fix Before Production Use:**
1. Symlink handling (Critical - data safety)
2. Empty directory behavior (Critical - edge case)

**Should Fix Soon:**
3. Special file handling (Medium - robustness)
4. Unit tests (Medium - maintainability)

**Nice to Have:**
5. Better error messages (Low - usability)
6. Documentation (Low - maintainability)
