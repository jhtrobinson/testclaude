# Parkr Implementation TODO

## Phase 1: Minimal Core (MVP)
Get basic project tracking working with essential commands.

### Core Infrastructure
- [ ] Set up Go module structure (`core/`, `cli/`)
- [ ] Define state file JSON structures
- [ ] Implement state file read/write (`~/.parkr/state.json`)
- [ ] Basic error handling and exit codes

### Commands
- [ ] `parkr init` - Create state file with default config
- [ ] `parkr add <local-path>` - Add local project to archive
- [ ] `parkr list` - Show all projects in archive
- [ ] `parkr park <project>` - Sync local changes back to archive
- [ ] `parkr rm <project>` - Delete local copy (with basic mtime verification)
- [ ] `parkr grab <project>` - Copy project from archive to local (for later retrieval)

### Testing Phase 1
```bash
# Clean up any previous state
rm -rf ~/.parkr /tmp/parkr-archive /tmp/parkr-test-project

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
./parkr park parkr-test-project
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

---

## Phase 2: Safety Verification
Add hash-based verification for safe deletion.

- [ ] Implement SHA256 content hashing (sorted file walk)
- [ ] Track `archive_content_hash`, `local_content_hash`
- [ ] Track `local_hash_computed_at`, `last_park_mtime`
- [ ] Implement `no_hash_mode` flag logic
- [ ] `parkr park --no-hash` option
- [ ] `parkr rm` with hash verification (default)
- [ ] `parkr rm --force` option
- [ ] Dirty detection (mtime vs hash computed time)

### Testing Phase 2
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

---

## Phase 3: Status and Information
Add visibility into project state.

- [ ] `parkr status` - Show all grabbed projects with sync status
- [ ] `parkr info <project>` - Detailed project information
- [ ] `parkr local` - Show all local projects (managed and unmanaged)
- [ ] `parkr verify` - Check state file consistency
- [ ] `parkr config` - Show current configuration

---

## Phase 4: Project Management
Add/remove projects from system.

- [ ] `parkr add <local-path>` - Add local project to archive
- [ ] Project type auto-detection (Python, R, Node)
- [ ] `parkr add --move` option
- [ ] `parkr checkout --force` option
- [ ] `parkr checkout --to <path>` option
- [ ] `parkr remove <project>` - Remove from archive
- [ ] `parkr remove --everywhere` option

---

## Phase 5: Space Management
Automated cleanup features.

- [ ] `parkr report` - Disk usage analysis
- [ ] `parkr prune <size>` - Dry-run space freeing
- [ ] `parkr prune --exec` - Actually delete
- [ ] `parkr prune --interactive` - Interactive selection
- [ ] Size parsing (10G, 500M, 2T)
- [ ] Sort by modified/size/name

---

## Phase 6: Advanced Features
Polish and additional functionality.

- [ ] `parkr sync <project>` - Backup without delete flag
- [ ] `parkr hash-update <project>` - Recompute hashes
- [ ] Multiple masters support
- [ ] `--all` flags for batch operations
- [ ] JSON output format (`--json`)
- [ ] `parkr help [command]`
- [ ] Progress indicators for rsync
- [ ] Better error messages

---

## Architecture Notes
- `core/` - Business logic (state, rsync, hashing)
- `cli/` - Command-line interface (cobra/flag parsing)
- State file: `~/.parkr/state.json`
- Uses rsync with `-av --delete` flags
