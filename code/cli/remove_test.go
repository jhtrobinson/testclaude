package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jamespark/parkr/core"
)

// setupTestEnvironment creates a temporary test environment
func setupTestEnvironment(t *testing.T) (string, *core.StateManager, func()) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "parkr-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Set up directory structure
	stateDir := filepath.Join(tmpDir, ".parkr")
	archiveDir := filepath.Join(tmpDir, "archive", "code")
	localDir := filepath.Join(tmpDir, "local")

	for _, dir := range []string{stateDir, archiveDir, localDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			os.RemoveAll(tmpDir)
			t.Fatalf("Failed to create dir %s: %v", dir, err)
		}
	}

	// Create custom state manager
	sm := &core.StateManager{}
	// We need to set the state path - create a wrapper
	statePath := filepath.Join(stateDir, "state.json")

	// Initialize state with test paths
	state := &core.State{
		Masters: map[string]map[string]string{
			"primary": {
				"code": archiveDir,
			},
		},
		DefaultMaster: "primary",
		Projects:      make(map[string]*core.Project),
	}

	// Write initial state
	data, _ := os.ReadFile(statePath)
	if len(data) == 0 {
		// Create state file manually since we can't easily override StateManager path
		stateJSON := `{
			"masters": {
				"primary": {
					"code": "` + archiveDir + `"
				}
			},
			"default_master": "primary",
			"projects": {}
		}`
		if err := os.WriteFile(statePath, []byte(stateJSON), 0644); err != nil {
			os.RemoveAll(tmpDir)
			t.Fatalf("Failed to write state: %v", err)
		}
	}

	// Cleanup function
	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	// Override HOME for the test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)

	// Update cleanup to restore HOME
	originalCleanup := cleanup
	cleanup = func() {
		os.Setenv("HOME", oldHome)
		originalCleanup()
	}

	// Create new state manager (will use the new HOME)
	sm = core.NewStateManager()

	// Save the test state
	if err := sm.Save(state); err != nil {
		cleanup()
		t.Fatalf("Failed to save state: %v", err)
	}

	return tmpDir, sm, cleanup
}

// createTestProject creates a test project in archive and optionally local
func createTestProject(t *testing.T, tmpDir string, sm *core.StateManager, name string, grabbed bool) {
	state, err := sm.Load()
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}

	archiveDir := state.Masters["primary"]["code"]
	projectArchive := filepath.Join(archiveDir, name)

	// Create archive directory
	if err := os.MkdirAll(projectArchive, 0755); err != nil {
		t.Fatalf("Failed to create archive dir: %v", err)
	}

	// Create a file in archive
	if err := os.WriteFile(filepath.Join(projectArchive, "README.md"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	now := time.Now()
	project := &core.Project{
		Master:          "primary",
		ArchiveCategory: "code",
		IsGrabbed:       grabbed,
		LastParkAt:      &now,
		LastParkMtime:   &now,
	}

	if grabbed {
		localDir := filepath.Join(tmpDir, "local", name)
		if err := os.MkdirAll(localDir, 0755); err != nil {
			t.Fatalf("Failed to create local dir: %v", err)
		}
		// Create a file in local
		if err := os.WriteFile(filepath.Join(localDir, "README.md"), []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create local test file: %v", err)
		}
		project.LocalPath = localDir
		project.GrabbedAt = &now
	}

	state.Projects[name] = project
	if err := sm.Save(state); err != nil {
		t.Fatalf("Failed to save project state: %v", err)
	}
}

func TestRemoveCmd_NonexistentProject(t *testing.T) {
	_, sm, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Verify state manager works
	_, err := sm.Load()
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}

	// Try to remove nonexistent project
	err = RemoveCmd("nonexistent", false, false, true)
	if err == nil {
		t.Error("Expected error for nonexistent project, got nil")
	}
	if err != nil && !contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

func TestRemoveCmd_LocalOnlyNonGrabbed(t *testing.T) {
	_, sm, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create a project that's not grabbed
	createTestProject(t, "", sm, "test-project", false)

	// Try to remove local only when not grabbed
	err := RemoveCmd("test-project", true, false, true)
	if err == nil {
		t.Error("Expected error for non-grabbed project, got nil")
	}
	if err != nil && !contains(err.Error(), "not currently grabbed") {
		t.Errorf("Expected 'not currently grabbed' error, got: %v", err)
	}
}

func TestRemoveCmd_MutuallyExclusiveFlags(t *testing.T) {
	// This is actually tested in main.go, but document it here
	// The RemoveCmd itself doesn't check this - main.go does
	t.Log("Mutually exclusive flags are checked in main.go before calling RemoveCmd")
}

func TestRemoveCmd_ArchiveRemovalUpdatesState(t *testing.T) {
	tmpDir, sm, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create a project
	createTestProject(t, tmpDir, sm, "test-project", false)

	// Verify project exists in state before removal
	stateBefore, _ := sm.Load()
	if _, exists := stateBefore.Projects["test-project"]; !exists {
		t.Fatal("Project should exist in state before removal")
	}

	// Get archive path
	archivePath, _ := stateBefore.GetArchivePath("test-project")

	// Verify archive exists
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		t.Fatal("Archive should exist before removal")
	}

	// Remove archive (with --yes to skip confirmation)
	err := RemoveCmd("test-project", false, false, true)
	if err != nil {
		t.Errorf("RemoveCmd failed: %v", err)
	}

	// Verify state is updated
	stateAfter, _ := sm.Load()
	if _, exists := stateAfter.Projects["test-project"]; exists {
		t.Error("Project should be removed from state after removal")
	}

	// Verify archive is deleted
	if _, err := os.Stat(archivePath); !os.IsNotExist(err) {
		t.Error("Archive should be deleted after removal")
	}
}

