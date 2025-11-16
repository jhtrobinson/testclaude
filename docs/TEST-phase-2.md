# Phase 2: Safety Verification - Test Script

## Prerequisites
- Phase 1 tests passing
- Go tests passing: `go test ./core/ -v`

## Test Script

```bash
# Setup (assumes Phase 1 passing)
rm -rf ~/.parkr /tmp/parkr-archive ~/code/parkr-test-project
cd code && go build -o parkr
./parkr init /tmp/parkr-archive
mkdir -p /tmp/parkr-test-project
echo "hello" > /tmp/parkr-test-project/README.md
./parkr add /tmp/parkr-test-project

# Test 1: Hash catches silent modification (mtime doesn't)
# Modify file content but reset mtime to fool mtime-based check
echo "modified" > /tmp/parkr-test-project/README.md
touch -t 202001010000 /tmp/parkr-test-project/README.md  # Set old mtime

# With --no-hash: mtime check passes (dangerous!)
PARKR_ALIVE=1 ./parkr rm parkr-test-project --no-hash
# Expected: Mtime verification passes, deletes despite content change

# Restore for hash test
./parkr grab parkr-test-project
echo "modified again" > ~/code/parkr-test-project/README.md
touch -t 202001010000 ~/code/parkr-test-project/README.md  # Frig mtime

# With hash: content check catches it
PARKR_ALIVE=1 ./parkr rm parkr-test-project
# Expected: ERROR - hash mismatch detected, refuses to delete

# Test 2: Workflow with hash enabled (default)
rm -rf ~/.parkr /tmp/parkr-archive ~/code/parkr-test-project
./parkr init /tmp/parkr-archive
mkdir -p /tmp/parkr-test-project
echo "original" > /tmp/parkr-test-project/README.md
./parkr add /tmp/parkr-test-project

# Park computes hashes
./parkr park parkr-test-project
# Expected: Sets archive_content_hash and local_content_hash in state

# Clean remove (hashes match)
PARKR_ALIVE=1 ./parkr rm parkr-test-project
# Expected: Hash verification passes, removes local

# Grab back
./parkr grab parkr-test-project

# Modify without parking
echo "unsaved work" >> ~/code/parkr-test-project/README.md

# Remove blocked (hashes don't match)
PARKR_ALIVE=1 ./parkr rm parkr-test-project
# Expected: ERROR - local hash differs from parked hash

# Test 3: Workflow with --no-hash parked
rm -rf ~/.parkr /tmp/parkr-archive ~/code/parkr-test-project
./parkr init /tmp/parkr-archive
mkdir -p /tmp/parkr-test-project
echo "original" > /tmp/parkr-test-project/README.md
./parkr add /tmp/parkr-test-project

# Park without hashes
./parkr park parkr-test-project --no-hash
# Expected: Syncs but doesn't compute hashes, sets no_hash_mode=true

# rm requires --no-hash flag when parked with --no-hash
PARKR_ALIVE=1 ./parkr rm parkr-test-project
# Expected: ERROR - project was parked with --no-hash, must specify --no-hash or --force

PARKR_ALIVE=1 ./parkr rm parkr-test-project --no-hash
# Expected: Uses mtime verification, removes local

# Test 4: --force bypasses all verification
./parkr grab parkr-test-project
echo "unsaved changes" >> ~/code/parkr-test-project/README.md
PARKR_ALIVE=1 ./parkr rm parkr-test-project --force
# Expected: Skips all verification, deletes anyway with warning
```

## Unit Tests

```bash
# Run hash unit tests
cd code && go test ./core/ -v
```

All 8 tests should pass:
- TestComputeProjectHash_NormalProject
- TestComputeProjectHash_Deterministic
- TestComputeProjectHash_EmptyDirectory
- TestComputeProjectHash_SkipsSymlinks
- TestComputeProjectHash_DifferentContent
- TestComputeProjectHash_DifferentFilenames
- TestComputeProjectHash_NestedDirectories
- TestComputeProjectHash_UnicodeFilenames

## Success Criteria

- [x] Hash computation is deterministic
- [x] Symlinks are skipped (not followed)
- [x] Empty directories return error
- [x] Hash catches content changes even with frigged mtime
- [x] --no-hash falls back to mtime verification
- [x] --force bypasses all verification
- [x] Hash mismatch provides helpful error messages
