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
			if err := verifyByMtime(project, projectName); err != nil {
				return err
			}
		} else {
			// Hash verification (default for projects with hashes)
			if err := verifyByHash(project, projectName); err != nil {
				return err
			}
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

// verifyByMtime checks if project has been modified using mtime comparison
func verifyByMtime(project *core.Project, projectName string) error {
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
	return nil
}

// verifyByHash checks if project matches archive using SHA256 hash comparison
func verifyByHash(project *core.Project, projectName string) error {
	if project.NoHashMode {
		return fmt.Errorf("project '%s' is in no-hash mode - cannot use hash verification", projectName)
	}

	if project.ArchiveContentHash == nil || project.LocalHashComputedAt == nil {
		return fmt.Errorf("project '%s' does not have stored hashes - use --no-hash or park without --no-hash first", projectName)
	}

	// Dirty detection: quick check using mtime vs hash computed time
	fmt.Println("Checking for modifications...")
	newestInfo, err := core.GetNewestMtime(project.LocalPath)
	if err != nil {
		return fmt.Errorf("failed to check local files: %w", err)
	}

	if newestInfo != nil && *newestInfo != nil {
		currentMtime := (*newestInfo).ModTime()
		if currentMtime.After(*project.LocalHashComputedAt) {
			return fmt.Errorf("project '%s' has been modified since hash was computed (newest: %s, hash computed: %s). Park first or use --force",
				projectName, currentMtime.Format("2006-01-02 15:04:05"), project.LocalHashComputedAt.Format("2006-01-02 15:04:05"))
		}
	}

	// Compute current local hash and compare to archive hash
	fmt.Println("Computing local hash for verification...")
	currentHash, err := core.ComputeProjectHash(project.LocalPath)
	if err != nil {
		return fmt.Errorf("failed to compute local hash: %w", err)
	}

	if currentHash != *project.ArchiveContentHash {
		return fmt.Errorf("project '%s' hash does not match archive (local: %s, archive: %s). Park first or use --force",
			projectName, currentHash[:16]+"...", (*project.ArchiveContentHash)[:16]+"...")
	}

	fmt.Println("Hash verification passed.")
	return nil
}
