package core

import (
	"fmt"
	"os"
	"sort"
)

// PruneOptions contains configuration for the prune operation
type PruneOptions struct {
	TargetBytes int64 // Target amount of space to free
	Execute     bool  // If true, actually delete; if false, dry-run
	NoHash      bool  // Use mtime verification instead of hash
	Force       bool  // Skip verification entirely
}

// PruneCandidate represents a project candidate for pruning
type PruneCandidate struct {
	ProjectReport
	Selected bool // Whether this project is selected for deletion
}

// PruneResult contains the result of a prune operation
type PruneResult struct {
	Candidates          []PruneCandidate
	SelectedProjects    []ProjectReport
	TotalSelected       int64
	TargetBytes         int64
	InsufficientSpace   bool
	NoCandidates        bool
	Deleted             []ProjectReport
	FailedDeletions     []ProjectReport
	TotalFreed          int64
	Warnings            []string
}

// SelectPruneCandidates selects projects to prune to reach the target size.
// Projects are selected oldest first (by last modified time).
// Returns candidates up to the target size.
func SelectPruneCandidates(state *State, targetBytes int64, opts PruneOptions) (*PruneResult, error) {
	result := &PruneResult{
		Candidates:       make([]PruneCandidate, 0),
		SelectedProjects: make([]ProjectReport, 0),
		TargetBytes:      targetBytes,
	}

	// Generate report to get all candidates
	summary, err := GenerateReport(state, false) // Don't recompute hashes during selection
	if err != nil {
		return nil, fmt.Errorf("failed to generate report: %w", err)
	}

	// Get safe candidates (already sorted by oldest first from GenerateReport)
	safeCandidates := summary.Candidates

	// If force mode, include all grabbed projects as candidates
	if opts.Force {
		result.Warnings = append(result.Warnings, "WARNING: --force skips verification. Data may be lost!")
		safeCandidates = summary.Projects
		// Sort all projects by oldest first
		sort.Slice(safeCandidates, func(i, j int) bool {
			return safeCandidates[i].LastModified.Before(safeCandidates[j].LastModified)
		})
	}

	if len(safeCandidates) == 0 {
		result.NoCandidates = true
		return result, nil
	}

	// Convert to PruneCandidates
	for _, p := range safeCandidates {
		result.Candidates = append(result.Candidates, PruneCandidate{
			ProjectReport: p,
			Selected:      false,
		})
	}

	// Select candidates until we reach the target
	var totalSelected int64
	for i := range result.Candidates {
		if totalSelected >= targetBytes {
			break
		}
		result.Candidates[i].Selected = true
		result.SelectedProjects = append(result.SelectedProjects, result.Candidates[i].ProjectReport)
		totalSelected += result.Candidates[i].LocalSize
	}

	result.TotalSelected = totalSelected

	// Check if we have insufficient space
	if totalSelected < targetBytes {
		result.InsufficientSpace = true
	}

	return result, nil
}

// newStateManagerFn allows overriding StateManager creation for testing
var newStateManagerFn = func() *StateManager {
	return NewStateManager()
}

// ExecutePrune actually deletes the selected projects
func ExecutePrune(state *State, result *PruneResult, opts PruneOptions, progressFn func(project ProjectReport, success bool, freed int64)) error {
	sm := newStateManagerFn()

	result.Deleted = make([]ProjectReport, 0)
	result.FailedDeletions = make([]ProjectReport, 0)
	result.TotalFreed = 0

	// Wrap progress callback in safe function to prevent panics
	safeProgressFn := func(project ProjectReport, success bool, freed int64) {
		if progressFn == nil {
			return
		}
		defer func() {
			if r := recover(); r != nil {
				// Log panic but don't let it stop the prune operation
				fmt.Fprintf(os.Stderr, "Warning: progress callback panicked: %v\n", r)
			}
		}()
		progressFn(project, success, freed)
	}

	for _, project := range result.SelectedProjects {
		// Get the project from state
		stateProject, exists := state.Projects[project.Name]
		if !exists {
			result.FailedDeletions = append(result.FailedDeletions, project)
			safeProgressFn(project, false, 0)
			continue
		}

		// Re-verify before deletion (unless force mode)
		if !opts.Force {
			isSafe, _ := verifyBeforeDeletion(stateProject, opts.NoHash)
			if !isSafe {
				result.FailedDeletions = append(result.FailedDeletions, project)
				safeProgressFn(project, false, 0)
				continue
			}
		}

		// Delete the project (common logic for both force and non-force modes)
		freed, err := deleteSingleProject(stateProject, project, sm, state)
		if err != nil {
			result.FailedDeletions = append(result.FailedDeletions, project)
			safeProgressFn(project, false, 0)
			continue
		}

		result.Deleted = append(result.Deleted, project)
		result.TotalFreed += freed
		safeProgressFn(project, true, freed)

		// Check if we've reached the target
		if result.TotalFreed >= result.TargetBytes {
			break
		}
	}

	return nil
}

// deleteSingleProject handles the actual deletion of a single project
// Returns the freed space and any error encountered
func deleteSingleProject(stateProject *Project, project ProjectReport, sm *StateManager, state *State) (int64, error) {
	// Get current size before deletion
	currentSize := project.LocalSize
	if newSize, err := GetDirSize(project.LocalPath); err == nil {
		currentSize = newSize
	}

	// Delete the local directory
	if err := os.RemoveAll(project.LocalPath); err != nil {
		return 0, fmt.Errorf("failed to delete directory: %w", err)
	}

	// Update state to mark as not grabbed
	stateProject.IsGrabbed = false
	stateProject.GrabbedAt = nil

	// Save state after deletion
	// Note: If this fails, the directory is already deleted but state is inconsistent
	// This is logged as a failure so the user knows to investigate
	if err := sm.Save(state); err != nil {
		return 0, fmt.Errorf("deleted directory but failed to save state: %w", err)
	}

	return currentSize, nil
}

// verifyBeforeDeletion checks if a project is still safe to delete
func verifyBeforeDeletion(project *Project, noHash bool) (bool, string) {
	// Check if project was never parked
	if project.LastParkAt == nil {
		return false, "Never checked in"
	}

	// Check if local path still exists
	if _, err := os.Stat(project.LocalPath); err != nil {
		return false, "Local path not found"
	}

	// Get current modification time
	newest, err := GetNewestMtime(project.LocalPath)
	if err != nil || newest == nil {
		return false, "Error getting modification time"
	}
	lastModified := (*newest).ModTime()

	// Use mtime-based check if noHash or if project is in no_hash_mode
	if noHash || project.NoHashMode {
		if project.LastParkMtime != nil {
			if lastModified.After(*project.LastParkMtime) {
				return false, "Has uncommitted work"
			}
		} else if project.LastParkAt != nil {
			if lastModified.After(*project.LastParkAt) {
				return false, "Has uncommitted work"
			}
		}
	} else {
		// Use hash-based check
		currentHash, err := ComputeProjectHash(project.LocalPath)
		if err != nil {
			return false, "Error computing hash"
		}

		if project.LocalContentHash == nil || currentHash != *project.LocalContentHash {
			return false, "Has uncommitted work"
		}
	}

	return true, "Safe to delete"
}
