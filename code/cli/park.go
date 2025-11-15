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

	if newestInfo != nil && *newestInfo != nil {
		mtime := (*newestInfo).ModTime()
		project.LastParkMtime = &mtime
	}

	if noHash {
		// No hash mode - only track mtime
		project.NoHashMode = true
		fmt.Println("Skipping hash computation (--no-hash)")
	} else {
		// Compute hashes for both local and archive
		fmt.Println("Computing project hash...")

		localHash, err := core.ComputeProjectHash(project.LocalPath)
		if err != nil {
			return fmt.Errorf("failed to compute local hash: %w", err)
		}

		archiveHash, err := core.ComputeProjectHash(archivePath)
		if err != nil {
			return fmt.Errorf("failed to compute archive hash: %w", err)
		}

		// After successful rsync, both should match
		if localHash != archiveHash {
			return fmt.Errorf("hash mismatch after sync - this should not happen")
		}

		project.LocalContentHash = &localHash
		project.ArchiveContentHash = &archiveHash
		project.LocalHashComputedAt = &now
		project.NoHashMode = false

		fmt.Printf("Hash computed: %s\n", localHash[:16]+"...")
	}

	if err := sm.Save(state); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	fmt.Printf("Successfully parked '%s'\n", projectName)
	return nil
}
