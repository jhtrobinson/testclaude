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

// ComputeProjectHash calculates SHA256 hash for a project directory
// It walks the directory in sorted order for consistency:
// For each file: hash = sha256(relative_path + file_content)
// Project hash = sha256(concatenate all file_hashes)
func ComputeProjectHash(projectPath string) (string, error) {
	var fileHashes [][]byte

	// Collect all files first
	var files []string
	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, err := filepath.Rel(projectPath, path)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to walk directory: %w", err)
	}

	// Sort files for consistent ordering
	sort.Strings(files)

	// Compute hash for each file
	for _, relPath := range files {
		fullPath := filepath.Join(projectPath, relPath)
		fileHash, err := computeFileHash(relPath, fullPath)
		if err != nil {
			return "", fmt.Errorf("failed to hash file %s: %w", relPath, err)
		}
		fileHashes = append(fileHashes, fileHash)
	}

	// Compute project hash from all file hashes
	projectHasher := sha256.New()
	for _, fh := range fileHashes {
		projectHasher.Write(fh)
	}

	return hex.EncodeToString(projectHasher.Sum(nil)), nil
}

// computeFileHash computes sha256(relative_path + file_content)
func computeFileHash(relPath string, fullPath string) ([]byte, error) {
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	hasher := sha256.New()
	// Write relative path first
	hasher.Write([]byte(relPath))

	// Then write file content
	if _, err := io.Copy(hasher, file); err != nil {
		return nil, err
	}

	return hasher.Sum(nil), nil
}
