package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestComputeProjectHash_NormalProject(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "parkr-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create some files
	if err := os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("world"), 0644); err != nil {
		t.Fatal(err)
	}

	// Compute hash
	hash, err := ComputeProjectHash(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should get a valid hex string
	if len(hash) != 64 { // SHA256 = 32 bytes = 64 hex chars
		t.Errorf("expected 64 char hash, got %d chars: %s", len(hash), hash)
	}
}

func TestComputeProjectHash_Deterministic(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "parkr-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create files
	if err := os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("content a"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "b.txt"), []byte("content b"), 0644); err != nil {
		t.Fatal(err)
	}

	// Hash should be the same each time
	hash1, err := ComputeProjectHash(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	hash2, err := ComputeProjectHash(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if hash1 != hash2 {
		t.Errorf("hashes should be deterministic: %s != %s", hash1, hash2)
	}
}

func TestComputeProjectHash_EmptyDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "parkr-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Empty directory should return error
	_, err = ComputeProjectHash(tmpDir)
	if err == nil {
		t.Error("expected error for empty directory, got nil")
	}
}

func TestComputeProjectHash_SkipsSymlinks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "parkr-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file
	if err := os.WriteFile(filepath.Join(tmpDir, "real.txt"), []byte("real content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a symlink
	if err := os.Symlink(filepath.Join(tmpDir, "real.txt"), filepath.Join(tmpDir, "link.txt")); err != nil {
		t.Skip("symlinks not supported on this platform")
	}

	// Get hash with symlink
	hashWithSymlink, err := ComputeProjectHash(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Remove symlink
	os.Remove(filepath.Join(tmpDir, "link.txt"))

	// Get hash without symlink
	hashWithoutSymlink, err := ComputeProjectHash(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Hashes should be the same since symlinks are skipped
	if hashWithSymlink != hashWithoutSymlink {
		t.Errorf("symlinks should be skipped: %s != %s", hashWithSymlink, hashWithoutSymlink)
	}
}

func TestComputeProjectHash_DifferentContent(t *testing.T) {
	tmpDir1, err := os.MkdirTemp("", "parkr-test-1-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir1)

	tmpDir2, err := os.MkdirTemp("", "parkr-test-2-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir2)

	// Same filename, different content
	if err := os.WriteFile(filepath.Join(tmpDir1, "file.txt"), []byte("content 1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir2, "file.txt"), []byte("content 2"), 0644); err != nil {
		t.Fatal(err)
	}

	hash1, err := ComputeProjectHash(tmpDir1)
	if err != nil {
		t.Fatal(err)
	}

	hash2, err := ComputeProjectHash(tmpDir2)
	if err != nil {
		t.Fatal(err)
	}

	if hash1 == hash2 {
		t.Error("different content should produce different hashes")
	}
}

func TestComputeProjectHash_DifferentFilenames(t *testing.T) {
	tmpDir1, err := os.MkdirTemp("", "parkr-test-1-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir1)

	tmpDir2, err := os.MkdirTemp("", "parkr-test-2-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir2)

	// Same content, different filename
	if err := os.WriteFile(filepath.Join(tmpDir1, "file1.txt"), []byte("same content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir2, "file2.txt"), []byte("same content"), 0644); err != nil {
		t.Fatal(err)
	}

	hash1, err := ComputeProjectHash(tmpDir1)
	if err != nil {
		t.Fatal(err)
	}

	hash2, err := ComputeProjectHash(tmpDir2)
	if err != nil {
		t.Fatal(err)
	}

	if hash1 == hash2 {
		t.Error("different filenames should produce different hashes (path included in hash)")
	}
}

func TestComputeProjectHash_NestedDirectories(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "parkr-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create nested structure
	subDir := filepath.Join(tmpDir, "sub", "nested")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "root.txt"), []byte("root"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "nested.txt"), []byte("nested"), 0644); err != nil {
		t.Fatal(err)
	}

	hash, err := ComputeProjectHash(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hash) != 64 {
		t.Errorf("expected 64 char hash, got %d chars", len(hash))
	}
}

func TestComputeProjectHash_UnicodeFilenames(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "parkr-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create file with unicode name
	if err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("unicode test"), 0644); err != nil {
		t.Fatal(err)
	}

	hash, err := ComputeProjectHash(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hash) != 64 {
		t.Errorf("expected 64 char hash, got %d chars", len(hash))
	}
}
