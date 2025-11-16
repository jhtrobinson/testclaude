package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jamespark/parkr/core"
)

// LocalCmd shows all projects in local directories
func LocalCmd(unmanagedOnly bool) error {
	sm := core.NewStateManager()
	state, err := sm.Load()
	if err != nil {
		return err
	}

	// Get all local directories to scan
	localDirs := getLocalDirectories()

	// Build a set of managed projects (by local path)
	managedPaths := make(map[string]string) // path -> project name
	for name, project := range state.Projects {
		if project.IsGrabbed {
			managedPaths[project.LocalPath] = name
		}
	}

	type localProject struct {
		name      string
		path      string
		size      int64
		isManaged bool
	}

	var projects []localProject

	// Scan each local directory
	for _, localDir := range localDirs {
		if _, err := os.Stat(localDir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(localDir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			// Skip hidden directories
			if entry.Name()[0] == '.' {
				continue
			}

			projectPath := filepath.Join(localDir, entry.Name())
			projectName := entry.Name()

			// Check if managed
			managedName, isManaged := managedPaths[projectPath]
			if isManaged {
				projectName = managedName
			}

			// Skip if only showing unmanaged and this is managed
			if unmanagedOnly && isManaged {
				continue
			}

			// Get size
			var size int64
			if s, err := core.GetDirSize(projectPath); err == nil {
				size = s
			}

			projects = append(projects, localProject{
				name:      projectName,
				path:      projectPath,
				size:      size,
				isManaged: isManaged,
			})
		}
	}

	if len(projects) == 0 {
		if unmanagedOnly {
			fmt.Println("No unmanaged projects found in local directories.")
		} else {
			fmt.Println("No projects found in local directories.")
		}
		return nil
	}

	// Sort by path
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].path < projects[j].path
	})

	// Print header
	if unmanagedOnly {
		fmt.Println("UNMANAGED LOCAL PROJECTS:")
	} else {
		fmt.Println("LOCAL PROJECTS:")
	}
	fmt.Printf("%-25s %-40s %-12s %s\n", "NAME", "PATH", "SIZE", "STATUS")
	fmt.Println(strings.Repeat("-", 95))

	// Print each project
	for _, p := range projects {
		sizeStr := core.FormatSize(p.size)
		statusStr := "unmanaged"
		if p.isManaged {
			statusStr = "managed"
		}

		// Truncate path if too long
		pathStr := p.path
		if len(pathStr) > 38 {
			pathStr = "..." + pathStr[len(pathStr)-35:]
		}

		fmt.Printf("%-25s %-40s %-12s %s\n", p.name, pathStr, sizeStr, statusStr)
	}

	// Summary
	managed := 0
	unmanaged := 0
	for _, p := range projects {
		if p.isManaged {
			managed++
		} else {
			unmanaged++
		}
	}

	fmt.Println()
	fmt.Printf("Total: %d projects (%d managed, %d unmanaged)\n", len(projects), managed, unmanaged)

	return nil
}

// getLocalDirectories returns all directories that should be scanned for local projects
func getLocalDirectories() []string {
	homeDir, _ := os.UserHomeDir()

	return []string{
		filepath.Join(homeDir, "code"),
		filepath.Join(homeDir, "PycharmProjects"),
		filepath.Join(homeDir, "RStudioProjects"),
	}
}
