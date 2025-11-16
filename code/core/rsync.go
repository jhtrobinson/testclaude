package core

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// Rsync performs rsync from source to destination
func Rsync(src, dst string) error {
	// Ensure trailing slash on source to copy contents
	if src[len(src)-1] != '/' {
		src = src + "/"
	}

	// Check if rsync is available
	if _, err := exec.LookPath("rsync"); err != nil {
		// Fall back to simple copy for environments without rsync
		return simpleCopy(src, dst)
	}

	cmd := exec.Command("rsync", "-av", "--delete", src, dst)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rsync failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// simpleCopy provides a basic file copy fallback when rsync is not available
func simpleCopy(src, dst string) error {
	// Remove trailing slash for filepath operations
	src = filepath.Clean(src)

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return err
		}

		return os.Chmod(dstPath, info.Mode())
	})
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
