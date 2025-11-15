package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/jamespark/parkr/core"
)

// ParkCmd syncs local changes back to archive
func ParkCmd(projectName string) error {
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

	// For Phase 1, we're in no-hash mode
	project.NoHashMode = true

	if err := sm.Save(state); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	fmt.Printf("Successfully parked '%s'\n", projectName)
	return nil
}
