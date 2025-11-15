package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jamespark/parkr/core"
)

// GrabCmd checks out a project from archive to local
func GrabCmd(projectName string) error {
	sm := core.NewStateManager()
	state, err := sm.Load()
	if err != nil {
		return err
	}

	// Check if already grabbed
	if existingProject, exists := state.Projects[projectName]; exists && existingProject.IsGrabbed {
		return fmt.Errorf("project '%s' is already grabbed at %s", projectName, existingProject.LocalPath)
	}

	// Find project in archive
	archiveProjects, err := core.DiscoverArchiveProjects(state)
	if err != nil {
		return fmt.Errorf("failed to scan archive: %w", err)
	}

	archiveProject, exists := archiveProjects[projectName]
	if !exists {
		return fmt.Errorf("project '%s' not found in archive", projectName)
	}

	// Determine local path
	localRoot := core.GetDefaultLocalPath(archiveProject.Category)
	localPath := filepath.Join(localRoot, projectName)

	// Check if local path already exists
	if _, err := os.Stat(localPath); err == nil {
		return fmt.Errorf("local path already exists: %s (use --force to overwrite)", localPath)
	}

	// Ensure local root exists
	if err := os.MkdirAll(localRoot, 0755); err != nil {
		return fmt.Errorf("failed to create local directory: %w", err)
	}

	// Create the destination directory
	if err := os.MkdirAll(localPath, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	fmt.Printf("Grabbing %s from %s to %s...\n", projectName, archiveProject.Path, localPath)

	// Rsync from archive to local
	if err := core.Rsync(archiveProject.Path, localPath); err != nil {
		// Clean up on failure
		os.RemoveAll(localPath)
		return fmt.Errorf("failed to copy project: %w", err)
	}

	// Update state
	now := time.Now()
	state.Projects[projectName] = &core.Project{
		LocalPath:       localPath,
		Master:          archiveProject.Master,
		ArchiveCategory: archiveProject.Category,
		GrabbedAt:       &now,
		IsGrabbed:       true,
		NoHashMode:      true, // Default to no-hash mode for Phase 1
	}

	if err := sm.Save(state); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	fmt.Printf("Successfully grabbed '%s' to %s\n", projectName, localPath)
	return nil
}
