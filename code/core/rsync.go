package core

import (
	"fmt"
	"os/exec"
)

// Rsync performs rsync from source to destination
func Rsync(src, dst string) error {
	// Ensure trailing slash on source to copy contents
	if src[len(src)-1] != '/' {
		src = src + "/"
	}

	cmd := exec.Command("rsync", "-av", "--delete", src, dst)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rsync failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// RsyncWithProgress performs rsync with progress output
func RsyncWithProgress(src, dst string) error {
	// Ensure trailing slash on source to copy contents
	if src[len(src)-1] != '/' {
		src = src + "/"
	}

	cmd := exec.Command("rsync", "-av", "--delete", "--progress", src, dst)
	cmd.Stdout = nil // Will be displayed directly
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rsync failed: %w", err)
	}

	return nil
}
