package cli

import (
	"fmt"
	"os"

	"github.com/jamespark/parkr/core"
)

// RmCmd removes the local copy of a project
func RmCmd(projectName string, noHash bool, force bool) error {
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
		// Local path doesn't exist, just update state
		fmt.Printf("Warning: local path does not exist: %s\n", project.LocalPath)
		project.IsGrabbed = false
		if err := sm.Save(state); err != nil {
			return fmt.Errorf("failed to update state: %w", err)
		}
		fmt.Printf("Updated state for '%s'\n", projectName)
		return nil
	}

	// Safety verification
	if !force {
		if project.NoHashMode && !noHash {
			return fmt.Errorf("project '%s' was parked with --no-hash. Use --no-hash or --force to delete", projectName)
		}

		if noHash || project.NoHashMode {
			// Mtime verification
			if project.LastParkMtime == nil {
				return fmt.Errorf("project '%s' has never been parked - cannot verify safety", projectName)
			}

			newestInfo, err := core.GetNewestMtime(project.LocalPath)
			if err != nil {
				return fmt.Errorf("failed to check local files: %w", err)
			}

			if newestInfo != nil && *newestInfo != nil {
				currentMtime := (*newestInfo).ModTime()
				if currentMtime.After(*project.LastParkMtime) {
					return fmt.Errorf("project '%s' has been modified since last park (newest: %s, parked: %s). Park first or use --force",
						projectName, currentMtime.Format("2006-01-02 15:04:05"), project.LastParkMtime.Format("2006-01-02 15:04:05"))
				}
			}

			fmt.Println("Mtime verification passed.")
		} else {
			// Hash verification
			if project.LocalContentHash == nil {
				return fmt.Errorf("project '%s' has no stored hash - park with hashing first or use --no-hash", projectName)
			}

			// Check if files were modified since hash was computed
			if project.LocalHashComputedAt != nil {
				newestInfo, err := core.GetNewestMtime(project.LocalPath)
				if err != nil {
					return fmt.Errorf("failed to check local files: %w", err)
				}
				if newestInfo != nil && *newestInfo != nil {
					currentMtime := (*newestInfo).ModTime()
					if currentMtime.After(*project.LocalHashComputedAt) {
						fmt.Printf("Warning: files modified since hash was computed (newest: %s, hash computed: %s)\n",
							currentMtime.Format("2006-01-02 15:04:05"), project.LocalHashComputedAt.Format("2006-01-02 15:04:05"))
						fmt.Println("Recomputing hash to verify...")
					}
				}
			}

			fmt.Println("Computing current local hash...")
			currentHash, err := core.ComputeProjectHash(project.LocalPath)
			if err != nil {
				return fmt.Errorf("failed to compute local hash: %w", err)
			}

			if currentHash != *project.LocalContentHash {
				return fmt.Errorf("hash mismatch - local content has changed since last park:\n"+
					"  Stored hash:  %s\n"+
					"  Current hash: %s\n"+
					"Park your changes first or use --force to delete anyway",
					*project.LocalContentHash, currentHash)
			}

			fmt.Println("Hash verification passed.")
		}
	} else {
		fmt.Println("Warning: Skipping verification (--force)")
	}

	// Delete local copy
	if os.Getenv("PARKR_ALIVE") == "" {
		// Mock mode - just print the command
		fmt.Printf("rm -rf %s\n", project.LocalPath)
		fmt.Println("(mock mode - set PARKR_ALIVE=1 to actually delete)")
		return nil
	}

	fmt.Printf("Removing local copy at %s...\n", project.LocalPath)
	if err := os.RemoveAll(project.LocalPath); err != nil {
		return fmt.Errorf("failed to remove local copy: %w", err)
	}

	// Update state
	project.IsGrabbed = false
	if err := sm.Save(state); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	fmt.Printf("Successfully removed local copy of '%s'\n", projectName)
	return nil
}
