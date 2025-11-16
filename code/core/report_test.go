package core

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGenerateReport_EmptyState(t *testing.T) {
	state := &State{
		Projects: make(map[string]*Project),
	}

	summary, err := GenerateReport(state, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if summary.TotalProjects != 0 {
		t.Errorf("expected 0 projects, got %d", summary.TotalProjects)
	}
	if summary.TotalSize != 0 {
		t.Errorf("expected 0 total size, got %d", summary.TotalSize)
	}
	if len(summary.Projects) != 0 {
		t.Errorf("expected empty projects list, got %d", len(summary.Projects))
	}
	if len(summary.Candidates) != 0 {
		t.Errorf("expected empty candidates list, got %d", len(summary.Candidates))
	}
}

func TestGenerateReport_NeverParked(t *testing.T) {
	// Create a temporary directory for the project
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "test-project")
	os.MkdirAll(projectPath, 0755)
	os.WriteFile(filepath.Join(projectPath, "test.txt"), []byte("test content"), 0644)

	state := &State{
		Projects: map[string]*Project{
			"test-project": {
				LocalPath:  projectPath,
				IsGrabbed:  true,
				LastParkAt: nil, // Never parked
			},
		},
	}

	summary, err := GenerateReport(state, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if summary.TotalProjects != 1 {
		t.Errorf("expected 1 project, got %d", summary.TotalProjects)
	}

	if len(summary.Projects) != 1 {
		t.Fatalf("expected 1 project in list, got %d", len(summary.Projects))
	}

	project := summary.Projects[0]
	if project.IsSafeDelete {
		t.Error("never-parked project should not be safe to delete")
	}
	if project.Status != "Never checked in" {
		t.Errorf("expected status 'Never checked in', got '%s'", project.Status)
	}
	if !project.NeverParked {
		t.Error("NeverParked should be true")
	}

	if len(summary.Candidates) != 0 {
		t.Errorf("expected 0 candidates, got %d", len(summary.Candidates))
	}
}

func TestGenerateReport_SafeToDelete(t *testing.T) {
	// Create a temporary directory for the project
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "safe-project")
	os.MkdirAll(projectPath, 0755)
	testFile := filepath.Join(projectPath, "test.txt")
	os.WriteFile(testFile, []byte("test content"), 0644)

	// Set file mtime to 1 hour ago
	oldTime := time.Now().Add(-time.Hour)
	os.Chtimes(testFile, oldTime, oldTime)

	// Park time is after the file modification
	parkTime := time.Now().Add(-30 * time.Minute)
	parkMtime := oldTime

	state := &State{
		Projects: map[string]*Project{
			"safe-project": {
				LocalPath:     projectPath,
				IsGrabbed:     true,
				LastParkAt:    &parkTime,
				LastParkMtime: &parkMtime,
			},
		},
	}

	summary, err := GenerateReport(state, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(summary.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(summary.Projects))
	}

	project := summary.Projects[0]
	if !project.IsSafeDelete {
		t.Error("project should be safe to delete")
	}
	if project.Status != "Safe to delete" {
		t.Errorf("expected status 'Safe to delete', got '%s'", project.Status)
	}

	if len(summary.Candidates) != 1 {
		t.Errorf("expected 1 candidate, got %d", len(summary.Candidates))
	}
}

func TestGenerateReport_HasUncommittedWork(t *testing.T) {
	// Create a temporary directory for the project
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "dirty-project")
	os.MkdirAll(projectPath, 0755)
	testFile := filepath.Join(projectPath, "test.txt")
	os.WriteFile(testFile, []byte("test content"), 0644)

	// Park time is before the file modification
	parkTime := time.Now().Add(-time.Hour)
	parkMtime := time.Now().Add(-time.Hour)

	// File is modified after park (current time)
	// Default file time is "now", so it's after parkTime

	state := &State{
		Projects: map[string]*Project{
			"dirty-project": {
				LocalPath:     projectPath,
				IsGrabbed:     true,
				LastParkAt:    &parkTime,
				LastParkMtime: &parkMtime,
			},
		},
	}

	summary, err := GenerateReport(state, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(summary.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(summary.Projects))
	}

	project := summary.Projects[0]
	if project.IsSafeDelete {
		t.Error("dirty project should not be safe to delete")
	}
	if project.Status != "Has uncommitted work" {
		t.Errorf("expected status 'Has uncommitted work', got '%s'", project.Status)
	}

	if len(summary.Candidates) != 0 {
		t.Errorf("expected 0 candidates, got %d", len(summary.Candidates))
	}
}

func TestSortProjects_BySize(t *testing.T) {
	projects := []ProjectReport{
		{Name: "small", LocalSize: 100},
		{Name: "large", LocalSize: 1000},
		{Name: "medium", LocalSize: 500},
	}

	SortProjects(projects, SortBySize)

	if projects[0].Name != "large" {
		t.Errorf("expected 'large' first, got '%s'", projects[0].Name)
	}
	if projects[1].Name != "medium" {
		t.Errorf("expected 'medium' second, got '%s'", projects[1].Name)
	}
	if projects[2].Name != "small" {
		t.Errorf("expected 'small' third, got '%s'", projects[2].Name)
	}
}

