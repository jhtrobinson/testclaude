# Phase 5: Space Management - Test Script

## Prerequisites
- Phase 1-4 tests passing
- At least 1GB free disk space for testing

## Setup

```bash
# Clean slate
rm -rf ~/.parkr /tmp/parkr-archive ~/code/parkr-test-* ~/PycharmProjects/parkr-test-* ~/RStudioProjects/parkr-test-*

# Build
cd code && go build -o parkr

# Initialize
./parkr init /tmp/parkr-archive

# Create test projects with varying sizes and ages
# Small project (old)
mkdir -p ~/code/parkr-test-old-small
dd if=/dev/zero of=~/code/parkr-test-old-small/data.bin bs=1M count=10
echo "old small" > ~/code/parkr-test-old-small/README.md
./parkr add ~/code/parkr-test-old-small
./parkr park parkr-test-old-small
touch -d "3 weeks ago" ~/code/parkr-test-old-small/README.md

# Medium project (recent)
mkdir -p ~/code/parkr-test-recent-medium
dd if=/dev/zero of=~/code/parkr-test-recent-medium/data.bin bs=1M count=50
echo "recent medium" > ~/code/parkr-test-recent-medium/README.md
./parkr add ~/code/parkr-test-recent-medium
./parkr park parkr-test-recent-medium

# Large project (never parked - dirty)
mkdir -p ~/code/parkr-test-dirty
dd if=/dev/zero of=~/code/parkr-test-dirty/data.bin bs=1M count=100
echo "dirty" > ~/code/parkr-test-dirty/README.md
./parkr add ~/code/parkr-test-dirty
# Intentionally NOT parking this one

# Project with no_hash_mode
mkdir -p ~/code/parkr-test-nohash
dd if=/dev/zero of=~/code/parkr-test-nohash/data.bin bs=1M count=25
echo "no hash mode" > ~/code/parkr-test-nohash/README.md
./parkr add ~/code/parkr-test-nohash
./parkr park parkr-test-nohash --no-hash
touch -d "1 week ago" ~/code/parkr-test-nohash/README.md
```

---

## Size Parsing Tests

```bash
# Test parsing various size formats
# (These would be unit tests, but also test via CLI)

# Valid formats
./parkr prune 10G
./parkr prune 500M
./parkr prune 2T
./parkr prune 1.5GB
./parkr prune 100MB
./parkr prune 1024K
./parkr prune 1g  # lowercase

# Invalid formats
./parkr prune abc
# Expected: ERROR - invalid size format

./parkr prune -10G
# Expected: ERROR - size must be positive

./parkr prune 10
# Expected: ERROR - missing unit (or default to bytes?)

./parkr prune 10X
# Expected: ERROR - unknown unit
```

**Success Criteria:**
- [ ] Parses G, GB, M, MB, K, KB, T, TB
- [ ] Case insensitive
- [ ] Supports decimal values (1.5G)
- [ ] Rejects invalid formats
- [ ] Rejects negative values

---

## Test: parkr report (Basic)

```bash
./parkr report
# Expected output like:
# LOCAL DISK USAGE: X GB / Y GB (Z%)
#
# CHECKED OUT PROJECTS:
# PROJECT                  LOCAL SIZE    LAST MODIFIED    LAST CHECKIN    STATUS
# parkr-test-old-small     10 MB         3 weeks ago      3 weeks ago     ✓ Safe to delete
# parkr-test-nohash        25 MB         1 week ago       1 week ago      ✓ Safe to delete
# parkr-test-recent-medium 50 MB         just now         just now        ✓ Safe to delete
# parkr-test-dirty         100 MB        just now         never           ✗ Never checked in
#
# PRUNING CANDIDATES (safe to delete, oldest first):
# 1. parkr-test-old-small (10 MB)
# 2. parkr-test-nohash (25 MB)
# 3. parkr-test-recent-medium (50 MB)
#
# TOTAL RECOVERABLE: 85 MB
```

**Success Criteria:**
- [ ] Shows all grabbed projects
- [ ] Computes sizes correctly
- [ ] Shows last modified time
- [ ] Shows last park time (or "never")
- [ ] Identifies safe vs unsafe projects
- [ ] Lists candidates oldest first
- [ ] Calculates total recoverable space

---

## Test: parkr report --candidates

```bash
./parkr report --candidates
# Expected: Only shows safe-to-delete projects
# Should NOT show parkr-test-dirty
```

**Success Criteria:**
- [ ] Filters to only pruning candidates
- [ ] Excludes dirty/never-parked projects

---

## Test: parkr report --sort

