# Project Manager (parkr) - Interface Specification

## Overview
A CLI tool for managing projects between an external archive disk and local working directories.
Projects can be "checked out" to work on locally, then "checked in" to sync back to archive.
Local copies can be deleted when space is needed, with the tool tracking what's safe to remove.

## Architecture Notes
- Initial implementation: CLI only
- Future: Refactor to support TUI (terminal UI) and web browser interfaces
- Core business logic should be separated from interface code for easy refactoring
- State tracked in JSON file (~/.parkr/state.json)
- Uses rsync for file synchronization

## Configuration

Archive locations are configurable in state file under `archive_locations`.

Default mappings:
- `code` → `/Volumes/Extra/project-archive/code`
- `pycharm` → `/Volumes/Extra/project-archive/pycharm`
- `rstudio` → `/Volumes/Extra/project-archive/rstudio`
- `misc` → `/Volumes/Extra/project-archive/misc`

Local roots for checkout (not in state, hardcoded or in config):
- `~/code/`
- `~/PycharmProjects/`

State file: `~/.parkr/state.json`

Archive locations can be modified directly in state.json or via commands (future).

## State File Format
```json
{
  "masters": {
    "primary": {
      "code": "/Volumes/Extra/project-archive/code",
      "pycharm": "/Volumes/Extra/project-archive/pycharm",
      "rstudio": "/Volumes/Extra/project-archive/rstudio",
      "misc": "/Volumes/Extra/project-archive/misc"
    },
    "work": {
      "code": "/Volumes/WorkDrive/projects/code",
      "pycharm": "/Volumes/WorkDrive/projects/pycharm"
    }
  },
  "default_master": "primary",
  "projects": {
    "project-name": {
      "local_path": "/Users/james/code/project-name",
      "master": "primary",
      "archive_category": "code",
      "grabbed_at": "2025-11-14T10:30:00Z",
      "last_park_at": "2025-11-14T12:45:00Z",
      "archive_content_hash": "sha256:a1b2c3d4e5f6...",
      "local_content_hash": "sha256:a1b2c3d4e5f6...",
      "local_hash_computed_at": "2025-11-14T12:45:00Z",
      "last_park_mtime": "2025-11-14T12:45:00Z",
      "no_hash_mode": false,
      "is_grabbed": true
    }
  }
}
```

**Field Descriptions:**
- `masters`: Named master archive locations, each with category mappings
- `default_master`: Which master to use when not specified
- `master`: Which master this project belongs to
- `archive_content_hash`: Hash of files in archive (null if parked with --no-hash)
- `local_content_hash`: Cached hash of local files (null if parked with --no-hash)
- `local_hash_computed_at`: When local hash was computed (null if no hash)
- `last_park_mtime`: Newest file mtime at time of park (always recorded)
- `no_hash_mode`: true if project was parked with --no-hash
- Archive path derived: `masters[master][archive_category] + "/" + project_name`

**Mode Enforcement:**
- If `no_hash_mode == true`: Can only delete with --no-hash or --force
- If `no_hash_mode == false`: Can delete with hash verification (default) or --no-hash or --force
- Computing hash (via park without --no-hash or hash-update) sets `no_hash_mode = false`

## Commands

### Project Lifecycle

**parkr add <local-path> [category]**
- Adds existing local project to archive
- Auto-detects category if not specified (Python→pycharm, R→rstudio, etc)
- Copies to archive, keeps local copy
- Marks as checked out
- Options:
  - `--move` : Delete local copy after adding to archive
  - `--category <cat>` : Override auto-detection

Example:
```bash
parkr add ~/Desktop/my-project
parkr add ~/code/experiment pycharm
parkr add ~/work/analysis --move
```

**parkr checkout <project>**
- Copies project from archive to appropriate local directory
- Records checkout in state file
- Fails if already checked out locally
- Options:
  - `--force` : Overwrite existing local copy
  - `--to <path>` : Checkout to specific location instead of default

Example:
```bash
parkr checkout ml-pipeline
parkr checkout legacy-app --force
```

**parkr park <project>**
- Syncs local changes to archive (rsync)
- Records current newest mtime as `last_park_mtime`
- **Default behavior (with hashing)**:
  - Computes hash of archive files → `archive_content_hash`
  - Computes hash of local files → `local_content_hash` and `local_hash_computed_at`
  - Sets `no_hash_mode = false`
