package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jamespark/parkr/core"
)

// DetectProjectCategory auto-detects project type based on files present.
// Returns "pycharm" for Python projects, "rstudio" for R projects,
// or "code" as default.
func DetectProjectCategory(localPath string) string {
	// Check for Python project indicators
	pythonIndicators := []string{
		"pyproject.toml",
		"requirements.txt",
		"setup.py",
		"Pipfile",
	}
	for _, indicator := range pythonIndicators {
		if _, err := os.Stat(filepath.Join(localPath, indicator)); err == nil {
			return "pycharm"
		}
	}

	// Check for R project indicators
	rIndicators := []string{
		".Rproj",
		"DESCRIPTION",
	}
	for _, indicator := range rIndicators {
		if indicator == ".Rproj" {
			// Check for any .Rproj file
			matches, _ := filepath.Glob(filepath.Join(localPath, "*.Rproj"))
			if len(matches) > 0 {
				return "rstudio"
			}
		} else {
			if _, err := os.Stat(filepath.Join(localPath, indicator)); err == nil {
				return "rstudio"
			}
		}
	}

	// Default to code
	return "code"
}

// AddCmd adds an existing local project to the archive.
// If move is true, the local copy is deleted after successful archiving.
func AddCmd(localPath string, category string, move bool) error {
	sm := core.NewStateManager()
	state, err := sm.Load()
	if err != nil {
		return err
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(localPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Verify local path exists
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("local path does not exist: %s", absPath)
		}
		return fmt.Errorf("failed to stat path: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", absPath)
	}

	// Extract project name from path
	projectName := filepath.Base(absPath)
	if projectName == "." || projectName == "/" {
		return fmt.Errorf("invalid project path: %s", absPath)
	}

	// Check if project is already tracked
	if existingProject, exists := state.Projects[projectName]; exists {
		if existingProject.IsGrabbed {
			return fmt.Errorf("project '%s' is already tracked and grabbed at %s", projectName, existingProject.LocalPath)
		}
		return fmt.Errorf("project '%s' is already tracked in archive", projectName)
	}

	// Auto-detect category if not specified
	if category == "" {
		category = DetectProjectCategory(absPath)
		fmt.Printf("Auto-detected category: %s\n", category)
	}

	// Get master configuration
	masterName := state.DefaultMaster
	master, exists := state.Masters[masterName]
	if !exists {
		return fmt.Errorf("default master '%s' not found", masterName)
	}

	// Get category path from master
	categoryPath, exists := master[category]
	if !exists {
		return fmt.Errorf("category '%s' not found in master '%s'", category, masterName)
	}

	// Ensure category directory exists (auto-create if needed)
	if err := os.MkdirAll(categoryPath, 0755); err != nil {
		return fmt.Errorf("failed to create category directory %s: %w", categoryPath, err)
	}

	// Construct archive path
	archivePath := filepath.Join(categoryPath, projectName)

	// Check if archive path already exists
	if _, err := os.Stat(archivePath); err == nil {
		return fmt.Errorf("archive path already exists: %s", archivePath)
	}

	// Create archive directory
	if err := os.MkdirAll(archivePath, 0755); err != nil {
		return fmt.Errorf("failed to create archive directory: %w", err)
	}

	fmt.Printf("Adding %s to archive at %s...\n", projectName, archivePath)

	// Rsync from local to archive
	if err := core.Rsync(absPath, archivePath); err != nil {
		// Clean up on failure
		os.RemoveAll(archivePath)
		return fmt.Errorf("failed to copy project to archive: %w", err)
	}

	// Get newest mtime from local path for LastParkMtime tracking
	newestInfo, err := core.GetNewestMtime(absPath)
	if err != nil {
		// Clean up on failure
		os.RemoveAll(archivePath)
		return fmt.Errorf("failed to get mtime: %w", err)
	}

	// Create project entry - initially marked as grabbed (we still have local copy)
	now := time.Now()
	project := &core.Project{
		LocalPath:       absPath,
		Master:          masterName,
		ArchiveCategory: category,
		GrabbedAt:       &now,
		LastParkAt:      &now,
		NoHashMode:      true, // Phase 1: no-hash mode
		IsGrabbed:       true, // Initially grabbed (local copy exists)
	}

	// Set LastParkMtime if we got valid mtime info
	if newestInfo != nil && *newestInfo != nil {
		mtime := (*newestInfo).ModTime()
		project.LastParkMtime = &mtime
	}

	state.Projects[projectName] = project

	// If move option is set, delete the local copy
	if move {
		if err := os.RemoveAll(absPath); err != nil {
			// Save state first to indicate project is in archive (but local still exists)
			// This leaves state in a consistent, recoverable state
			if saveErr := sm.Save(state); saveErr != nil {
				return fmt.Errorf("failed to remove local copy: %w (also failed to save state: %v)", err, saveErr)
			}
			return fmt.Errorf("failed to remove local copy after archiving (project added to archive, local kept): %w", err)
		}

		// Update state to reflect that local copy is gone
		project.IsGrabbed = false
		project.LocalPath = ""
		project.GrabbedAt = nil
	}

	// Save final state
	if err := sm.Save(state); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	if move {
		fmt.Printf("Successfully added and moved '%s' to archive\n", projectName)
	} else {
		fmt.Printf("Successfully added '%s' to archive (local copy kept)\n", projectName)
	}

	return nil
}
