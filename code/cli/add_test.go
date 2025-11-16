package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jamespark/parkr/core"
)

func setupAddTestEnv(t *testing.T) (string, string, func()) {
	// Create temp directories for testing
	tmpDir, err := os.MkdirTemp("", "parkr-add-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Create archive structure
	archiveDir := filepath.Join(tmpDir, "archive")
	codeArchive := filepath.Join(archiveDir, "code")
	pycharmArchive := filepath.Join(archiveDir, "pycharm")
	rstudioArchive := filepath.Join(archiveDir, "rstudio")

	for _, dir := range []string{codeArchive, pycharmArchive, rstudioArchive} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create archive dir: %v", err)
		}
	}

	// Create parkr state directory
	parkrDir := filepath.Join(tmpDir, ".parkr")
	if err := os.MkdirAll(parkrDir, 0755); err != nil {
		t.Fatalf("failed to create parkr dir: %v", err)
	}

	// Create state file
	statePath := filepath.Join(parkrDir, "state.json")
	state := &core.State{
		Masters: map[string]map[string]string{
			"primary": {
				"code":    codeArchive,
				"pycharm": pycharmArchive,
				"rstudio": rstudioArchive,
			},
		},
		DefaultMaster: "primary",
		Projects:      make(map[string]*core.Project),
	}

	stateData, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal state: %v", err)
	}
	if err := os.WriteFile(statePath, stateData, 0644); err != nil {
		t.Fatalf("failed to write state: %v", err)
	}

	// Override HOME for state manager
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)

	cleanup := func() {
		os.Setenv("HOME", origHome)
		os.RemoveAll(tmpDir)
	}

	return tmpDir, archiveDir, cleanup
}

func TestDetectProjectCategory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "parkr-detect-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name     string
		files    []string
		expected string
	}{
		{
			name:     "Python with requirements.txt",
			files:    []string{"requirements.txt", "main.py"},
			expected: "pycharm",
		},
		{
			name:     "Python with pyproject.toml",
			files:    []string{"pyproject.toml", "src/app.py"},
			expected: "pycharm",
		},
		{
			name:     "Python with setup.py",
			files:    []string{"setup.py"},
			expected: "pycharm",
		},
		{
			name:     "Python with Pipfile",
			files:    []string{"Pipfile"},
			expected: "pycharm",
		},
		{
			name:     "R with DESCRIPTION",
			files:    []string{"DESCRIPTION", "R/functions.R"},
			expected: "rstudio",
		},
		{
			name:     "Default to code",
			files:    []string{"main.go", "README.md"},
			expected: "code",
		},
		{
			name:     "Node project defaults to code",
			files:    []string{"package.json", "index.js"},
			expected: "code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectDir := filepath.Join(tmpDir, tt.name)
			if err := os.MkdirAll(projectDir, 0755); err != nil {
				t.Fatalf("failed to create project dir: %v", err)
			}

			for _, file := range tt.files {
				filePath := filepath.Join(projectDir, file)
				dir := filepath.Dir(filePath)
				if dir != projectDir {
					os.MkdirAll(dir, 0755)
				}
				if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
					t.Fatalf("failed to create file %s: %v", file, err)
				}
			}

			result := DetectProjectCategory(projectDir)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestAddCmd_BasicAdd(t *testing.T) {
	tmpDir, archiveDir, cleanup := setupAddTestEnv(t)
	defer cleanup()

	// Create a local project
	projectPath := filepath.Join(tmpDir, "test-project")
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectPath, "README.md"), []byte("# Test"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Add project without move
	err := AddCmd(projectPath, "code", false)
	if err != nil {
		t.Fatalf("AddCmd failed: %v", err)
	}

	// Verify archive was created
	archivePath := filepath.Join(archiveDir, "code", "test-project")
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		t.Error("archive directory was not created")
	}

	// Verify README was copied
	if _, err := os.Stat(filepath.Join(archivePath, "README.md")); os.IsNotExist(err) {
		t.Error("README.md was not copied to archive")
	}

	// Verify local copy still exists (no move)
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		t.Error("local copy should still exist when not using --move")
	}

	// Verify state was updated
	sm := core.NewStateManager()
	state, err := sm.Load()
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}

	project, exists := state.Projects["test-project"]
	if !exists {
		t.Error("project was not added to state")
	}

	if !project.IsGrabbed {
		t.Error("project should be marked as grabbed (local copy exists)")
	}

	if project.LocalPath != projectPath {
		t.Errorf("expected LocalPath %s, got %s", projectPath, project.LocalPath)
	}

	if project.LastParkMtime == nil {
		t.Error("LastParkMtime should be set")
	}

	if project.LastParkAt == nil {
		t.Error("LastParkAt should be set")
	}

	if !project.NoHashMode {
		t.Error("NoHashMode should be true for Phase 1")
	}
}