func TestSortProjects_ByName(t *testing.T) {
	projects := []ProjectReport{
		{Name: "charlie"},
		{Name: "alpha"},
		{Name: "bravo"},
	}

	SortProjects(projects, SortByName)

	if projects[0].Name != "alpha" {
		t.Errorf("expected 'alpha' first, got '%s'", projects[0].Name)
	}
	if projects[1].Name != "bravo" {
		t.Errorf("expected 'bravo' second, got '%s'", projects[1].Name)
	}
	if projects[2].Name != "charlie" {
		t.Errorf("expected 'charlie' third, got '%s'", projects[2].Name)
	}
}

func TestSortProjects_ByModified(t *testing.T) {
	now := time.Now()
	projects := []ProjectReport{
		{Name: "recent", LastModified: now},
		{Name: "oldest", LastModified: now.Add(-2 * time.Hour)},
		{Name: "middle", LastModified: now.Add(-time.Hour)},
	}

	SortProjects(projects, SortByModified)

	if projects[0].Name != "oldest" {
		t.Errorf("expected 'oldest' first, got '%s'", projects[0].Name)
	}
	if projects[1].Name != "middle" {
		t.Errorf("expected 'middle' second, got '%s'", projects[1].Name)
	}
	if projects[2].Name != "recent" {
		t.Errorf("expected 'recent' third, got '%s'", projects[2].Name)
	}
}

func TestFilterCandidates(t *testing.T) {
	projects := []ProjectReport{
		{Name: "safe1", IsSafeDelete: true},
		{Name: "unsafe", IsSafeDelete: false},
		{Name: "safe2", IsSafeDelete: true},
	}

	candidates := FilterCandidates(projects)

	if len(candidates) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(candidates))
	}

	for _, c := range candidates {
		if !c.IsSafeDelete {
			t.Errorf("candidate '%s' should be safe to delete", c.Name)
		}
	}
}

func TestGenerateReport_RecoverableSpace(t *testing.T) {
	// Create temporary directories for projects
	tmpDir := t.TempDir()

	// Safe project 1
	safeProject1 := filepath.Join(tmpDir, "safe1")
	os.MkdirAll(safeProject1, 0755)
	os.WriteFile(filepath.Join(safeProject1, "data.txt"), make([]byte, 1000), 0644)

	// Safe project 2
	safeProject2 := filepath.Join(tmpDir, "safe2")
	os.MkdirAll(safeProject2, 0755)
	os.WriteFile(filepath.Join(safeProject2, "data.txt"), make([]byte, 2000), 0644)

	// Unsafe project
	unsafeProject := filepath.Join(tmpDir, "unsafe")
	os.MkdirAll(unsafeProject, 0755)
	os.WriteFile(filepath.Join(unsafeProject, "data.txt"), make([]byte, 500), 0644)

	// Set old mtimes for safe projects
	oldTime := time.Now().Add(-time.Hour)
	os.Chtimes(filepath.Join(safeProject1, "data.txt"), oldTime, oldTime)
	os.Chtimes(filepath.Join(safeProject2, "data.txt"), oldTime, oldTime)

	parkTime := time.Now().Add(-30 * time.Minute)

	state := &State{
		Projects: map[string]*Project{
			"safe1": {
				LocalPath:     safeProject1,
				IsGrabbed:     true,
				LastParkAt:    &parkTime,
				LastParkMtime: &oldTime,
			},
			"safe2": {
				LocalPath:     safeProject2,
				IsGrabbed:     true,
				LastParkAt:    &parkTime,
				LastParkMtime: &oldTime,
			},
			"unsafe": {
				LocalPath:  unsafeProject,
				IsGrabbed:  true,
				LastParkAt: nil, // Never parked
			},
		},
	}

	summary, err := GenerateReport(state, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if summary.SafeToDelete != 2 {
		t.Errorf("expected 2 safe to delete, got %d", summary.SafeToDelete)
	}

	// Total size should be all 3 projects
	if summary.TotalSize != 3500 {
		t.Errorf("expected total size 3500, got %d", summary.TotalSize)
	}

	// Recoverable space should only include safe projects
	if summary.RecoverableSpace != 3000 {
		t.Errorf("expected recoverable space 3000, got %d", summary.RecoverableSpace)
	}
}

func TestGenerateReport_SkipsNonGrabbedProjects(t *testing.T) {
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "archived-project")
	os.MkdirAll(projectPath, 0755)
	os.WriteFile(filepath.Join(projectPath, "test.txt"), []byte("test"), 0644)

	state := &State{
		Projects: map[string]*Project{
			"archived-project": {
				LocalPath: projectPath,
				IsGrabbed: false, // Not grabbed
			},
		},
	}

	summary, err := GenerateReport(state, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if summary.TotalProjects != 0 {
		t.Errorf("expected 0 projects (non-grabbed should be skipped), got %d", summary.TotalProjects)
	}
}
