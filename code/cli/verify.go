package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jamespark/parkr/core"
)

// VerifyCmd checks state file consistency
func VerifyCmd() error {
	sm := core.NewStateManager()
	state, err := sm.Load()
	if err != nil {
		return err
	}

	fmt.Println("Verifying state file consistency...")
	fmt.Println()

	issues := []string{}
	warnings := []string{}

	// Check master configurations
	for masterName, categories := range state.Masters {
		for categoryName, categoryPath := range categories {
			if _, err := os.Stat(categoryPath); os.IsNotExist(err) {
				warnings = append(warnings, fmt.Sprintf("Master '%s' category '%s' path does not exist: %s", masterName, categoryName, categoryPath))
			}
		}
	}

	// Check each project in state
	for projectName, project := range state.Projects {
		// Check archive path exists
		archivePath, err := state.GetArchivePath(projectName)
		if err != nil {
			issues = append(issues, fmt.Sprintf("Project '%s': %v", projectName, err))
			continue
		}

		if _, err := os.Stat(archivePath); os.IsNotExist(err) {
			issues = append(issues, fmt.Sprintf("Project '%s': archive path does not exist: %s", projectName, archivePath))
		}

		// Check local path if grabbed
		if project.IsGrabbed {
			if project.LocalPath == "" {
				issues = append(issues, fmt.Sprintf("Project '%s': marked as grabbed but no local path set", projectName))
			} else if _, err := os.Stat(project.LocalPath); os.IsNotExist(err) {
				issues = append(issues, fmt.Sprintf("Project '%s': marked as grabbed but local path does not exist: %s", projectName, project.LocalPath))
			}

			// Check for timestamp inconsistencies
			if project.GrabbedAt == nil {
				warnings = append(warnings, fmt.Sprintf("Project '%s': grabbed but no grabbed_at timestamp", projectName))
			}
		} else {
			// Not grabbed but has local path that exists
			if project.LocalPath != "" {
				if _, err := os.Stat(project.LocalPath); err == nil {
					warnings = append(warnings, fmt.Sprintf("Project '%s': not marked as grabbed but local path exists: %s", projectName, project.LocalPath))
				}
			}
		}

		// Check hash consistency
		if !project.NoHashMode && project.LocalContentHash != nil && project.LocalHashComputedAt == nil {
			warnings = append(warnings, fmt.Sprintf("Project '%s': has local hash but no hash computed timestamp", projectName))
		}

		// Check for lastParkAt without lastParkMtime
		if project.LastParkAt != nil && project.LastParkMtime == nil {
			warnings = append(warnings, fmt.Sprintf("Project '%s': has last_park_at but no last_park_mtime", projectName))
		}
	}

	// Check for orphaned local projects (projects in local dirs not tracked)
	localDirs := getLocalDirectoriesFromState(state)
	trackedLocalPaths := make(map[string]bool)
	for _, project := range state.Projects {
		if project.IsGrabbed && project.LocalPath != "" {
			trackedLocalPaths[project.LocalPath] = true
		}
	}

	// Scan local directories for untracked projects
	for _, localDir := range localDirs {
		if _, err := os.Stat(localDir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(localDir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() || entry.Name()[0] == '.' {
				continue
			}

			projectPath := filepath.Join(localDir, entry.Name())
			if !trackedLocalPaths[projectPath] {
				warnings = append(warnings, fmt.Sprintf("Untracked project found in local directory: %s", projectPath))
			}
		}
	}

	// Report results
	if len(issues) == 0 && len(warnings) == 0 {
		fmt.Printf("%s State file is consistent. No issues found.\n", SymbolCheck)
		return nil
	}

	if len(issues) > 0 {
		fmt.Println("ERRORS (require attention):")
		for _, issue := range issues {
			fmt.Printf("  %s %s\n", SymbolCross, issue)
		}
		fmt.Println()
	}

	if len(warnings) > 0 {
		fmt.Println("WARNINGS (potential issues):")
		for _, warning := range warnings {
			fmt.Printf("  %s %s\n", SymbolWarning, warning)
		}
		fmt.Println()
	}

	fmt.Printf("Summary: %d errors, %d warnings\n", len(issues), len(warnings))

	// Check local directories info
	fmt.Println()
	fmt.Println("Local directories configured:")
	for _, dir := range localDirs {
		exists := "exists"
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			exists = "does not exist"
		}
		fmt.Printf("  - %s (%s)\n", dir, exists)
	}

	if len(issues) > 0 {
		return fmt.Errorf("state verification found %d errors", len(issues))
	}

	return nil
}
