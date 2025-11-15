package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jamespark/parkr/core"
)

// AddCmd adds a local project to the archive
func AddCmd(localPath string, category string) error {
	sm := core.NewStateManager()
	state, err := sm.Load()
	if err != nil {
		return err
	}

	// Resolve absolute path
	absPath, err := filepath.Abs(localPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Verify local path exists
	info, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("local path does not exist: %s", absPath)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", absPath)
	}

	// Get project name from directory name
	projectName := filepath.Base(absPath)

	// Check if already tracked
	if _, exists := state.Projects[projectName]; exists {
		return fmt.Errorf("project '%s' is already tracked", projectName)
	}

	// Auto-detect category if not specified
	if category == "" {
		category = detectProjectType(absPath)
		fmt.Printf("Auto-detected category: %s\n", category)
	}

	// Get archive path for this category
	master := state.DefaultMaster
	masterCategories, exists := state.Masters[master]
	if !exists {
		return fmt.Errorf("default master '%s' not found", master)
	}

	categoryPath, exists := masterCategories[category]
	if !exists {
		return fmt.Errorf("category '%s' not found in master '%s'", category, master)
	}

	// Ensure category directory exists
	if err := os.MkdirAll(categoryPath, 0755); err != nil {
		return fmt.Errorf("failed to create category directory: %w", err)
	}

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

	// Get newest mtime
	newestInfo, err := core.GetNewestMtime(absPath)
	if err != nil {
		return fmt.Errorf("failed to get mtime: %w", err)
	}

	// Update state
	now := time.Now()
	project := &core.Project{
		LocalPath:       absPath,
		Master:          master,
		ArchiveCategory: category,
		GrabbedAt:       &now,
		LastParkAt:      &now,
		IsGrabbed:       true,
		NoHashMode:      true,
	}

	if newestInfo != nil && *newestInfo != nil {
		mtime := (*newestInfo).ModTime()
		project.LastParkMtime = &mtime
	}

	state.Projects[projectName] = project

	if err := sm.Save(state); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	fmt.Printf("Successfully added '%s' to archive\n", projectName)
	fmt.Printf("Local: %s\n", absPath)
	fmt.Printf("Archive: %s\n", archivePath)
	return nil
}

// detectProjectType auto-detects the project category based on files present
func detectProjectType(path string) string {
	// Check for Python
	pythonFiles := []string{"pyproject.toml", "requirements.txt", "setup.py", "Pipfile"}
	for _, f := range pythonFiles {
		if _, err := os.Stat(filepath.Join(path, f)); err == nil {
			return "pycharm"
		}
	}

	// Check for R
	rFiles := []string{".Rproj", "DESCRIPTION"}
	for _, f := range rFiles {
		if _, err := os.Stat(filepath.Join(path, f)); err == nil {
			return "rstudio"
		}
	}
	// Also check for .Rproj files
	matches, _ := filepath.Glob(filepath.Join(path, "*.Rproj"))
	if len(matches) > 0 {
		return "rstudio"
	}

	// Default to code
	return "code"
}
