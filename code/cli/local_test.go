package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jamespark/parkr/core"
)

func TestGetLocalDirectories(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	dirs := getLocalDirectories()

	if len(dirs) != 3 {
		t.Errorf("getLocalDirectories() returned %d directories, want 3", len(dirs))
	}

	expectedDirs := []string{
		filepath.Join(homeDir, "code"),
		filepath.Join(homeDir, "PycharmProjects"),
		filepath.Join(homeDir, "RStudioProjects"),
	}

	for i, expected := range expectedDirs {
		if i >= len(dirs) {
			t.Errorf("Missing directory at index %d", i)
			continue
		}
		if dirs[i] != expected {
			t.Errorf("getLocalDirectories()[%d] = %q, want %q", i, dirs[i], expected)
		}
	}
}

func TestGetLocalDirectoriesFromState(t *testing.T) {
	t.Run("uses state directories when configured", func(t *testing.T) {
		customDirs := []string{"/custom/dir1", "/custom/dir2"}
		state := &core.State{
			LocalDirectories: customDirs,
		}

		result := getLocalDirectoriesFromState(state)

		if len(result) != len(customDirs) {
			t.Errorf("getLocalDirectoriesFromState() returned %d directories, want %d", len(result), len(customDirs))
		}

		for i, expected := range customDirs {
			if result[i] != expected {
				t.Errorf("getLocalDirectoriesFromState()[%d] = %q, want %q", i, result[i], expected)
			}
		}
	})

	t.Run("uses defaults when state has empty directories", func(t *testing.T) {
		state := &core.State{
			LocalDirectories: []string{},
		}

		result := getLocalDirectoriesFromState(state)
		defaults := getLocalDirectories()

		if len(result) != len(defaults) {
			t.Errorf("getLocalDirectoriesFromState() returned %d directories, want %d", len(result), len(defaults))
		}

		for i, expected := range defaults {
			if result[i] != expected {
				t.Errorf("getLocalDirectoriesFromState()[%d] = %q, want %q", i, result[i], expected)
			}
		}
	})

	t.Run("uses defaults when state is nil", func(t *testing.T) {
		result := getLocalDirectoriesFromState(nil)
		defaults := getLocalDirectories()

		if len(result) != len(defaults) {
			t.Errorf("getLocalDirectoriesFromState(nil) returned %d directories, want %d", len(result), len(defaults))
		}
	})

	t.Run("uses defaults when LocalDirectories is nil", func(t *testing.T) {
		state := &core.State{
			LocalDirectories: nil,
		}

		result := getLocalDirectoriesFromState(state)
		defaults := getLocalDirectories()

		if len(result) != len(defaults) {
			t.Errorf("getLocalDirectoriesFromState() returned %d directories, want %d", len(result), len(defaults))
		}
	})
}
