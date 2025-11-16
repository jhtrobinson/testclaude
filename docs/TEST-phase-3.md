# Phase 3: Status and Information - Test Script

## Prerequisites
- Phase 1 and Phase 2 tests passing
- Multiple projects in various states for testing

## Setup

```bash
# Clean slate
rm -rf ~/.parkr /tmp/parkr-archive ~/code/parkr-test-*

# Build
cd code && go build -o parkr

# Initialize
./parkr init /tmp/parkr-archive

# Create and add multiple test projects
mkdir -p /tmp/parkr-test-active
echo "active project" > /tmp/parkr-test-active/README.md
./parkr add /tmp/parkr-test-active

mkdir -p /tmp/parkr-test-archived
echo "archived project" > /tmp/parkr-test-archived/README.md
./parkr add /tmp/parkr-test-archived
./parkr park parkr-test-archived
PARKR_ALIVE=1 ./parkr rm parkr-test-archived --no-hash

mkdir -p /tmp/parkr-test-dirty
echo "dirty project" > /tmp/parkr-test-dirty/README.md
./parkr add /tmp/parkr-test-dirty
./parkr park parkr-test-dirty
echo "unsaved changes" >> /tmp/parkr-test-dirty/README.md
```

## Test: parkr status

```bash
./parkr status
```

**Expected Output:**
- Shows all grabbed projects
- Displays sync status (clean, dirty, unparked)
- Shows last park time
- Indicates disk usage

**Success Criteria:**
- [ ] Lists all grabbed projects
- [ ] parkr-test-active shown as "grabbed" or "unparked"
- [ ] parkr-test-archived NOT shown (not grabbed)
- [ ] parkr-test-dirty shown as "dirty" or "modified"
- [ ] Displays timestamps in readable format
- [ ] Shows project size

## Test: parkr info <project>

```bash
./parkr info parkr-test-active
./parkr info parkr-test-dirty
./parkr info parkr-test-archived
```

**Expected Output:**
- Project name and paths (local/archive)
- Category and master
- Grabbed status and timestamps
- Hash information (if computed)
- File count and size statistics

**Success Criteria:**
- [ ] Shows local path for grabbed projects
- [ ] Shows archive path for all projects
- [ ] Displays grabbed_at timestamp
- [ ] Displays last_park_at timestamp (if parked)
- [ ] Shows hash mode (hash vs no-hash)
- [ ] Shows content hashes (if computed)

## Test: parkr local

```bash
./parkr local
```

**Expected Output:**
- Lists all directories in configured local roots (~/code/, ~/PycharmProjects/, ~/RStudioProjects/)
- Marks which are managed vs unmanaged
- Shows project sizes

**Success Criteria:**
- [ ] Scans configured local directories
- [ ] Shows parkr-test-active as "managed"
- [ ] Shows parkr-test-dirty as "managed"
- [ ] Shows other directories as "unmanaged"
- [ ] Displays directory sizes

## Test: parkr verify

```bash
./parkr verify
```

**Expected Output:**
- Checks state file consistency
- Verifies all archive paths exist
- Verifies grabbed projects have valid local paths
- Reports any orphaned or missing projects

**Success Criteria:**
- [ ] No errors for consistent state
- [ ] Reports missing archive paths
- [ ] Reports missing local paths for grabbed projects
- [ ] Identifies state/filesystem mismatches

```bash
# Test with inconsistent state
mkdir -p ~/code/parkr-orphan
./parkr verify
# Expected: Reports untracked project in local directory
```

## Test: parkr config

```bash
./parkr config
```

**Expected Output:**
- Current archive root
- Configured masters and categories
- Default master
- State file location

**Success Criteria:**
- [ ] Shows archive root path
- [ ] Lists all categories (code, pycharm, rstudio, misc)
- [ ] Shows default master
- [ ] Shows state file path (~/.parkr/state.json)

## Edge Cases

```bash
# Test with no grabbed projects
rm -rf ~/.parkr /tmp/parkr-archive ~/code/parkr-test-*
./parkr init /tmp/parkr-archive
./parkr status
# Expected: "No projects currently grabbed" or empty list

# Test info on non-existent project
./parkr info nonexistent
# Expected: Error message

# Test local with empty directories
./parkr local
# Expected: Handles missing directories gracefully
```

## Success Criteria Summary

- [ ] status command shows grabbed projects with sync status
- [ ] info command provides detailed project information
- [ ] local command discovers managed and unmanaged projects
- [ ] verify command checks state consistency
- [ ] config command displays current configuration
- [ ] All commands handle edge cases gracefully
- [ ] Output is formatted and readable
