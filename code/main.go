package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jamespark/parkr/cli"
	"github.com/jamespark/parkr/core"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}

	command := os.Args[1]
	var err error

	switch command {
	case "init":
		archiveRoot := ""
		if len(os.Args) > 2 {
			archiveRoot = os.Args[2]
		}
		err = cli.InitCmd(archiveRoot)

	case "add":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: local path required")
			fmt.Fprintln(os.Stderr, "Usage: parkr add <local-path> [category] [--move]")
			os.Exit(2)
		}
		localPath := os.Args[2]
		category := ""
		move := false

		// Parse remaining arguments
		for i := 3; i < len(os.Args); i++ {
			arg := os.Args[i]
			if arg == "--move" {
				move = true
			} else if arg[0] != '-' && category == "" {
				category = arg
			} else if arg[0] == '-' {
				fmt.Fprintf(os.Stderr, "Error: unknown option '%s'\n", arg)
				os.Exit(2)
			} else {
				fmt.Fprintf(os.Stderr, "Error: unexpected argument '%s'\n", arg)
				os.Exit(2)
			}
		}

		err = cli.AddCmd(localPath, category, move)

	case "list", "ls":
		category := ""
		if len(os.Args) > 2 {
			category = os.Args[2]
		}
		err = cli.ListCmd(category)

	case "grab", "checkout":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: project name required")
			fmt.Fprintln(os.Stderr, "Usage: parkr grab <project> [--force] [--to <path>]")
			os.Exit(2)
		}
		projectName := os.Args[2]
		force := false
		customPath := ""

		for i := 3; i < len(os.Args); i++ {
			switch os.Args[i] {
			case "--force":
				force = true
			case "--to":
				if i+1 >= len(os.Args) {
					fmt.Fprintln(os.Stderr, "Error: --to requires a path argument")
					os.Exit(2)
				}
				customPath = os.Args[i+1]
				i++ // Skip next argument as it's the path
			default:
				fmt.Fprintf(os.Stderr, "Error: unknown option '%s'\n", os.Args[i])
				os.Exit(2)
			}
		}

		err = cli.GrabCmd(projectName, force, customPath)

	case "park":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: project name required")
			fmt.Fprintln(os.Stderr, "Usage: parkr park <project> [--no-hash]")
			os.Exit(2)
		}
		projectName := os.Args[2]
		noHash := false

		for i := 3; i < len(os.Args); i++ {
			switch os.Args[i] {
			case "--no-hash":
				noHash = true
			default:
				fmt.Fprintf(os.Stderr, "Error: unknown option '%s'\n", os.Args[i])
				os.Exit(2)
			}
		}

		err = cli.ParkCmd(projectName, noHash)

	case "rm":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: project name required")
			fmt.Fprintln(os.Stderr, "Usage: parkr rm <project> [--no-hash] [--force]")
			os.Exit(2)
		}
		projectName := os.Args[2]
		noHash := false
		force := false

		for i := 3; i < len(os.Args); i++ {
			switch os.Args[i] {
			case "--no-hash":
				noHash = true
			case "--force":
				force = true
			default:
				fmt.Fprintf(os.Stderr, "Error: unknown option '%s'\n", os.Args[i])
				os.Exit(2)
			}
		}

		err = cli.RmCmd(projectName, noHash, force)

	case "remove":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: project name required")
			fmt.Fprintln(os.Stderr, "Usage: parkr remove <project> [--archive] [--yes]")
			os.Exit(2)
		}
		projectName := os.Args[2]
		archive := false
		yes := false

		for i := 3; i < len(os.Args); i++ {
			switch os.Args[i] {
			case "--archive":
				archive = true
			case "--yes", "-y":
				yes = true
			default:
				fmt.Fprintf(os.Stderr, "Error: unknown option '%s'\n", os.Args[i])
				os.Exit(2)
			}
		}

		err = cli.RemoveCmd(projectName, false, archive, yes)
		if err != nil {
			errStr := err.Error()
			// Map error types to exit codes
			if strings.Contains(errStr, "state file error") {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(4)
			} else if strings.Contains(errStr, "archive not accessible") {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(3)
			} else if strings.Contains(errStr, "not found") || strings.Contains(errStr, "does not exist") {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "status":
		err = cli.StatusCmd()

	case "info":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: project name required")
			fmt.Fprintln(os.Stderr, "Usage: parkr info <project>")
			os.Exit(2)
		}
		err = cli.InfoCmd(os.Args[2])

	case "local":
		unmanagedOnly := false
		for i := 2; i < len(os.Args); i++ {
			if os.Args[i] == "--unmanaged" {
				unmanagedOnly = true
			} else {
				fmt.Fprintf(os.Stderr, "Error: unknown option '%s'\n", os.Args[i])
				os.Exit(2)
			}
		}
		err = cli.LocalCmd(unmanagedOnly)

	case "verify":
		err = cli.VerifyCmd()

	case "config":
		err = cli.ConfigCmd()

	case "report":
		opts := cli.ReportOptions{
			CandidatesOnly:  false,
			RecomputeHashes: false,
			SortBy:          core.SortByModified,
			JSONOutput:      false,
		}

		for i := 2; i < len(os.Args); i++ {
			switch os.Args[i] {
			case "--candidates":
				opts.CandidatesOnly = true
			case "--recompute-hashes":
				opts.RecomputeHashes = true
			case "--json":
				opts.JSONOutput = true
			case "--sort":
				if i+1 >= len(os.Args) {
					fmt.Fprintln(os.Stderr, "Error: --sort requires a field argument (size|modified|name)")
					os.Exit(2)
				}
				sortField := os.Args[i+1]
				switch sortField {
				case "size":
					opts.SortBy = core.SortBySize
				case "modified":
					opts.SortBy = core.SortByModified
				case "name":
					opts.SortBy = core.SortByName
				default:
					fmt.Fprintf(os.Stderr, "Error: invalid sort field '%s' (use size|modified|name)\n", sortField)
					os.Exit(2)
				}
				i++ // Skip next argument as it's the sort field
			default:
				fmt.Fprintf(os.Stderr, "Error: unknown option '%s'\n", os.Args[i])
				os.Exit(2)
			}
		}

		err = cli.ReportCmd(opts)

	case "help", "--help", "-h":
		printUsage()

	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command '%s'\n", command)
		printUsage()
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("parkr - Project archive manager")
	fmt.Println()
	fmt.Println("Usage: parkr <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  init [path]       Initialize parkr (path, $PARKR_ARCHIVE_ROOT, or prompt)")
	fmt.Println("  add <path> [cat]  Add local project to archive")
	fmt.Println("                    Options: --move")
	fmt.Println("  list [category]   List all projects in archive")
	fmt.Println("  grab <project>    Copy project from archive to local")
	fmt.Println("                    Options: --force, --to <path>")
	fmt.Println("  park <project>    Sync local changes back to archive")
	fmt.Println("                    Options: --no-hash")
	fmt.Println("  rm <project>      Remove local copy (keeps archive)")
	fmt.Println("                    Options: --no-hash, --force")
	fmt.Println("  remove <project>  Remove project from state (and optionally archive)")
	fmt.Println("                    Options: --archive, --yes")
	fmt.Println()
	fmt.Println("Status and Information:")
	fmt.Println("  status            Show all grabbed projects with sync status")
	fmt.Println("  info <project>    Show detailed project information")
	fmt.Println("  local             Show all local projects")
	fmt.Println("                    Options: --unmanaged")
	fmt.Println("  report            Show disk usage report for grabbed projects")
	fmt.Println("                    Options: --candidates, --recompute-hashes, --sort <field>, --json")
	fmt.Println("  verify            Check state file consistency")
	fmt.Println("  config            Show current configuration")
	fmt.Println()
	fmt.Println("  help              Show this help message")
}
