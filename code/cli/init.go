package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jamespark/parkr/core"
)

// InitCmd initializes parkr state file
func InitCmd(archiveRoot string) error {
	sm := core.NewStateManager()

	if sm.Exists() {
		return fmt.Errorf("state file already exists at %s", sm.StatePath())
	}

	// Get archive root: parameter > env var > prompt
	if archiveRoot == "" {
		archiveRoot = os.Getenv("PARKR_ARCHIVE_ROOT")
	}

	if archiveRoot == "" {
		fmt.Print("Enter archive root path: ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
		archiveRoot = strings.TrimSpace(input)
	}

	if archiveRoot == "" {
		return fmt.Errorf("archive root path is required")
	}

	if err := sm.CreateWithRoot(archiveRoot); err != nil {
		return fmt.Errorf("failed to create state file: %w", err)
	}

	fmt.Printf("Initialized parkr state file at %s\n", sm.StatePath())
	fmt.Printf("Archive root: %s\n", archiveRoot)
	return nil
}
