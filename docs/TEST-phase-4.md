# Phase 4: Project Management - Test Script

## Prerequisites
- Phase 1-3 tests passing

## Setup

```bash
# Clean slate
rm -rf ~/.parkr /tmp/parkr-archive ~/code/parkr-test-* ~/PycharmProjects/parkr-test-* ~/RStudioProjects/parkr-test-*

# Build
cd code && go build -o parkr

# Initialize
./parkr init /tmp/parkr-archive
```

## Test: parkr add with explicit category

```bash
mkdir -p /tmp/parkr-test-explicit
echo "generic project" > /tmp/parkr-test-explicit/README.md

./parkr add /tmp/parkr-test-explicit misc
# Expected: Uses "misc" category
# Expected: Archives to /tmp/parkr-archive/misc/parkr-test-explicit
```

**Success Criteria:**
- [ ] Explicit category override works
- [ ] Archives to correct category path

## Test: parkr add --move

```bash
mkdir -p /tmp/parkr-test-move
echo "will be moved" > /tmp/parkr-test-move/README.md
ls /tmp/parkr-test-move/
# Expected: Directory exists

./parkr add /tmp/parkr-test-move --move
# Expected: Copies to archive AND removes original location

ls /tmp/parkr-test-move/
# Expected: No such file or directory

ls /tmp/parkr-archive/code/parkr-test-move/
# Expected: README.md exists

./parkr list
# Expected: parkr-test-move shown as "archived" (not grabbed since original was moved)
```

**Success Criteria:**
- [ ] --move copies to archive
- [ ] --move removes original directory
- [ ] Project is tracked but not marked as grabbed
- [ ] Content preserved in archive

## Test: parkr checkout --force

```bash
# Setup: add and grab a project
mkdir -p /tmp/parkr-test-force
echo "original" > /tmp/parkr-test-force/README.md
./parkr add /tmp/parkr-test-force

./parkr grab parkr-test-force
ls ~/code/parkr-test-force/
# Expected: README.md exists

# Modify local version
echo "local changes" >> ~/code/parkr-test-force/README.md

# Try to grab again (should fail without --force)
./parkr grab parkr-test-force
# Expected: ERROR - already grabbed or local path exists

# Force overwrite
./parkr checkout parkr-test-force --force
# Expected: Overwrites local with archive version

cat ~/code/parkr-test-force/README.md
# Expected: Original content only (local changes overwritten)
```

**Success Criteria:**
- [ ] Regular grab fails if already grabbed
- [ ] --force overwrites local copy
- [ ] Archive content replaces local content
- [ ] State updated correctly

## Test: parkr checkout --to <path>

```bash
mkdir -p /tmp/parkr-test-custom
echo "custom path test" > /tmp/parkr-test-custom/README.md
./parkr add /tmp/parkr-test-custom
PARKR_ALIVE=1 ./parkr rm parkr-test-custom --no-hash

./parkr checkout parkr-test-custom --to /tmp/custom-location
# Expected: Copies from archive to /tmp/custom-location/parkr-test-custom

ls /tmp/custom-location/parkr-test-custom/
# Expected: README.md exists

./parkr info parkr-test-custom
# Expected: Local path shows /tmp/custom-location/parkr-test-custom
```

**Success Criteria:**
- [ ] Copies to custom location instead of default
- [ ] Updates state with custom local path
- [ ] Project marked as grabbed

## Test: parkr remove <project>

```bash
# Create a project to remove
mkdir -p /tmp/parkr-test-remove
echo "will be removed" > /tmp/parkr-test-remove/README.md
./parkr add /tmp/parkr-test-remove

./parkr list
# Expected: parkr-test-remove listed

./parkr remove parkr-test-remove
# Expected: Removes from state file only
# Expected: Archive still exists

ls /tmp/parkr-archive/code/parkr-test-remove/
# Expected: README.md still exists (archive preserved)

./parkr list
# Expected: parkr-test-remove no longer listed
```

**Success Criteria:**
- [ ] Removes project from state file
- [ ] Does NOT delete archive contents
- [ ] Does NOT delete local copy (if grabbed)
- [ ] Project no longer appears in list

## Test: parkr remove --archive

```bash
# Setup
mkdir -p /tmp/parkr-test-nuke
echo "nuke me" > /tmp/parkr-test-nuke/README.md
./parkr add /tmp/parkr-test-nuke
./parkr grab parkr-test-nuke

./parkr remove parkr-test-nuke --archive
# Expected: WARNING - this will delete archive copy
# Expected: Prompts for confirmation

# After confirmation:
ls /tmp/parkr-archive/code/parkr-test-nuke/
# Expected: No such file or directory

ls ~/code/parkr-test-nuke/
# Expected: Directory still exists (local copy preserved)

./parkr list
# Expected: parkr-test-nuke not listed
```

**Success Criteria:**
- [ ] Warns about permanent deletion
- [ ] Prompts for confirmation
- [ ] Removes from state file
- [ ] Deletes archive copy
- [ ] Local copy preserved (not deleted)

## Edge Cases

```bash
# Test add on already tracked project
./parkr add /tmp/parkr-test-explicit
# Expected: ERROR - already tracked

# Test add on non-existent path
./parkr add /tmp/nonexistent
# Expected: ERROR - path does not exist

# Test remove on non-existent project
./parkr remove nonexistent
# Expected: ERROR - project not found

# Test checkout --force with uncommitted changes
# (should warn about potential data loss)

# Test add with existing archive path
mkdir -p /tmp/parkr-archive/code/duplicate
mkdir -p /tmp/duplicate
echo "collision" > /tmp/duplicate/README.md
./parkr add /tmp/duplicate
# If project name collision, should error or rename
```

## Success Criteria Summary

- [ ] Explicit category override works
- [ ] --move option removes original after archiving
- [ ] checkout --force overwrites existing local copy
- [ ] checkout --to allows custom local path
- [ ] remove removes from state but preserves files
- [ ] remove --archive deletes archive copy with confirmation
- [ ] All commands handle edge cases gracefully
