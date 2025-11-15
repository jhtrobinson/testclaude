package cli

import (
	"fmt"

	"github.com/jamespark/parkr/core"
)

// InitCmd initializes parkr state file
func InitCmd() error {
	sm := core.NewStateManager()

	if sm.Exists() {
		return fmt.Errorf("state file already exists at %s", sm.StatePath())
	}

	if err := sm.CreateDefault(); err != nil {
		return fmt.Errorf("failed to create state file: %w", err)
	}

	fmt.Printf("Initialized parkr state file at %s\n", sm.StatePath())
	return nil
}
