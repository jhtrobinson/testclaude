package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/jamespark/parkr/core"
)

// ParkCmd syncs local changes back to archive
func ParkCmd(projectName string, noHash bool) error {
	sm := core.NewStateManager()
	state, err := sm.Load()
	if err != nil {
		return err
	}

	// Check if project is grabbed
	project, exists := state.Projects[projectName]
	if !exists || !project.IsGrabbed {
		return fmt.Errorf("project '%s' is not currently grabbed", projectName)
	}

	// Verify local path exists
	if _, err := os.Stat(project.LocalPath); os.IsNotExist(err) {
		return fmt.Errorf("local path does not exist: %s", project.LocalPath)
	}

	// Get archive path
	archivePath, err := state.GetArchivePath(projectName)
	if err != nil {
		return err
	}

	// Verify archive path exists
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		return fmt.Errorf("archive path does not exist: %s", archivePath)
	}

	fmt.Printf("Parking %s from %s to %s...\n", projectName, project.LocalPath, archivePath)

	// Compute local hash before sync (if not in no-hash mode)
	var localHashBefore string
	if !noHash {
		fmt.Println("Computing local content hash...")
		localHashBefore, err = core.ComputeProjectHash(project.LocalPath)
		if err != nil {
			return fmt.Errorf("failed to compute local hash: %w", err)
		}
	}

	// Rsync from local to archive
	if err := core.Rsync(project.LocalPath, archivePath); err != nil {
		return fmt.Errorf("failed to sync project: %w", err)
	}

	// Get newest mtime from local
	newestInfo, err := core.GetNewestMtime(project.LocalPath)
	if err != nil {
		return fmt.Errorf("failed to get mtime: %w", err)
	}

	// Update state
	now := time.Now()
	project.LastParkAt = &now
	project.NoHashMode = noHash

	if newestInfo != nil && *newestInfo != nil {
		mtime := (*newestInfo).ModTime()
		project.LastParkMtime = &mtime
	}

	// Compute and verify hashes if not in no-hash mode
	if !noHash {
		fmt.Println("Verifying archive content hash...")
		archiveHash, err := core.ComputeProjectHash(archivePath)
		if err != nil {
			return fmt.Errorf("failed to compute archive hash: %w", err)
		}

		if localHashBefore != archiveHash {
			return fmt.Errorf("hash mismatch after sync:\n"+
				"  Local hash:   %s\n"+
				"  Archive hash: %s\n"+
				"Possible causes:\n"+
				"  - Files were modified during rsync operation\n"+
				"  - Rsync failed to copy some files (check permissions)\n"+
				"  - Disk I/O errors occurred during sync\n"+
				"  - Symlinks or special files handled differently",
				localHashBefore, archiveHash)
		}

		// Store hashes
		project.LocalContentHash = &localHashBefore
		project.ArchiveContentHash = &archiveHash
		hashTime := time.Now()
		project.LocalHashComputedAt = &hashTime

		fmt.Println("Hash verification passed.")
	}

	if err := sm.Save(state); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	fmt.Printf("Successfully parked '%s'\n", projectName)
	return nil
}
