package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jamespark/parkr/core"
)

// ListCmd lists all projects in archive
func ListCmd(category string) error {
	sm := core.NewStateManager()
	state, err := sm.Load()
	if err != nil {
		return err
	}

	// Discover projects in archive
	archiveProjects, err := core.DiscoverArchiveProjects(state)
	if err != nil {
		return fmt.Errorf("failed to scan archive: %w", err)
	}

	if len(archiveProjects) == 0 {
		fmt.Println("No projects found in archive.")
		return nil
	}

	// Filter by category if specified
	var projects []core.ArchiveProject
	for _, p := range archiveProjects {
		if category == "" || p.Category == category {
			projects = append(projects, p)
		}
	}

	// Sort by name
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	// Print header
	fmt.Printf("%-30s %-12s %-12s %s\n", "PROJECT", "CATEGORY", "SIZE", "STATUS")
	fmt.Println(strings.Repeat("-", 70))

	// Print each project
	for _, ap := range projects {
		status := "archived"

		// Check if grabbed in state
		if stateProject, exists := state.Projects[ap.Name]; exists && stateProject.IsGrabbed {
			status = "grabbed"
		}

		// Get size
		size, err := core.GetDirSize(ap.Path)
		sizeStr := "?"
		if err == nil {
			sizeStr = core.FormatSize(size)
		}

		fmt.Printf("%-30s %-12s %-12s %s\n", ap.Name, ap.Category, sizeStr, status)
	}

	return nil
}
