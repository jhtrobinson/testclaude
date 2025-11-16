package main

import (
	"fmt"
	"os"

	"github.com/jamespark/parkr/cli"
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
			fmt.Fprintln(os.Stderr, "Usage: parkr add <local-path> [category]")
			os.Exit(2)
		}
		localPath := os.Args[2]
		category := ""
		if len(os.Args) > 3 {
			category = os.Args[3]
		}
		err = cli.AddCmd(localPath, category)

	case "list", "ls":
		category := ""
		if len(os.Args) > 2 {
			category = os.Args[2]
		}
		err = cli.ListCmd(category)

	case "grab", "checkout":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: project name required")
			fmt.Fprintln(os.Stderr, "Usage: parkr grab <project>")
			os.Exit(2)
		}
		err = cli.GrabCmd(os.Args[2])

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
	fmt.Println("  list [category]   List all projects in archive")
	fmt.Println("  grab <project>    Copy project from archive to local")
	fmt.Println("  park <project>    Sync local changes back to archive")
	fmt.Println("                    Options: --no-hash")
	fmt.Println("  rm <project>      Remove local copy (keeps archive)")
	fmt.Println("                    Options: --no-hash, --force")
	fmt.Println()
	fmt.Println("Status and Information:")
	fmt.Println("  status            Show all grabbed projects with sync status")
	fmt.Println("  info <project>    Show detailed project information")
	fmt.Println("  local             Show all local projects")
	fmt.Println("                    Options: --unmanaged")
	fmt.Println("  verify            Check state file consistency")
	fmt.Println("  config            Show current configuration")
	fmt.Println()
	fmt.Println("  help              Show this help message")
}