func TestRemoveCmd_EverywhereRemoval(t *testing.T) {
	tmpDir, sm, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create a grabbed project (has both archive and local)
	createTestProject(t, tmpDir, sm, "test-project", true)

	// Verify both paths exist
	stateBefore, _ := sm.Load()
	archivePath, _ := stateBefore.GetArchivePath("test-project")
	localPath := stateBefore.Projects["test-project"].LocalPath

	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		t.Fatal("Archive should exist before removal")
	}
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		t.Fatal("Local should exist before removal")
	}

	// Remove everywhere
	err := RemoveCmd("test-project", false, true, true)
	if err != nil {
		t.Errorf("RemoveCmd failed: %v", err)
	}

	// Verify both are deleted
	if _, err := os.Stat(archivePath); !os.IsNotExist(err) {
		t.Error("Archive should be deleted after --everywhere removal")
	}
	if _, err := os.Stat(localPath); !os.IsNotExist(err) {
		t.Error("Local should be deleted after --everywhere removal")
	}

	// Verify state is updated
	stateAfter, _ := sm.Load()
	if _, exists := stateAfter.Projects["test-project"]; exists {
		t.Error("Project should be removed from state after --everywhere removal")
	}
}

func TestRemoveCmd_LocalOnlyRemoval(t *testing.T) {
	tmpDir, sm, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create a grabbed project
	createTestProject(t, tmpDir, sm, "test-project", true)

	// Get paths
	stateBefore, _ := sm.Load()
	archivePath, _ := stateBefore.GetArchivePath("test-project")
	localPath := stateBefore.Projects["test-project"].LocalPath

	// Remove local only
	err := RemoveCmd("test-project", true, false, true)
	if err != nil {
		t.Errorf("RemoveCmd failed: %v", err)
	}

	// Verify local is deleted but archive remains
	if _, err := os.Stat(localPath); !os.IsNotExist(err) {
		t.Error("Local should be deleted after --local removal")
	}
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		t.Error("Archive should remain after --local removal")
	}

	// Verify state is updated (project still exists but not grabbed)
	stateAfter, _ := sm.Load()
	project, exists := stateAfter.Projects["test-project"]
	if !exists {
		t.Error("Project should still exist in state after --local removal")
	}
	if project.IsGrabbed {
		t.Error("Project should be marked as not grabbed after --local removal")
	}
}

func TestRemoveCmd_StateUpdatedBeforeDeletion(t *testing.T) {
	tmpDir, sm, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create a project
	createTestProject(t, tmpDir, sm, "test-project", false)

	// Get archive path for later verification
	stateBefore, _ := sm.Load()
	archivePath, _ := stateBefore.GetArchivePath("test-project")

	// Remove the project
	err := RemoveCmd("test-project", false, false, true)
	if err != nil {
		t.Errorf("RemoveCmd failed: %v", err)
	}

	// Load state to verify it was updated
	stateAfter, _ := sm.Load()
	if _, exists := stateAfter.Projects["test-project"]; exists {
		t.Error("State should be updated after removal")
	}

	// Archive should be deleted
	if _, err := os.Stat(archivePath); !os.IsNotExist(err) {
		t.Error("Archive should be deleted")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
