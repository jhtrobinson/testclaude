# Phase 1: Minimal Core (MVP) - Test Script

## Prerequisites
- Go installed (1.23+)
- rsync available

## Test Script

```bash
# Clean up any previous state
rm -rf ~/.parkr /tmp/parkr-archive /tmp/parkr-test-project ~/code/parkr-test-project

# Build the binary
cd code && go build -o parkr

# 1. Initialize
./parkr init /tmp/parkr-archive
# Expected: Creates ~/.parkr/state.json with archive root at /tmp/parkr-archive

# 2. Create a local project
mkdir -p /tmp/parkr-test-project
echo "hello" > /tmp/parkr-test-project/README.md

# 3. Add project to archive
./parkr add /tmp/parkr-test-project
# Expected: Copies to /tmp/parkr-archive/code/parkr-test-project
# Expected: Auto-detects category as "code"
# Expected: Shows in state as grabbed

# 4. List projects
./parkr list
# Expected: Shows parkr-test-project as "grabbed"

# 5. Modify and park
echo "world" >> /tmp/parkr-test-project/README.md
./parkr park parkr-test-project --no-hash
# Expected: Syncs changes to archive (archive now has "hello\nworld")

# 6. Remove local copy (with verification)
PARKR_ALIVE=1 ./parkr rm parkr-test-project --no-hash
# Expected: Mtime verification passes, deletes /tmp/parkr-test-project

# 7. Verify it's gone
ls /tmp/parkr-test-project
# Expected: No such file or directory

# 8. Grab it back
./parkr grab parkr-test-project
# Expected: Restores from archive to ~/code/parkr-test-project

# 9. Verify content preserved
cat ~/code/parkr-test-project/README.md
# Expected: "hello\nworld"
```

## Success Criteria

- [x] State file created at ~/.parkr/state.json
- [x] Projects copied to archive with rsync
- [x] Category auto-detection works
- [x] List command shows project status
- [x] Park syncs local changes to archive
- [x] Rm deletes local copy after verification
- [x] Grab restores from archive
- [x] Content integrity preserved through full cycle
