package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jamespark/parkr/core"
)

// ReportOptions contains configuration for the report command
type ReportOptions struct {
	CandidatesOnly  bool
	RecomputeHashes bool
	SortBy          core.SortField
	JSONOutput      bool
}

// ReportCmd generates a disk usage report for grabbed projects
func ReportCmd(opts ReportOptions) error {
	sm := core.NewStateManager()
	state, err := sm.Load()
	if err != nil {
		return err
	}

	// Generate the report
	summary, err := core.GenerateReport(state, opts.RecomputeHashes)
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	// Handle empty state
	if summary.TotalProjects == 0 {
		if opts.JSONOutput {
			return outputJSON(summary, opts.CandidatesOnly)
		}
		fmt.Println("No projects currently checked out.")
		return nil
	}

	// Sort projects
	core.SortProjects(summary.Projects, opts.SortBy)

	// Filter to candidates only if requested
	projectsToShow := summary.Projects
	if opts.CandidatesOnly {
		projectsToShow = core.FilterCandidates(summary.Projects)
	}

	// Output format
	if opts.JSONOutput {
		return outputJSON(summary, opts.CandidatesOnly)
	}

	return outputHumanReadable(summary, projectsToShow, opts.CandidatesOnly)
}

// outputJSON outputs the report as JSON
func outputJSON(summary *core.ReportSummary, candidatesOnly bool) error {
	var output interface{}
	if candidatesOnly {
		// Output only candidates when --candidates flag is used
		output = struct {
			SafeToDelete     int                   `json:"safe_to_delete"`
			RecoverableSpace int64                 `json:"recoverable_space"`
			Candidates       []core.ProjectReport  `json:"candidates"`
		}{
			SafeToDelete:     summary.SafeToDelete,
			RecoverableSpace: summary.RecoverableSpace,
			Candidates:       summary.Candidates,
		}
	} else {
		output = summary
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

// outputHumanReadable outputs the report in human-readable format
func outputHumanReadable(summary *core.ReportSummary, projects []core.ProjectReport, candidatesOnly bool) error {
	// Header
	fmt.Printf("LOCAL DISK USAGE: %s\n", core.FormatSize(summary.TotalSize))
	fmt.Println()

	// Projects table
	if candidatesOnly {
		fmt.Println("PRUNING CANDIDATES:")
	} else {
		fmt.Println("CHECKED OUT PROJECTS:")
	}

	fmt.Printf("%-25s %-12s %-16s %-16s %s\n", "PROJECT", "LOCAL SIZE", "LAST MODIFIED", "LAST CHECKIN", "STATUS")
	fmt.Println(strings.Repeat("-", 95))

	for _, p := range projects {
		sizeStr := core.FormatSize(p.LocalSize)
		modifiedStr := formatTimeAgo(p.LastModified)

		checkinStr := "never"
		if !p.NeverParked {
			checkinStr = formatTimeAgo(p.LastParkAt)
		}

		// Determine status display
		var statusStr string
		if p.IsSafeDelete {
			statusStr = SymbolCheck + " " + p.Status
		} else if p.NeverParked {
			statusStr = SymbolCross + " " + p.Status
		} else {
			statusStr = SymbolWarning + " " + p.Status
		}

		fmt.Printf("%-25s %-12s %-16s %-16s %s\n", p.Name, sizeStr, modifiedStr, checkinStr, statusStr)
	}

	fmt.Println()

	// Summary section (only for full report)
	if !candidatesOnly {
		if len(summary.Candidates) > 0 {
			fmt.Println("PRUNING CANDIDATES (safe to delete, oldest first):")
			for i, c := range summary.Candidates {
				fmt.Printf("%d. %s (%s) - last modified %s\n", i+1, c.Name, core.FormatSize(c.LocalSize), formatTimeAgo(c.LastModified))
			}
			fmt.Println()
		}

		fmt.Printf("TOTAL RECOVERABLE: %s\n", core.FormatSize(summary.RecoverableSpace))
	} else if len(projects) == 0 {
		fmt.Println("No safe candidates found.")
	} else {
		fmt.Printf("TOTAL RECOVERABLE: %s\n", core.FormatSize(summary.RecoverableSpace))
	}

	return nil
}