- **With --no-hash**:
  - Skips hash calculation
  - Sets `no_hash_mode = true`
  - Can only verify later with --no-hash (or --force)
- Updates `last_park_at` timestamp
- Does NOT delete local copy
- Options:
  - `--all` : Park all grabbed projects
  - `--no-hash` : Skip hash calculation, use mtime-only mode

Example:
```bash
parkr park ml-pipeline              # Full hash verification enabled
parkr park big-dataset --no-hash    # Fast, but limited to mtime verification
parkr park --all
```

**parkr sync <project>**
- Like checkin but does NOT mark as safe-to-delete
- Use when you want to backup but continue working
- Options:
  - `--all` : Sync all checked-out projects

Example:
```bash
parkr sync experiment-in-progress
```

**parkr rm <project>**
- Deletes LOCAL copy only (archive remains safe)
- Before deletion, verifies safety based on project's mode:
  - **If no_hash_mode == false (has hashes)**:
    - Default: Recomputes local hash, compares to `archive_content_hash`
    - With --no-hash: Uses mtime verification instead
    - With --force: No verification
  - **If no_hash_mode == true (no hashes)**:
    - Default: ERROR - must use --no-hash or --force
    - With --no-hash: Uses mtime verification (compares to `last_park_mtime`)
    - With --force: No verification
- Removes from local disk
- Updates state to is_grabbed: false
- Options:
  - `--no-hash` : Use mtime verification instead of hash
  - `--force` : Delete without verification (dangerous)

Example:
```bash
# Project parked with hash
parkr rm ml-pipeline              # Hash verification
parkr rm ml-pipeline --no-hash    # Mtime verification

# Project parked with --no-hash
parkr rm big-dataset              # ERROR: must use --no-hash
parkr rm big-dataset --no-hash    # OK: mtime verification
parkr rm big-dataset --force      # OK: no verification (dangerous)
```

**parkr remove <project>**
- Removes project from archive
- Options:
  - `--local` : Remove local copy only (same as rm)
  - `--everywhere` : Remove from both archive and local
  - `--confirm` : Skip confirmation prompt

Example:
```bash
parkr remove old-abandoned-project
parkr remove temp-experiment --everywhere --confirm
```

### Status & Information

**parkr list [category]**
- Lists all projects in archive
- Shows: name, category, size, checked out status
- Optional filter by category
- Options:
  - `--sort <field>` : Sort by name|size|modified (default: name)
  - `--json` : Output as JSON

Example:
```bash
parkr list
parkr list pycharm
parkr list --sort size
```

Output format:
```
PROJECT              CATEGORY    SIZE      CHECKED OUT    LAST MODIFIED
ml-pipeline          code        8.2 GB    Yes            2 hours ago
customer-analysis    pycharm     3.1 GB    Yes            5 days ago
legacy-scraper       code        450 MB    No             3 weeks ago
```

**parkr status**
- Shows all currently checked-out projects
- Indicates which have uncommitted work (modified since last checkin)
- Options:
  - `--json` : Output as JSON

Example:
```bash
parkr status
```

Output format:
```
CHECKED OUT PROJECTS:
PROJECT              LOCAL SIZE    LAST MODIFIED    LAST CHECKIN    STATUS
ml-pipeline          8.2 GB        2 hours ago      2 hours ago     ✓ Safe to delete
customer-analysis    3.1 GB        5 days ago       3 days ago      ⚠ Has uncommitted work
new-experiment       12.5 GB       30 mins ago      never           ✗ Never checked in
```

**parkr local**
- Shows all projects in local directories (~/code, ~/PycharmProjects, etc)
- Includes both managed (tracked) and unmanaged projects
- Useful for finding projects that should be added
- Options:
  - `--unmanaged` : Show only unmanaged projects

Example:
```bash
parkr local
parkr local --unmanaged
```

**parkr info <project>**
- Shows detailed information about a specific project
- Archive path, local path, sizes, timestamps, status

Example:
```bash
parkr info ml-pipeline
```

Output format:
```
Project: ml-pipeline
Archive: /Volumes/Extra/project-archive/code/ml-pipeline (8.2 GB)
Local: /Users/james/code/ml-pipeline (8.2 GB)
Checked out: 2025-11-14 10:30:00
Last checkin: 2025-11-14 12:45:00
Last modified: 2025-11-14 12:45:00
Status: Safe to delete
Archive exists: Yes
Local exists: Yes
```

