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
- [ ] `parkr grab <project>` - Copy project from archive to local (alias for checkout)
- [ ] `parkr park <project>` - Sync local changes back to archive
- [ ] `parkr rm <project>` - Delete local copy (with basic mtime verification)
- [ ] `parkr list` - Show all projects in archive

### Testing Phase 1
```bash
# 1. Initialize
parkr init

# 2. Manually place a test project in archive
mkdir -p /Volumes/Extra/project-archive/code/test-project
echo "hello" > /Volumes/Extra/project-archive/code/test-project/README.md

# 3. List projects
parkr list
# Should show: test-project

# 4. Grab project
parkr grab test-project
# Should copy to ~/code/test-project

# 5. Verify grabbed
parkr list
# Should show test-project as "grabbed"

# 6. Modify and park
echo "world" >> ~/code/test-project/README.md
parkr park test-project
# Should sync to archive

# 7. Remove local
parkr rm test-project --no-hash
# Should delete ~/code/test-project
```

---

## Phase 2: Safety Verification
Add hash-based verification for safe deletion.

- [x] Implement SHA256 content hashing (sorted file walk)
- [x] Track `archive_content_hash`, `local_content_hash`
- [x] Track `local_hash_computed_at`, `last_park_mtime`
- [x] Implement `no_hash_mode` flag logic
- [x] `parkr park --no-hash` option
- [x] `parkr rm` with hash verification (default)
- [x] `parkr rm --force` option
- [x] Dirty detection (mtime vs hash computed time)

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
