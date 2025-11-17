package cli

import (
	"fmt"
	"os"

	"github.com/jamespark/parkr/core"
)

// RemoveCmd removes a project from the archive (and optionally local)
func RemoveCmd(projectName string, archiveOnly bool, localOnly bool, everywhere bool, confirm bool) error {
	sm := core.NewStateManager()
	state, err := sm.Load()
	if err != nil {
		return err
	}

	// Check if project exists in state
	project, exists := state.Projects[projectName]
	if !exists {
		return fmt.Errorf("project '%s' not found in state", projectName)
	}

	// Get archive path
	archivePath, err := state.GetArchivePath(projectName)
	if err != nil {
		return err
	}

	// Determine what to remove
	removeArchive := !localOnly || everywhere
	removeLocal := localOnly || everywhere

	// Default behavior: remove from archive
	if !localOnly && !everywhere {
		removeArchive = true
		removeLocal = false
	}

	// Confirmation (unless --confirm flag is set to skip)
	if !confirm {
		fmt.Printf("About to remove project '%s':\n", projectName)
		if removeArchive {
			fmt.Printf("  - Archive: %s\n", archivePath)
		}
		if removeLocal && project.IsGrabbed {
			fmt.Printf("  - Local: %s\n", project.LocalPath)
		}
		fmt.Print("Continue? [y/N] ")

		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// Remove archive copy
	if removeArchive {
		if _, err := os.Stat(archivePath); err == nil {
			fmt.Printf("Removing archive copy at %s...\n", archivePath)
			if err := os.RemoveAll(archivePath); err != nil {
				return fmt.Errorf("failed to remove archive copy: %w", err)
			}
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("failed to check archive path: %w", err)
		} else {
			fmt.Printf("Warning: archive path does not exist: %s\n", archivePath)
		}
	}

	// Remove local copy
	if removeLocal && project.IsGrabbed {
		if _, err := os.Stat(project.LocalPath); err == nil {
			fmt.Printf("Removing local copy at %s...\n", project.LocalPath)
			if err := os.RemoveAll(project.LocalPath); err != nil {
				return fmt.Errorf("failed to remove local copy: %w", err)
			}
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("failed to check local path: %w", err)
		}
		project.IsGrabbed = false
	}

	// Remove project from state if archive was removed
	if removeArchive {
		delete(state.Projects, projectName)
		fmt.Printf("Removed project '%s' from state\n", projectName)
	} else if removeLocal {
		// Just update grabbed status
		if err := sm.Save(state); err != nil {
			return fmt.Errorf("failed to update state: %w", err)
		}
	}

	// Save state
	if removeArchive {
		if err := sm.Save(state); err != nil {
			return fmt.Errorf("failed to update state: %w", err)
		}
	}

	fmt.Printf("Successfully removed '%s'\n", projectName)
	return nil
}
