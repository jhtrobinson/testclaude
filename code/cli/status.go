package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/jamespark/parkr/core"
)

// StatusCmd shows all grabbed projects with sync status
func StatusCmd() error {
	sm := core.NewStateManager()
	state, err := sm.Load()
	if err != nil {
		return err
	}

	// Collect all grabbed projects
	type projectInfo struct {
		name         string
		project      *core.Project
		localSize    int64
		lastModified time.Time
	}

	var grabbedProjects []projectInfo

	for name, project := range state.Projects {
		if !project.IsGrabbed {
			continue
		}

		info := projectInfo{
			name:    name,
			project: project,
		}

		// Get local size and last modified time
		if _, err := os.Stat(project.LocalPath); err == nil {
			if size, err := core.GetDirSize(project.LocalPath); err == nil {
				info.localSize = size
			}
			if newest, err := core.GetNewestMtime(project.LocalPath); err == nil && newest != nil {
				info.lastModified = (*newest).ModTime()
			}
		}

		grabbedProjects = append(grabbedProjects, info)
	}

	if len(grabbedProjects) == 0 {
		fmt.Println("No projects currently checked out.")
		return nil
	}

	// Sort by name
	sort.Slice(grabbedProjects, func(i, j int) bool {
		return grabbedProjects[i].name < grabbedProjects[j].name
	})

	// Print header
	fmt.Println("CHECKED OUT PROJECTS:")
	fmt.Printf("%-20s %-12s %-16s %-16s %s\n", "PROJECT", "LOCAL SIZE", "LAST MODIFIED", "LAST CHECKIN", "STATUS")
	fmt.Println(strings.Repeat("-", 90))

	// Print each project
	for _, p := range grabbedProjects {
		sizeStr := core.FormatSize(p.localSize)
		modifiedStr := formatTimeAgo(p.lastModified)

		// Last checkin time
		checkinStr := "never"
		if p.project.LastParkAt != nil {
			checkinStr = formatTimeAgo(*p.project.LastParkAt)
		}

		// Determine status
		status := determineStatus(p.project, p.lastModified)

		fmt.Printf("%-20s %-12s %-16s %-16s %s\n", p.name, sizeStr, modifiedStr, checkinStr, status)
	}

	return nil
}

// StatusInfo contains the emoji and text components of a status
type StatusInfo struct {
	Emoji string
	Text  string
}

// String returns the full status string with emoji
func (s StatusInfo) String() string {
	return s.Emoji + " " + s.Text
}

// determineStatusInfo determines the sync status of a project and returns separate components
func determineStatusInfo(project *core.Project, lastModified time.Time) StatusInfo {
	// Never checked in
	if project.LastParkAt == nil {
		return StatusInfo{Emoji: SymbolCross, Text: "Never checked in"}
	}

	// Check if modified after last park
	if project.LastParkMtime != nil {
		if lastModified.After(*project.LastParkMtime) {
			return StatusInfo{Emoji: SymbolWarning, Text: "Has uncommitted work"}
		}
	} else {
		// Fallback to comparing with LastParkAt
		if lastModified.After(*project.LastParkAt) {
			return StatusInfo{Emoji: SymbolWarning, Text: "Has uncommitted work"}
		}
	}

	return StatusInfo{Emoji: SymbolCheck, Text: "Safe to delete"}
}

// determineStatus determines the sync status of a project (backward compatible)
func determineStatus(project *core.Project, lastModified time.Time) string {
	return determineStatusInfo(project, lastModified).String()
}

// Time duration constants for formatTimeAgo
const (
	hoursPerDay   = 24
	daysPerWeek   = 7
	daysPerMonth  = 30 // Approximate average
	hoursPerWeek  = hoursPerDay * daysPerWeek
	hoursPerMonth = hoursPerDay * daysPerMonth
)

// formatTimeAgo formats a time as a human-readable relative string
func formatTimeAgo(t time.Time) string {
	if t.IsZero() {
		return "never"
	}

	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	}
	if duration < time.Hour {
		mins := int(duration.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", mins)
	}
	if duration < hoursPerDay*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	if duration < hoursPerWeek*time.Hour {
		days := int(duration.Hours() / hoursPerDay)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
	if duration < hoursPerMonth*time.Hour {
		weeks := int(duration.Hours() / hoursPerWeek)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	}

	months := int(duration.Hours() / hoursPerMonth)
	if months == 1 {
		return "1 month ago"
	}
	return fmt.Sprintf("%d months ago", months)
}
