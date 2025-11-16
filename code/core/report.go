package core

import (
	"os"
	"sort"
	"time"
)

// ProjectReport contains information about a grabbed project for reporting
type ProjectReport struct {
	Name          string    `json:"name"`
	LocalPath     string    `json:"local_path"`
	LocalSize     int64     `json:"local_size"`
	LastModified  time.Time `json:"last_modified"`
	LastParkAt    time.Time `json:"last_park_at"`
	NeverParked   bool      `json:"never_parked"`
	IsSafeDelete  bool      `json:"is_safe_delete"`
	Status        string    `json:"status"`
	NoHashMode    bool      `json:"no_hash_mode"`
}

// ReportSummary contains overall report statistics
type ReportSummary struct {
	TotalProjects     int              `json:"total_projects"`
	TotalSize         int64            `json:"total_size"`
	SafeToDelete      int              `json:"safe_to_delete"`
	RecoverableSpace  int64            `json:"recoverable_space"`
	Projects          []ProjectReport  `json:"projects"`
	Candidates        []ProjectReport  `json:"candidates"`
}

// SortField defines how to sort projects
type SortField string

const (
	SortBySize     SortField = "size"
	SortByModified SortField = "modified"
	SortByName     SortField = "name"
)

// GenerateReport generates a report of all grabbed projects
func GenerateReport(state *State, recomputeHashes bool) (*ReportSummary, error) {
	summary := &ReportSummary{
		Projects:   make([]ProjectReport, 0),
		Candidates: make([]ProjectReport, 0),
	}

	for name, project := range state.Projects {
		if !project.IsGrabbed {
			continue
		}

		report := ProjectReport{
			Name:       name,
			LocalPath:  project.LocalPath,
			NoHashMode: project.NoHashMode,
		}

		// Get local size
		if _, err := os.Stat(project.LocalPath); err == nil {
			if size, err := GetDirSize(project.LocalPath); err == nil {
				report.LocalSize = size
			}

			// Get last modified time
			if newest, err := GetNewestMtime(project.LocalPath); err == nil && newest != nil {
				report.LastModified = (*newest).ModTime()
			}
		}

		// Set last park time
		if project.LastParkAt != nil {
			report.LastParkAt = *project.LastParkAt
			report.NeverParked = false
		} else {
			report.NeverParked = true
		}

		// Determine safety status
		report.IsSafeDelete, report.Status = determineSafetyStatus(project, report.LastModified, recomputeHashes)

		summary.Projects = append(summary.Projects, report)
		summary.TotalSize += report.LocalSize

		if report.IsSafeDelete {
			summary.SafeToDelete++
			summary.RecoverableSpace += report.LocalSize
			summary.Candidates = append(summary.Candidates, report)
		}
	}

	summary.TotalProjects = len(summary.Projects)

	// Sort candidates by last modified (oldest first)
	sort.Slice(summary.Candidates, func(i, j int) bool {
		return summary.Candidates[i].LastModified.Before(summary.Candidates[j].LastModified)
	})

	return summary, nil
}

// determineSafetyStatus determines if a project is safe to delete
func determineSafetyStatus(project *Project, lastModified time.Time, recomputeHashes bool) (bool, string) {
	// Never parked - not safe
	if project.LastParkAt == nil {
		return false, "Never checked in"
	}

	// If recomputing hashes
	if recomputeHashes && !project.NoHashMode {
		// Compute current local hash
		currentHash, err := ComputeProjectHash(project.LocalPath)
		if err != nil {
			return false, "Error computing hash"
		}

		// Compare with stored local hash from last park
		if project.LocalContentHash == nil {
			// Missing hash data for non-NoHashMode project - treat as unsafe
			return false, "Missing hash data"
		}
		if currentHash != *project.LocalContentHash {
			return false, "Has uncommitted work"
		}
	} else {
		// Use mtime-based check
		if project.LastParkMtime != nil {
			if lastModified.After(*project.LastParkMtime) {
				return false, "Has uncommitted work"
			}
		} else {
			// Fallback to comparing with LastParkAt
			if lastModified.After(*project.LastParkAt) {
				return false, "Has uncommitted work"
			}
		}
	}

	return true, "Safe to delete"
}

// SortProjects sorts a slice of ProjectReport by the given field
func SortProjects(projects []ProjectReport, field SortField) {
	switch field {
	case SortBySize:
		sort.Slice(projects, func(i, j int) bool {
			return projects[i].LocalSize > projects[j].LocalSize // Largest first
		})
	case SortByName:
		sort.Slice(projects, func(i, j int) bool {
			return projects[i].Name < projects[j].Name
		})
	case SortByModified:
		fallthrough
	default:
		sort.Slice(projects, func(i, j int) bool {
			return projects[i].LastModified.Before(projects[j].LastModified) // Oldest first
		})
	}
}

// FilterCandidates returns only the projects that are safe to delete
func FilterCandidates(projects []ProjectReport) []ProjectReport {
	candidates := make([]ProjectReport, 0)
	for _, p := range projects {
		if p.IsSafeDelete {
			candidates = append(candidates, p)
		}
	}
	return candidates
}
