package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jamespark/parkr/core"
)

// RemoveCmd removes a project from the archive and/or local
func RemoveCmd(projectName string, localOnly bool, archive bool, yes bool) error {
	sm := core.NewStateManager()
	state, err := sm.Load()
	if err != nil {
		return fmt.Errorf("state file error: %w", err)
	}

	// Check if project exists in state
	project, existsInState := state.Projects[projectName]

	// If --local is specified, delegate to rm command behavior
	if localOnly {
		if !existsInState || !project.IsGrabbed {
			return fmt.Errorf("project '%s' is not currently grabbed", projectName)
		}
		// Implement local removal directly (same as rm --force but with confirmation)
		fmt.Println("WARNING: This will remove the LOCAL copy only.")
		fmt.Printf("Project: %s\n", projectName)
		fmt.Printf("Local:   %s\n", project.LocalPath)
		fmt.Println()
		fmt.Println("Note: Archive copy will be kept safe.")
		fmt.Println()

		// Verify local path exists
		localInfo, err := os.Stat(project.LocalPath)
		if os.IsNotExist(err) {
			// Local path doesn't exist, just update state
			fmt.Printf("Warning: local path does not exist: %s\n", project.LocalPath)
			project.IsGrabbed = false
			if err := sm.Save(state); err != nil {
				return fmt.Errorf("failed to update state: %w", err)
			}
			fmt.Printf("Updated state for '%s'\n", projectName)
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to check local path: %w", err)
		}
		if !localInfo.IsDir() {
			return fmt.Errorf("local path is not a directory: %s", project.LocalPath)
		}

		// Get local size for display
		localSize, err := core.GetDirSize(project.LocalPath)
		if err != nil {
			return fmt.Errorf("failed to calculate local size: %w", err)
		}
		fmt.Printf("Local size: %s\n", core.FormatSize(localSize))

		// Check for uncommitted changes (safety verification)
		if project.LastParkMtime != nil {
			newestInfo, err := core.GetNewestMtime(project.LocalPath)
			if err != nil {
				fmt.Printf("Warning: could not check for uncommitted changes: %v\n", err)
			} else if newestInfo != nil && *newestInfo != nil {
				currentMtime := (*newestInfo).ModTime()
				if currentMtime.After(*project.LastParkMtime) {
					fmt.Println()
					fmt.Println("WARNING: Local copy has uncommitted changes!")
					fmt.Printf("  Last parked: %s\n", project.LastParkMtime.Format("2006-01-02 15:04:05"))
					fmt.Printf("  Newest file: %s\n", currentMtime.Format("2006-01-02 15:04:05"))
					fmt.Println("  Consider running 'parkr park' first to save your changes.")
					fmt.Println()
				}
			}
		} else {
			fmt.Println()
			fmt.Println("WARNING: Project has never been parked - changes may be lost!")
			fmt.Println()
		}

		// Confirm deletion
		if !yes {
			fmt.Print("Type the project name to confirm deletion: ")
			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			input = strings.TrimSpace(input)

			if input != projectName {
				return fmt.Errorf("confirmation failed - project name did not match")
			}
		}

		// Update state FIRST (before deletion for atomicity)
		project.IsGrabbed = false
		if err := sm.Save(state); err != nil {
			return fmt.Errorf("failed to update state: %w", err)
		}

		// Delete local copy
		fmt.Printf("Removing local copy at %s...\n", project.LocalPath)
		if err := os.RemoveAll(project.LocalPath); err != nil {
			// Rollback state on failure
			project.IsGrabbed = true
			_ = sm.Save(state)
			return fmt.Errorf("failed to remove local copy: %w", err)
		}

		fmt.Printf("Successfully removed local copy of '%s'\n", projectName)
		return nil
	}

	// Get archive path
	var archivePath string
	if existsInState {
		archivePath, err = state.GetArchivePath(projectName)
		if err != nil {
			return fmt.Errorf("failed to get archive path: %w", err)
		}
	} else {
		// Project not in state, try to discover it
		archiveProjects, err := core.DiscoverArchiveProjects(state)
		if err != nil {
			return fmt.Errorf("archive not accessible: %w", err)
		}

		archiveProject, found := archiveProjects[projectName]
		if !found {
			return fmt.Errorf("project '%s' not found in archive or state", projectName)
		}
		archivePath = archiveProject.Path
	}

	// Verify archive exists
	archiveInfo, err := os.Stat(archivePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("archive path does not exist: %s", archivePath)
	}
	if err != nil {
		return fmt.Errorf("archive not accessible: %w", err)
	}
	if !archiveInfo.IsDir() {
		return fmt.Errorf("archive path is not a directory: %s", archivePath)
	}

	// Calculate archive size for warning
	archiveSize, err := core.GetDirSize(archivePath)
	if err != nil {
		return fmt.Errorf("failed to calculate archive size: %w", err)
	}

	// Determine what will be removed
	var localPath string
	var localExists bool
	if existsInState && project.IsGrabbed {
		localPath = project.LocalPath
		localInfo, err := os.Stat(localPath)
		localExists = err == nil && localInfo.IsDir()
	}

	// Build warning message
	fmt.Println("WARNING: This operation will permanently delete project files!")
	fmt.Println()

	if archive {
		// --archive removes state + archive, preserves local
		fmt.Printf("Project: %s\n", projectName)
		fmt.Printf("Archive: %s (%s)\n", archivePath, core.FormatSize(archiveSize))
		fmt.Println()
		fmt.Println("This will remove the archive copy and state entry.")
		if localExists {
			fmt.Printf("Note: Local copy at %s will be preserved.\n", localPath)
		}
	} else {
		fmt.Printf("Project: %s\n", projectName)
		fmt.Printf("Archive: %s (%s)\n", archivePath, core.FormatSize(archiveSize))
		fmt.Println()
		fmt.Println("This will remove the archive copy ONLY.")
		if localExists {
			fmt.Printf("Note: Local copy at %s will be kept.\n", localPath)
		}
	}

	// Warn if this is the only copy
	if !localExists {
		fmt.Println()
		fmt.Println("DANGER: This is the ONLY copy of this project!")
		fmt.Println("Deletion will result in permanent data loss.")
	}

	fmt.Println()

	// Confirm deletion
	if !yes {
		fmt.Print("Type the project name to confirm deletion: ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
		input = strings.TrimSpace(input)

		if input != projectName {
			return fmt.Errorf("confirmation failed - project name did not match")
		}
	}

	// Update state FIRST (before deletions for better atomicity)
	// Save original state for potential rollback
	var originalIsGrabbed bool
	if existsInState {
		originalIsGrabbed = project.IsGrabbed
		delete(state.Projects, projectName)
		if err := sm.Save(state); err != nil {
			return fmt.Errorf("failed to update state: %w", err)
		}
		fmt.Println("State updated.")
	}

	// Remove archive copy (--archive removes state + archive, preserves local)
	fmt.Printf("Removing archive at %s...\n", archivePath)
	if err := os.RemoveAll(archivePath); err != nil {
		// Rollback state on failure
		if existsInState {
			state.Projects[projectName] = project
			state.Projects[projectName].IsGrabbed = originalIsGrabbed
			_ = sm.Save(state)
		}
		return fmt.Errorf("failed to remove archive: %w", err)
	}
	fmt.Println("Archive copy removed.")

	fmt.Printf("\nSuccessfully removed project '%s'\n", projectName)
	return nil
}
