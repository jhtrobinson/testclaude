package cli

import (
	"fmt"

	"github.com/jamespark/parkr/core"
)

// PruneOptions contains configuration for the prune command
type PruneOptions struct {
	TargetSize  string // Human-readable target size (e.g., "10G", "500M")
	Execute     bool   // If true, actually delete; if false, dry-run
	Interactive bool   // If true, allow user to interactively select projects
	NoHash      bool   // Use mtime verification instead of hash
	Force       bool   // Skip verification entirely (with warning)
}

// PruneCmd executes the prune command
func PruneCmd(opts PruneOptions) error {
	// Parse the target size
	targetBytes, err := core.ParseSize(opts.TargetSize)
	if err != nil {
		return fmt.Errorf("invalid size: %w", err)
	}

	// Load state
	sm := core.NewStateManager()
	state, err := sm.Load()
	if err != nil {
		return err
	}

	// Create prune options
	pruneOpts := core.PruneOptions{
		TargetBytes: targetBytes,
		Execute:     opts.Execute,
		NoHash:      opts.NoHash,
		Force:       opts.Force,
	}

	// Select candidates
	result, err := core.SelectPruneCandidates(state, targetBytes, pruneOpts)
	if err != nil {
		return err
	}

	// Print warnings first
	for _, warning := range result.Warnings {
		fmt.Println(warning)
		fmt.Println()
	}

	// Handle edge cases
	if result.NoCandidates {
		if opts.Force {
			fmt.Println("No projects currently checked out.")
		} else {
			fmt.Println("No safe candidates available for pruning.")
			fmt.Println("All grabbed projects have uncommitted changes or have never been parked.")
		}
		return nil
	}

	// Interactive mode
	if opts.Interactive {
		return runInteractiveMode(state, result, pruneOpts)
	}

	if !opts.Execute {
		// Dry-run mode
		return outputDryRun(result)
	}

	// Execute mode
	return executeAndReport(state, result, pruneOpts)
}

// outputDryRun displays what would be deleted without actually deleting
func outputDryRun(result *core.PruneResult) error {
	fmt.Println("DRY-RUN: The following projects would be deleted:")
	fmt.Println()

	for i, project := range result.SelectedProjects {
		sizeStr := core.FormatSize(project.LocalSize)
		fmt.Printf("%d. %s (%s)\n", i+1, project.Name, sizeStr)
	}

	fmt.Println()
	fmt.Printf("Total to free: %s (target: %s)\n",
		core.FormatSize(result.TotalSelected),
		core.FormatSize(result.TargetBytes))

	if result.InsufficientSpace {
		fmt.Println()
		fmt.Printf("WARNING: Only %s available for pruning.\n", core.FormatSize(result.TotalSelected))
		fmt.Printf("Cannot reach target of %s.\n", core.FormatSize(result.TargetBytes))
	}

	fmt.Println()
	fmt.Println("Run with --exec to actually delete.")

	return nil
}

// executeAndReport executes the prune operation and reports progress
func executeAndReport(state *core.State, result *core.PruneResult, opts core.PruneOptions) error {
	fmt.Println("Deleting projects...")
	fmt.Println()

	// Progress callback
	progressFn := func(project core.ProjectReport, success bool, freed int64) {
		if success {
			fmt.Printf("Deleting %s... %s (freed %s)\n", project.Name, SymbolCheck, core.FormatSize(freed))
		} else {
			fmt.Printf("Deleting %s... %s (failed)\n", project.Name, SymbolCross)
		}
	}

	// Execute the prune
	err := core.ExecutePrune(state, result, opts, progressFn)
	if err != nil {
		return err
	}

	fmt.Println()

	// Report results
	if len(result.Deleted) > 0 {
		fmt.Printf("Successfully freed %s\n", core.FormatSize(result.TotalFreed))
	}

	if len(result.FailedDeletions) > 0 {
		fmt.Println()
		fmt.Printf("%s Failed to delete %d project(s):\n", SymbolWarning, len(result.FailedDeletions))
		for _, p := range result.FailedDeletions {
			fmt.Printf("  - %s\n", p.Name)
		}
	}

	if result.InsufficientSpace && result.TotalFreed < result.TargetBytes {
		fmt.Println()
		fmt.Printf("Note: Only freed %s of target %s\n",
			core.FormatSize(result.TotalFreed),
			core.FormatSize(result.TargetBytes))
	}

	return nil
}

// runInteractiveMode runs the interactive selection mode for pruning
func runInteractiveMode(state *core.State, result *core.PruneResult, opts core.PruneOptions) error {
	// Run interactive selection
	selector, err := core.RunInteractiveSelection(result.Candidates, result.TargetBytes)
	if err != nil {
		return fmt.Errorf("interactive selection failed: %w", err)
	}

	// Check if user quit without confirming
	if selector.WasQuit() {
		fmt.Println("Selection cancelled. No projects deleted.")
		return nil
	}

	// Get selected candidates
	selectedCandidates := selector.GetSelected()
	if len(selectedCandidates) == 0 {
		fmt.Println("No projects selected. Nothing to delete.")
		return nil
	}

	// Update result with user selections
	result.SelectedProjects = make([]core.ProjectReport, 0)
	result.TotalSelected = 0
	for _, c := range selectedCandidates {
		result.SelectedProjects = append(result.SelectedProjects, c.ProjectReport)
		result.TotalSelected += c.LocalSize
	}

	// Show confirmation
	fmt.Printf("\nYou selected %d project(s) to delete:\n", len(selectedCandidates))
	for i, c := range selectedCandidates {
		fmt.Printf("%d. %s (%s)\n", i+1, c.Name, core.FormatSize(c.LocalSize))
	}
	fmt.Printf("\nTotal to free: %s\n", core.FormatSize(result.TotalSelected))

	// Ask for final confirmation
	fmt.Print("\nProceed with deletion? [y/N]: ")
	var response string
	_, err = fmt.Scanln(&response)
	if err != nil || (response != "y" && response != "Y" && response != "yes" && response != "Yes") {
		fmt.Println("Deletion cancelled.")
		return nil
	}

	// Execute the deletion
	return executeAndReport(state, result, opts)
}
