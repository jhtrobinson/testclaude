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
		err = cli.InitCmd()

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
			fmt.Fprintln(os.Stderr, "Usage: parkr park <project>")
			os.Exit(2)
		}
		err = cli.ParkCmd(os.Args[2])

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
			fmt.Fprintln(os.Stderr, "Usage: parkr remove <project> [--archive] [--local] [--everywhere] [--confirm]")
			os.Exit(2)
		}
		projectName := os.Args[2]
		archiveOnly := false
		localOnly := false
		everywhere := false
		confirm := false

		for i := 3; i < len(os.Args); i++ {
			switch os.Args[i] {
			case "--archive":
				archiveOnly = true
			case "--local":
				localOnly = true
			case "--everywhere":
				everywhere = true
			case "--confirm":
				confirm = true
			default:
				fmt.Fprintf(os.Stderr, "Error: unknown option '%s'\n", os.Args[i])
				os.Exit(2)
			}
		}

		err = cli.RemoveCmd(projectName, archiveOnly, localOnly, everywhere, confirm)

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
	fmt.Println("  init              Initialize parkr state file")
	fmt.Println("  list [category]   List all projects in archive")
	fmt.Println("  grab <project>    Copy project from archive to local")
	fmt.Println("  park <project>    Sync local changes back to archive")
	fmt.Println("  rm <project>      Remove local copy (keeps archive)")
	fmt.Println("                    Options: --no-hash, --force")
	fmt.Println("  remove <project>  Remove project from archive")
	fmt.Println("                    Options: --archive, --local, --everywhere, --confirm")
	fmt.Println("  help              Show this help message")
}
