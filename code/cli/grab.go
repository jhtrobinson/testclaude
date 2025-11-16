package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jamespark/parkr/core"
)

// GrabCmd checks out a project from archive to local
func GrabCmd(projectName string, force bool, customPath string) error {
	sm := core.NewStateManager()
	state, err := sm.Load()
	if err != nil {
		return err
	}

	// Check if already grabbed
	if existingProject, exists := state.Projects[projectName]; exists && existingProject.IsGrabbed {
		if !force {
			return fmt.Errorf("project '%s' is already grabbed at %s (use --force to overwrite)", projectName, existingProject.LocalPath)
		}
		fmt.Printf("Warning: project '%s' is already grabbed, --force specified, overwriting...\n", projectName)
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
	var localPath string
	if customPath != "" {
		// Expand ~ to home directory
		if len(customPath) > 0 && customPath[0] == '~' {
			homeDir, _ := os.UserHomeDir()
			localPath = filepath.Join(homeDir, customPath[1:])
		} else {
			localPath = customPath
		}
		// Convert relative to absolute path
		if !filepath.IsAbs(localPath) {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
			localPath = filepath.Join(cwd, localPath)
		}
	} else {
		localRoot := core.GetDefaultLocalPath(archiveProject.Category)
		localPath = filepath.Join(localRoot, projectName)
	}

	// Check if local path already exists
	if _, err := os.Stat(localPath); err == nil {
		if !force {
			return fmt.Errorf("local path already exists: %s (use --force to overwrite)", localPath)
		}
		fmt.Printf("Warning: removing existing local copy at %s...\n", localPath)
		if err := os.RemoveAll(localPath); err != nil {
			return fmt.Errorf("failed to remove existing local copy: %w", err)
		}
	}

	// Ensure parent directory exists
	parentDir := filepath.Dir(localPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
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
