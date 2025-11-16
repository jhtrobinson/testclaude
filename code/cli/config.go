package cli

import (
	"fmt"

	"github.com/jamespark/parkr/core"
)

// ConfigCmd shows current configuration
func ConfigCmd() error {
	sm := core.NewStateManager()
	state, err := sm.Load()
	if err != nil {
		return err
	}

	fmt.Println("PARKR CONFIGURATION")
	fmt.Println()

	// State file location
	fmt.Printf("State file: %s\n", sm.StatePath())
	fmt.Println()

	// Default master
	fmt.Printf("Default master: %s\n", state.DefaultMaster)
	fmt.Println()

	// Masters and their categories
	fmt.Println("Archive masters:")
	for masterName, categories := range state.Masters {
		defaultMark := ""
		if masterName == state.DefaultMaster {
			defaultMark = " (default)"
		}
		fmt.Printf("  %s%s:\n", masterName, defaultMark)

		for categoryName, categoryPath := range categories {
			fmt.Printf("    %s: %s\n", categoryName, categoryPath)
		}
	}
	fmt.Println()

	// Local directories
	localDirs := getLocalDirectoriesFromState(state)
	if len(state.LocalDirectories) > 0 {
		fmt.Println("Local directories (configured in state):")
	} else {
		fmt.Println("Local directories (using defaults):")
	}
	for _, dir := range localDirs {
		fmt.Printf("  - %s\n", dir)
	}
	fmt.Println()

	// Statistics
	totalProjects := len(state.Projects)
	grabbedCount := 0
	for _, project := range state.Projects {
		if project.IsGrabbed {
			grabbedCount++
		}
	}

	fmt.Println("Statistics:")
	fmt.Printf("  Total tracked projects: %d\n", totalProjects)
	fmt.Printf("  Currently checked out: %d\n", grabbedCount)
	fmt.Printf("  Archived: %d\n", totalProjects-grabbedCount)

	return nil
}