```bash
./parkr report --sort size
# Expected: Sorted by size (largest first or smallest first?)

./parkr report --sort name
# Expected: Alphabetical order

./parkr report --sort modified
# Expected: By last modified time (default behavior)
```

**Success Criteria:**
- [ ] --sort size orders by project size
- [ ] --sort name orders alphabetically
- [ ] --sort modified orders by modification time
- [ ] Default is modified

---

## Test: parkr report --json

```bash
./parkr report --json | jq .
# Expected: Valid JSON output with same information
```

**Success Criteria:**
- [ ] Outputs valid JSON
- [ ] Contains all project info
- [ ] Machine-parseable

---

## Test: parkr report --recompute-hashes

```bash
# Modify a file without updating mtime
echo "sneaky change" >> ~/code/parkr-test-recent-medium/README.md
touch -r ~/code/parkr-test-recent-medium/data.bin ~/code/parkr-test-recent-medium/README.md

./parkr report
# Expected: parkr-test-recent-medium shows as "Safe to delete" (mtime-based)

./parkr report --recompute-hashes
# Expected: parkr-test-recent-medium shows as "Uncommitted work" (hash mismatch)
```

**Success Criteria:**
- [ ] Recomputes local hashes
- [ ] Detects hash mismatches
- [ ] More thorough than mtime-based check
- [ ] Slower but safer

---

## Test: parkr prune (Dry-run)

```bash
./parkr prune 50M
# Expected: DRY-RUN output
# Would delete:
# 1. parkr-test-old-small (10 MB)
# 2. parkr-test-nohash (25 MB)
# 3. parkr-test-recent-medium (50 MB) [partial: only need 15 MB more]
#
# Total to free: 50 MB (target: 50 MB)
# Run with --exec to actually delete

ls ~/code/parkr-test-old-small/
# Expected: Still exists (dry-run only)
```

**Success Criteria:**
- [ ] Default is dry-run (no deletion)
- [ ] Shows what WOULD be deleted
- [ ] Selects oldest first
- [ ] Stops when target size reached
- [ ] Provides clear messaging

---

## Test: parkr prune --exec

```bash
./parkr prune 40M --exec
# Expected: Actually deletes projects
# Deleting parkr-test-old-small... ✓
# Deleting parkr-test-nohash... ✓
# Freed 35 MB

ls ~/code/parkr-test-old-small/
# Expected: No such file or directory

ls ~/code/parkr-test-nohash/
# Expected: No such file or directory

./parkr status
# Expected: Both projects show as "archived" (not grabbed)
```

**Success Criteria:**
- [ ] Actually deletes local copies
- [ ] Verifies before deletion (hash or mtime)
- [ ] Updates state file
- [ ] Reports progress
- [ ] Reports total freed

---

## Test: parkr prune with no_hash_mode project

```bash
# Recreate nohash project
mkdir -p ~/code/parkr-test-nohash2
dd if=/dev/zero of=~/code/parkr-test-nohash2/data.bin bs=1M count=25
./parkr add ~/code/parkr-test-nohash2
./parkr park parkr-test-nohash2 --no-hash

./parkr prune 30M --exec
# Expected: ERROR or skip for no_hash_mode projects
# Should require --no-hash or --force flag

./parkr prune 30M --exec --no-hash
# Expected: Uses mtime verification, proceeds with deletion
```

**Success Criteria:**
- [ ] Respects no_hash_mode setting
- [ ] Requires explicit --no-hash for no_hash_mode projects
- [ ] --no-hash uses mtime verification

---

## Test: parkr prune --force

```bash
# Setup: Create dirty project
mkdir -p ~/code/parkr-test-forceme
dd if=/dev/zero of=~/code/parkr-test-forceme/data.bin bs=1M count=30
./parkr add ~/code/parkr-test-forceme
./parkr park parkr-test-forceme
echo "local changes" >> ~/code/parkr-test-forceme/README.md

./parkr prune 30M --exec
# Expected: Skips parkr-test-forceme (dirty/modified)

./parkr prune 30M --exec --force
# Expected: WARNING about skipping verification
# Deletes anyway (dangerous!)

ls ~/code/parkr-test-forceme/
# Expected: No such file or directory
```

**Success Criteria:**
- [ ] --force skips all verification
- [ ] Shows warning about danger
- [ ] Deletes even dirty projects

---

## Test: parkr prune --interactive