### Space Management

**parkr report**
- Shows disk usage analysis
- Lists checked-out projects with sizes (computed on-demand) and modification times
- Shows pruning candidates (safe to delete)
- For safety checking: compares `local_content_hash` vs `archive_content_hash` (if hashes exist)
- Falls back to mtime check if no hashes available
- Calculates total recoverable space
- Options:
  - `--candidates` : Show only projects safe to delete
  - `--recompute-hashes` : Recompute local hashes before comparison (slow but thorough)
  - `--json` : Output as JSON
  - `--sort <field>` : Sort by size|modified|name (default: modified)

Example:
```bash
parkr report
parkr report --candidates
parkr report --recompute-hashes  # Thorough check
```

Output format:
```
LOCAL DISK USAGE: 45.2 GB / 250 GB (18%)

CHECKED OUT PROJECTS:
PROJECT              LOCAL SIZE    LAST MODIFIED    LAST CHECKIN    STATUS
ml-pipeline          8.2 GB        2 hours ago      2 hours ago     ✓ Safe to delete
customer-analysis    3.1 GB        5 days ago       3 days ago      ⚠ Uncommitted work
legacy-scraper       450 MB        3 weeks ago      3 weeks ago     ✓ Safe to delete
new-experiment       12.5 GB       30 mins ago      never           ✗ Never checked in

PRUNING CANDIDATES (safe to delete, oldest first):
1. legacy-scraper (450 MB) - last modified 3 weeks ago
2. ml-pipeline (8.2 GB) - last modified 2 hours ago

TOTAL RECOVERABLE: 8.65 GB
```

**parkr prune <size>**
- Free up disk space by removing local copies
- **Default behavior: DRY-RUN** - shows what would be deleted, doesn't delete
- Shows candidates sorted by last modified (oldest first)
- Before deletion (with --exec): Verifies each project based on its mode:
  - **If no_hash_mode == false**: Hash verification (or mtime with --no-hash)
  - **If no_hash_mode == true**: Must use --no-hash (or --force)
- `<size>` can be: 10G, 500M, 2T, etc.
- Options:
  - `--exec` : Actually delete (default is dry-run)
  - `--interactive` : Interactively select which projects to delete
  - `--no-hash` : Use mtime verification for all projects
  - `--force` : Skip verification entirely (dangerous)

Example:
```bash
parkr prune 10G                      # Dry-run: shows what would be deleted
parkr prune 10G --exec               # Deletes with appropriate verification per project
parkr prune 20G --exec --interactive # Pick what to delete
parkr prune 5G --exec --no-hash      # Mtime verification for all
parkr prune 10G --exec --force       # Dangerous: no verification
```

Interactive output:
```
Need to free up 10 GB. Candidates (oldest first):

1. [ ] legacy-scraper (450 MB) - 3 weeks old
2. [ ] old-analysis (2.1 GB) - 5 days old
3. [ ] ml-pipeline (8.2 GB) - 2 hours old

Total if all selected: 10.55 GB

Select projects (space to toggle, a for all, enter to confirm): 
```

### Utility Commands

**parkr init**
- Initialize parkr (create state file, verify archive exists)
- Run once on first use

**parkr config**
- Show current configuration
- Archive root, local roots, state file location

**parkr verify**
- Verify state file consistency
- Check that archived projects exist
- Check that local paths match reality
- Report any inconsistencies

**parkr hash-update <project>**
- Recompute and update `local_content_hash` for a project
- Updates `local_hash_computed_at` timestamp
- Sets `no_hash_mode = false` (enables hash-based verification)
- Useful for:
  - Converting a --no-hash project to full hash verification
  - Refreshing cached hash after making changes
- Options:
  - `--all` : Update hashes for all grabbed projects

Example:
```bash
parkr hash-update ml-pipeline       # Refresh cached hash
parkr hash-update big-dataset       # Enable hash mode for --no-hash project
parkr hash-update --all
```

**parkr help [command]**
- Show help for all commands or specific command

## Error Handling

Clear error messages for common issues:
- Archive disk not mounted
- Project doesn't exist
- Project already checked out
- Not enough space
- Permission issues

Exit codes:
- 0: Success
- 1: General error
- 2: Invalid arguments
- 3: Archive not accessible
- 4: State file error

