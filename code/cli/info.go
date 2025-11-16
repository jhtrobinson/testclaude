package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/jamespark/parkr/core"
)

// InfoCmd shows detailed information about a specific project
func InfoCmd(projectName string) error {
	sm := core.NewStateManager()
	state, err := sm.Load()
	if err != nil {
		return err
	}

	// Check if project exists in state
	project, exists := state.Projects[projectName]
	if !exists {
		// Check if it exists in archive but not tracked
		archiveProjects, err := core.DiscoverArchiveProjects(state)
		if err != nil {
			return err
		}

		if archiveProject, found := archiveProjects[projectName]; found {
			// Project exists in archive but not tracked in state
			fmt.Printf("Project: %s\n", projectName)
			fmt.Printf("Archive: %s", archiveProject.Path)
			if size, err := core.GetDirSize(archiveProject.Path); err == nil {
				fmt.Printf(" (%s)", core.FormatSize(size))
			}
			fmt.Println()
			fmt.Printf("Local: not checked out\n")
			fmt.Printf("Status: Archived (not tracked in state)\n")
			fmt.Printf("Archive exists: Yes\n")
			fmt.Printf("Local exists: No\n")
			return nil
		}

		return fmt.Errorf("project '%s' not found in state or archive", projectName)
	}

	// Get archive path
	archivePath, err := state.GetArchivePath(projectName)
	if err != nil {
		return err
	}

	// Print project information
	fmt.Printf("Project: %s\n", projectName)

	// Archive info
	archiveExists := false
	if info, err := os.Stat(archivePath); err == nil && info.IsDir() {
		archiveExists = true
		fmt.Printf("Archive: %s", archivePath)
		if size, err := core.GetDirSize(archivePath); err == nil {
			fmt.Printf(" (%s)", core.FormatSize(size))
		}
		fmt.Println()
	} else {
		fmt.Printf("Archive: %s (missing)\n", archivePath)
	}

	// Local info
	localExists := false
	var lastModified time.Time
	if project.IsGrabbed {
		if info, err := os.Stat(project.LocalPath); err == nil && info.IsDir() {
			localExists = true
			fmt.Printf("Local: %s", project.LocalPath)
			if size, err := core.GetDirSize(project.LocalPath); err == nil {
				fmt.Printf(" (%s)", core.FormatSize(size))
			}
			fmt.Println()

			// Get last modified time
			if newest, err := core.GetNewestMtime(project.LocalPath); err == nil && newest != nil {
				lastModified = (*newest).ModTime()
			}
		} else {
			fmt.Printf("Local: %s (missing)\n", project.LocalPath)
		}
	} else {
		fmt.Printf("Local: not checked out\n")
	}

	// Timestamps
	if project.GrabbedAt != nil {
		fmt.Printf("Checked out: %s\n", project.GrabbedAt.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Printf("Checked out: never\n")
	}

	if project.LastParkAt != nil {
		fmt.Printf("Last checkin: %s\n", project.LastParkAt.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Printf("Last checkin: never\n")
	}

	if !lastModified.IsZero() {
		fmt.Printf("Last modified: %s\n", lastModified.Format("2006-01-02 15:04:05"))
	} else if project.IsGrabbed {
		fmt.Printf("Last modified: unknown\n")
	}

	// Status
	if project.IsGrabbed && localExists {
		statusInfo := determineStatusInfo(project, lastModified)
		fmt.Printf("Status: %s\n", statusInfo.Text)
	} else if !project.IsGrabbed {
		fmt.Printf("Status: Archived\n")
	} else {
		fmt.Printf("Status: Unknown\n")
	}

	// Existence checks
	fmt.Printf("Archive exists: %s\n", boolToYesNo(archiveExists))
	fmt.Printf("Local exists: %s\n", boolToYesNo(localExists))

	// Hash mode info
	if project.NoHashMode {
		fmt.Printf("Hash mode: disabled (mtime-based verification)\n")
	} else if project.LocalContentHash != nil {
		fmt.Printf("Hash mode: enabled\n")
		if project.LocalHashComputedAt != nil {
			fmt.Printf("Hash computed: %s\n", project.LocalHashComputedAt.Format("2006-01-02 15:04:05"))
		}
	}

	return nil
}

func boolToYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
