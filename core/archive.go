package core

import (
	"fmt"
	"os"
	"path/filepath"
)

// DiscoverArchiveProjects finds all projects in archive directories
func DiscoverArchiveProjects(state *State) (map[string]ArchiveProject, error) {
	projects := make(map[string]ArchiveProject)

	for masterName, categories := range state.Masters {
		for categoryName, categoryPath := range categories {
			entries, err := os.ReadDir(categoryPath)
			if err != nil {
				if os.IsNotExist(err) {
					continue // Skip non-existent directories
				}
				return nil, fmt.Errorf("failed to read %s: %w", categoryPath, err)
			}

			for _, entry := range entries {
				if entry.IsDir() {
					projectName := entry.Name()
					// Skip hidden directories
					if projectName[0] == '.' {
						continue
					}

					projects[projectName] = ArchiveProject{
						Name:     projectName,
						Master:   masterName,
						Category: categoryName,
						Path:     filepath.Join(categoryPath, projectName),
					}
				}
			}
		}
	}

	return projects, nil
}

// ArchiveProject represents a project found in the archive
type ArchiveProject struct {
	Name     string
	Master   string
	Category string
	Path     string
}

// GetNewestMtime finds the newest modification time in a directory tree
func GetNewestMtime(dirPath string) (*os.FileInfo, error) {
	var newest os.FileInfo
	var newestTime int64

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if info.ModTime().Unix() > newestTime {
				newestTime = info.ModTime().Unix()
				newest = info
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &newest, nil
}

// GetDirSize calculates the total size of a directory
func GetDirSize(dirPath string) (int64, error) {
	var size int64

	err := filepath.Walk(dirPath, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// FormatSize formats bytes into human-readable format
func FormatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.1f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