func TestAddCmd_WithMove(t *testing.T) {
	tmpDir, archiveDir, cleanup := setupAddTestEnv(t)
	defer cleanup()

	// Create a local project
	projectPath := filepath.Join(tmpDir, "move-test-project")
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectPath, "main.py"), []byte("print('hello')"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectPath, "requirements.txt"), []byte("flask"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Add project with move and auto-detect category
	err := AddCmd(projectPath, "", true)
	if err != nil {
		t.Fatalf("AddCmd with move failed: %v", err)
	}

	// Verify archive was created (should be in pycharm due to requirements.txt)
	archivePath := filepath.Join(archiveDir, "pycharm", "move-test-project")
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		t.Error("archive directory was not created")
	}

	// Verify files were copied
	if _, err := os.Stat(filepath.Join(archivePath, "main.py")); os.IsNotExist(err) {
		t.Error("main.py was not copied to archive")
	}
	if _, err := os.Stat(filepath.Join(archivePath, "requirements.txt")); os.IsNotExist(err) {
		t.Error("requirements.txt was not copied to archive")
	}

	// Verify local copy was deleted
	if _, err := os.Stat(projectPath); !os.IsNotExist(err) {
		t.Error("local copy should be deleted when using --move")
	}

	// Verify state was updated
	sm := core.NewStateManager()
	state, err := sm.Load()
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}

	project, exists := state.Projects["move-test-project"]
	if !exists {
		t.Error("project was not added to state")
	}

	if project.IsGrabbed {
		t.Error("project should NOT be marked as grabbed (local copy deleted)")
	}

	if project.LocalPath != "" {
		t.Errorf("LocalPath should be empty after move, got %s", project.LocalPath)
	}

	if project.ArchiveCategory != "pycharm" {
		t.Errorf("expected category pycharm (auto-detected), got %s", project.ArchiveCategory)
	}

	if project.LastParkMtime == nil {
		t.Error("LastParkMtime should be set")
	}

	if project.GrabbedAt != nil {
		t.Error("GrabbedAt should be nil after move")
	}
}

func TestAddCmd_AlreadyTracked(t *testing.T) {
	tmpDir, _, cleanup := setupAddTestEnv(t)
	defer cleanup()

	// Create and add a project first
	projectPath := filepath.Join(tmpDir, "duplicate-project")
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectPath, "README.md"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	err := AddCmd(projectPath, "code", false)
	if err != nil {
		t.Fatalf("first AddCmd failed: %v", err)
	}

	// Try to add it again
	err = AddCmd(projectPath, "code", false)
	if err == nil {
		t.Error("expected error when adding already tracked project")
	}
}

func TestAddCmd_NonExistentPath(t *testing.T) {
	_, _, cleanup := setupAddTestEnv(t)
	defer cleanup()

	err := AddCmd("/nonexistent/path/to/project", "code", false)
	if err == nil {
		t.Error("expected error for non-existent path")
	}
}

func TestAddCmd_AutoCreateCategory(t *testing.T) {
	tmpDir, archiveDir, cleanup := setupAddTestEnv(t)
	defer cleanup()

	// Remove the pycharm category directory
	os.RemoveAll(filepath.Join(archiveDir, "pycharm"))

	// Create a Python project
	projectPath := filepath.Join(tmpDir, "auto-create-test")
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectPath, "requirements.txt"), []byte("flask"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Add should auto-create the category directory
	err := AddCmd(projectPath, "", false)
	if err != nil {
		t.Fatalf("AddCmd failed (should auto-create category dir): %v", err)
	}

	// Verify category directory was created
	categoryPath := filepath.Join(archiveDir, "pycharm")
	if _, err := os.Stat(categoryPath); os.IsNotExist(err) {
		t.Error("category directory should have been auto-created")
	}

	// Verify project was archived
	archivePath := filepath.Join(categoryPath, "auto-create-test")
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		t.Error("project was not archived")
	}
}

func TestAddCmd_ArchiveAlreadyExists(t *testing.T) {
	tmpDir, archiveDir, cleanup := setupAddTestEnv(t)
	defer cleanup()

	// Create archive path first
	existingArchive := filepath.Join(archiveDir, "code", "existing-project")
	if err := os.MkdirAll(existingArchive, 0755); err != nil {
		t.Fatalf("failed to create existing archive: %v", err)
	}

	// Create local project with same name
	projectPath := filepath.Join(tmpDir, "existing-project")
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectPath, "README.md"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	err := AddCmd(projectPath, "code", false)
	if err == nil {
		t.Error("expected error when archive path already exists")
	}
}
