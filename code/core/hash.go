package core

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

// ComputeProjectHash computes a SHA256 hash of all files in a project directory.
// Files are sorted by relative path for deterministic results.
// Symlinks are skipped (not followed) to avoid security issues and infinite loops.
// Non-regular files (devices, sockets, pipes) are skipped.
func ComputeProjectHash(projectPath string) (string, error) {
	var fileHashes []fileHashEntry
	var fileCount int

	err := filepath.WalkDir(projectPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing %s: %w", path, err)
		}

		// Skip symlinks entirely to avoid security issues and infinite loops
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}

		// Skip non-regular files (directories, devices, sockets, pipes)
		if !d.Type().IsRegular() {
			return nil
		}

		// Get relative path for consistent hashing across machines
		relPath, err := filepath.Rel(projectPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}

		// Compute hash of this file
		hash, err := hashFile(path)
		if err != nil {
			return fmt.Errorf("failed to hash file %s: %w", relPath, err)
		}

		fileHashes = append(fileHashes, fileHashEntry{
			path: relPath,
			hash: hash,
		})
		fileCount++

		return nil
	})

	if err != nil {
		return "", err
	}

	// Error on empty directories to prevent masking data loss
	if fileCount == 0 {
		return "", fmt.Errorf("project directory is empty or contains no regular files: %s", projectPath)
	}

	// Sort by path for deterministic results
	sort.Slice(fileHashes, func(i, j int) bool {
		return fileHashes[i].path < fileHashes[j].path
	})

	// Combine all file hashes into project hash
	projectHasher := sha256.New()
	for _, fh := range fileHashes {
		// Include path in hash to detect renames
		projectHasher.Write([]byte(fh.path))
		projectHasher.Write([]byte{0}) // null separator
		projectHasher.Write(fh.hash)
	}

	return hex.EncodeToString(projectHasher.Sum(nil)), nil
}

type fileHashEntry struct {
	path string
	hash []byte
}

// hashFile computes the SHA256 hash of a single file
func hashFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	h := sha256.New()
	// io.Copy is memory-efficient for large files
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}