## Project Type Detection

Auto-detect project type based on files present:
- Python: pyproject.toml, requirements.txt, setup.py → pycharm
- R: .Rproj, DESCRIPTION → rstudio  
- Node: package.json → code
- Default: code

## Safety Verification (At Deletion Time)

Safety is NEVER marked in advance. Verification always happens at deletion time.

### Hash Calculation

```
For each file in project (sorted by path):
  file_hash = sha256(relative_path + file_content)
project_hash = sha256(concatenate all file_hashes)
```

### Project Modes

Projects have a `no_hash_mode` flag that determines verification requirements:

**no_hash_mode == false (has hashes)**
- Project was parked with hash calculation
- Can verify with: hash (default), mtime (--no-hash), or force (--force)

**no_hash_mode == true (no hashes)**
- Project was parked with --no-hash
- Can verify with: mtime (--no-hash) or force (--force) only
- Cannot use hash verification (hashes don't exist)
- Use `parkr hash-update` to compute hashes and set no_hash_mode = false

### Verification Methods

**1. Hash Verification (default for projects with hashes)**
```
If no_hash_mode == true: ERROR - cannot use hash verification
If no_hash_mode == false:
  1. Check if dirty: current_newest_mtime > local_hash_computed_at
  2. If dirty: UNSAFE - refuse deletion
  3. Recompute current local hash
  4. Compare to archive_content_hash
  5. If hashes match: SAFE to delete
  6. If mismatch: UNSAFE - refuse deletion
```

**2. Timestamp Verification (--no-hash, works for all projects)**
```
1. Get current newest mtime from local project
2. Compare to last_park_mtime
3. If current_mtime <= last_park_mtime: SAFE to delete
4. If current_mtime > last_park_mtime: UNSAFE - refuse deletion
```

**3. Force Mode (--force, dangerous)**
```
Delete without any verification (works for all projects)
```

### Commands That Delete

**parkr rm <project>**
- Respects no_hash_mode flag
- If no_hash_mode == true: requires --no-hash or --force
- If no_hash_mode == false: can use hash (default), --no-hash, or --force

**parkr prune <size>**
- Default (dry-run): Shows candidates, no actual deletion
- With --exec: Deletes projects, respecting each project's no_hash_mode
- With --exec --no-hash: Uses mtime verification for all projects
- With --exec --force: No verification for any projects

### Dirty Detection (Fast Pre-check for Hash Verification)

Before expensive hash computation:
```
current_newest_mtime = GetNewestMtime(local_path)
if current_newest_mtime > local_hash_computed_at:
    return DIRTY (refuse deletion, don't bother computing hash)
```

This catches obviously modified projects without hash computation.

## Size Parsing

Accept size arguments in various formats:
- 10G, 10GB, 10g (gigabytes)
- 500M, 500MB (megabytes)
- 2T, 2TB (terabytes)

## Future UI Considerations

Core logic should be separated to allow:

1. **TUI (terminal UI)** using libraries like bubbletea (Go) or textual (Python)
   - Interactive file browser view
   - Real-time sync progress
   - Visual project status dashboard

2. **Web UI**
   - Browse archive via web browser
   - Drag-and-drop project management
   - Charts/graphs of disk usage
   - Multi-machine support (future)

Suggested architecture:
- `core/` : Business logic (checkout, checkin, state management, rsync)
- `cli/`  : Command-line interface 
- `tui/`  : Terminal UI (future)
- `web/`  : Web interface (future)

## Implementation Notes

- Use rsync with flags: `-av --delete` (archive mode, verbose, delete extraneous)
- Check disk space before checkout
- Atomic operations where possible (temp files, then rename)
- Handle interrupted operations gracefully
- Respect .gitignore and similar files (consider adding .parkr)
- Progress indication for long operations (rsync progress)
- Consider adding hooks (pre-checkin, post-checkout scripts)
- **Hashing**: Use SHA256 for content hashes. Walk directory tree in sorted order for consistency
- **Performance**: Hash calculation can be slow for large projects - provide `--no-hash` option
- **Concurrency**: Consider parallel hashing for multi-core systems (future optimization)

## Open Questions / Future Features

1. Compression for archived projects?
2. Multiple archive locations (work drive, home drive)?
3. Project tags/labels for organization?
4. Search functionality?
5. Hooks (pre-checkin, post-checkout scripts)?