```bash
./parkr prune 100M --exec --interactive
# Expected: Interactive selection UI
# Need to free up 100 MB. Candidates (oldest first):
#
# 1. [ ] parkr-test-old-small (10 MB)
# 2. [ ] parkr-test-nohash (25 MB)
# 3. [ ] parkr-test-recent-medium (50 MB)
#
# Selected: 0 MB / 100 MB target
#
# Controls: [space] toggle, [a] all, [enter] confirm, [q] quit

# User presses space on item 1 and 3
# 1. [x] parkr-test-old-small (10 MB)
# 2. [ ] parkr-test-nohash (25 MB)
# 3. [x] parkr-test-recent-medium (50 MB)
#
# Selected: 60 MB / 100 MB target

# User presses enter to confirm
# Expected: Deletes selected projects
```

**Success Criteria:**
- [ ] Shows interactive selection UI
- [ ] Space toggles selection
- [ ] 'a' selects all
- [ ] Enter confirms
- [ ] 'q' quits without changes
- [ ] Shows running total
- [ ] Only deletes selected projects

---

## Edge Cases

### Prune with insufficient candidates

```bash
./parkr prune 1T
# Expected: WARNING - Only X MB available for pruning
# Cannot reach target of 1 TB
# Shows what CAN be pruned
```

**Success Criteria:**
- [ ] Warns when target exceeds available space
- [ ] Shows maximum recoverable
- [ ] Does not error out

---

### Prune with no grabbed projects

```bash
# Remove all local copies first
PARKR_ALIVE=1 ./parkr rm parkr-test-old-small --no-hash
PARKR_ALIVE=1 ./parkr rm parkr-test-recent-medium --no-hash

./parkr prune 10M
# Expected: No projects to prune
```

**Success Criteria:**
- [ ] Handles empty candidate list gracefully

---

### Prune with all dirty projects

```bash
# Setup: Only dirty projects grabbed
mkdir -p ~/code/parkr-test-dirty1
echo "dirty" > ~/code/parkr-test-dirty1/README.md
./parkr add ~/code/parkr-test-dirty1
# Don't park

./parkr prune 10M
# Expected: No safe candidates available
# All grabbed projects have uncommitted changes
```

**Success Criteria:**
- [ ] Reports no safe candidates
- [ ] Does not delete dirty projects without --force

---

### Report with no projects

```bash
# Clean slate
rm -rf ~/.parkr
./parkr init /tmp/parkr-archive

./parkr report
# Expected: No projects tracked
```

**Success Criteria:**
- [ ] Handles empty state gracefully

---

### Concurrent modification during prune

```bash
# Simulate: File modified after selection but before deletion
# (This is a race condition edge case)
# Expected: Re-verify before each deletion if possible
```

---

### Size calculation edge cases

```bash
# Very large file
dd if=/dev/zero of=~/code/parkr-test-huge/big.bin bs=1G count=1
# Expected: Handles GB+ files correctly

# Empty directory
mkdir -p ~/code/parkr-test-empty
./parkr add ~/code/parkr-test-empty
./parkr report
# Expected: Shows 0 bytes or minimal size

# Many small files
mkdir -p ~/code/parkr-test-many
for i in {1..1000}; do echo "file $i" > ~/code/parkr-test-many/file$i.txt; done
# Expected: Aggregates correctly
```

---

### Invalid prune size

```bash
./parkr prune 0M
# Expected: ERROR - size must be positive

./parkr prune 0.0G
# Expected: ERROR - size must be positive
```

---

### Prune target exactly met

```bash
# Setup exactly 50MB of safe projects
./parkr prune 50M --exec
# Expected: Deletes exactly enough to meet target
# Does not over-delete
```

---

## Success Criteria Summary

### Size Parsing
- [ ] Parses all standard size formats (G, GB, M, MB, K, KB, T, TB)
- [ ] Case insensitive
- [ ] Supports decimal values
- [ ] Proper error handling for invalid input

### Report Command
- [ ] Shows all grabbed projects with sizes
- [ ] Computes disk usage statistics
- [ ] Identifies safe vs unsafe deletion candidates
- [ ] --candidates filters to safe projects only
- [ ] --sort supports size/name/modified
- [ ] --json outputs valid JSON
- [ ] --recompute-hashes provides thorough safety check

### Prune Command
- [ ] Default is dry-run (safe)
- [ ] --exec actually deletes
- [ ] Selects oldest candidates first
- [ ] Stops when target size reached
- [ ] Respects no_hash_mode settings
- [ ] --no-hash uses mtime verification
- [ ] --force skips verification (with warning)
- [ ] --interactive provides selection UI
- [ ] Handles edge cases gracefully (insufficient space, no candidates, etc.)

### Overall
- [ ] All commands handle empty state
- [ ] Proper error messages for invalid input
- [ ] State file updated correctly after operations
- [ ] No data loss for unverified projects (unless --force)
