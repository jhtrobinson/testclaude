package core

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSelectPruneCandidates_EmptyState(t *testing.T) {
	state := &State{
		Projects: make(map[string]*Project),
	}

	opts := PruneOptions{
		TargetBytes: 100 * Megabyte,
	}

	result, err := SelectPruneCandidates(state, opts.TargetBytes, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.NoCandidates {
		t.Error("expected NoCandidates to be true")
	}
	if len(result.Candidates) != 0 {
		t.Errorf("expected 0 candidates, got %d", len(result.Candidates))
	}
}

func TestSelectPruneCandidates_NoCandidates_AllDirty(t *testing.T) {
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "dirty-project")
	os.MkdirAll(projectPath, 0755)
	os.WriteFile(filepath.Join(projectPath, "test.txt"), []byte("test"), 0644)

	state := &State{
		Projects: map[string]*Project{
			"dirty-project": {
				LocalPath:  projectPath,
				IsGrabbed:  true,
				LastParkAt: nil, // Never parked
			},
		},
	}

	opts := PruneOptions{
		TargetBytes: 100 * Megabyte,
	}

	result, err := SelectPruneCandidates(state, opts.TargetBytes, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.NoCandidates {
		t.Error("expected NoCandidates to be true when all projects are dirty")
	}
}

func TestSelectPruneCandidates_SelectsOldestFirst(t *testing.T) {
	tmpDir := t.TempDir()

	// Create three projects with different ages
	oldProjectPath := filepath.Join(tmpDir, "old-project")
	os.MkdirAll(oldProjectPath, 0755)
	oldFile := filepath.Join(oldProjectPath, "test.txt")
	os.WriteFile(oldFile, []byte("old content - 10MB padding"), 0644)

	mediumProjectPath := filepath.Join(tmpDir, "medium-project")
	os.MkdirAll(mediumProjectPath, 0755)
	mediumFile := filepath.Join(mediumProjectPath, "test.txt")
	os.WriteFile(mediumFile, []byte("medium content"), 0644)

	newProjectPath := filepath.Join(tmpDir, "new-project")
	os.MkdirAll(newProjectPath, 0755)
	newFile := filepath.Join(newProjectPath, "test.txt")
	os.WriteFile(newFile, []byte("new content"), 0644)

	// Set modification times
	oldTime := time.Now().Add(-3 * 24 * time.Hour)
	mediumTime := time.Now().Add(-2 * 24 * time.Hour)
	newTime := time.Now().Add(-1 * 24 * time.Hour)

	os.Chtimes(oldFile, oldTime, oldTime)
	os.Chtimes(mediumFile, mediumTime, mediumTime)
	os.Chtimes(newFile, newTime, newTime)

	// Park times are after modifications
	oldParkTime := oldTime.Add(time.Minute)
	mediumParkTime := mediumTime.Add(time.Minute)
	newParkTime := newTime.Add(time.Minute)

	state := &State{
		Projects: map[string]*Project{
			"old-project": {
				LocalPath:     oldProjectPath,
				IsGrabbed:     true,
				LastParkAt:    &oldParkTime,
				LastParkMtime: &oldTime,
			},
			"medium-project": {
				LocalPath:     mediumProjectPath,
				IsGrabbed:     true,
				LastParkAt:    &mediumParkTime,
				LastParkMtime: &mediumTime,
			},
			"new-project": {
				LocalPath:     newProjectPath,
				IsGrabbed:     true,
				LastParkAt:    &newParkTime,
				LastParkMtime: &newTime,
			},
		},
	}

	opts := PruneOptions{
		TargetBytes: 100, // Small target to select just one
	}

	result, err := SelectPruneCandidates(state, opts.TargetBytes, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.SelectedProjects) == 0 {
		t.Fatal("expected at least one selected project")
	}

	// First selected should be the oldest
	if result.SelectedProjects[0].Name != "old-project" {
		t.Errorf("expected oldest project to be selected first, got %s", result.SelectedProjects[0].Name)
	}
}

func TestSelectPruneCandidates_StopsAtTarget(t *testing.T) {
	tmpDir := t.TempDir()

	// Create projects with known sizes
	project1Path := filepath.Join(tmpDir, "project1")
	os.MkdirAll(project1Path, 0755)
	file1 := filepath.Join(project1Path, "data.bin")
	// Create a 1KB file
	os.WriteFile(file1, make([]byte, 1024), 0644)

	project2Path := filepath.Join(tmpDir, "project2")
	os.MkdirAll(project2Path, 0755)
	file2 := filepath.Join(project2Path, "data.bin")
	os.WriteFile(file2, make([]byte, 1024), 0644)

	project3Path := filepath.Join(tmpDir, "project3")
	os.MkdirAll(project3Path, 0755)
	file3 := filepath.Join(project3Path, "data.bin")
	os.WriteFile(file3, make([]byte, 1024), 0644)

	// Set times so they're in order
	time1 := time.Now().Add(-3 * time.Hour)
	time2 := time.Now().Add(-2 * time.Hour)
	time3 := time.Now().Add(-1 * time.Hour)

	os.Chtimes(file1, time1, time1)
	os.Chtimes(file2, time2, time2)
	os.Chtimes(file3, time3, time3)

	parkTime1 := time1.Add(time.Minute)
	parkTime2 := time2.Add(time.Minute)
	parkTime3 := time3.Add(time.Minute)

	state := &State{
		Projects: map[string]*Project{
			"project1": {
				LocalPath:     project1Path,
				IsGrabbed:     true,
				LastParkAt:    &parkTime1,
				LastParkMtime: &time1,
			},
			"project2": {
				LocalPath:     project2Path,
				IsGrabbed:     true,
				LastParkAt:    &parkTime2,
				LastParkMtime: &time2,
			},
			"project3": {
				LocalPath:     project3Path,
				IsGrabbed:     true,
				LastParkAt:    &parkTime3,
				LastParkMtime: &time3,
			},
		},
	}

	opts := PruneOptions{
		TargetBytes: 2048, // Just enough for 2 projects
	}

	result, err := SelectPruneCandidates(state, opts.TargetBytes, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should select 2 projects to reach target
	if len(result.SelectedProjects) != 2 {
		t.Errorf("expected 2 selected projects, got %d", len(result.SelectedProjects))
	}

	if result.TotalSelected < 2048 {
		t.Errorf("expected total selected >= 2048, got %d", result.TotalSelected)
	}
}

func TestSelectPruneCandidates_InsufficientSpace(t *testing.T) {
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "small-project")
	os.MkdirAll(projectPath, 0755)
	testFile := filepath.Join(projectPath, "test.txt")
	os.WriteFile(testFile, []byte("small"), 0644)

	oldTime := time.Now().Add(-time.Hour)
	os.Chtimes(testFile, oldTime, oldTime)
	parkTime := oldTime.Add(time.Minute)

	state := &State{
		Projects: map[string]*Project{
			"small-project": {
				LocalPath:     projectPath,
				IsGrabbed:     true,
				LastParkAt:    &parkTime,
				LastParkMtime: &oldTime,
			},
		},
	}

	opts := PruneOptions{
		TargetBytes: 1 * Terabyte, // Huge target
	}

	result, err := SelectPruneCandidates(state, opts.TargetBytes, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.InsufficientSpace {
		t.Error("expected InsufficientSpace to be true")
	}
}

func TestSelectPruneCandidates_ForceIncludesDirty(t *testing.T) {
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "dirty-project")
	os.MkdirAll(projectPath, 0755)
	testFile := filepath.Join(projectPath, "test.txt")
	os.WriteFile(testFile, []byte("content"), 0644)

	state := &State{
		Projects: map[string]*Project{
			"dirty-project": {
				LocalPath:  projectPath,
				IsGrabbed:  true,
				LastParkAt: nil, // Never parked
			},
		},
	}

	opts := PruneOptions{
		TargetBytes: 100,
		Force:       true,
	}

	result, err := SelectPruneCandidates(state, opts.TargetBytes, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.NoCandidates {
		t.Error("expected force mode to include dirty projects")
	}

	if len(result.SelectedProjects) != 1 {
		t.Errorf("expected 1 selected project, got %d", len(result.SelectedProjects))
	}

	// Check for warning
	hasWarning := false
	for _, w := range result.Warnings {
		if w != "" {
			hasWarning = true
			break
		}
	}
	if !hasWarning {
		t.Error("expected warning about force mode")
	}
}

func TestExecutePrune_DeletesProjects(t *testing.T) {
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "test-project")
	os.MkdirAll(projectPath, 0755)
	testFile := filepath.Join(projectPath, "test.txt")
	os.WriteFile(testFile, []byte("test content"), 0644)

	oldTime := time.Now().Add(-time.Hour)
	os.Chtimes(testFile, oldTime, oldTime)
	parkTime := oldTime.Add(time.Minute)

	// Create a temporary state file
	stateDir := filepath.Join(tmpDir, ".parkr")
	os.MkdirAll(stateDir, 0755)
	sm := &StateManager{statePath: filepath.Join(stateDir, "state.json")}

	state := &State{
		Masters: map[string]map[string]string{
			"primary": {
				"code": filepath.Join(tmpDir, "archive"),
			},
		},
		DefaultMaster: "primary",
		Projects: map[string]*Project{
			"test-project": {
				LocalPath:       projectPath,
				Master:          "primary",
				ArchiveCategory: "code",
				IsGrabbed:       true,
				LastParkAt:      &parkTime,
				LastParkMtime:   &oldTime,
			},
		},
	}

	// Save initial state
	sm.Save(state)

	// Create result with selected projects
	result := &PruneResult{
		SelectedProjects: []ProjectReport{
			{
				Name:      "test-project",
				LocalPath: projectPath,
				LocalSize: 12, // "test content" size
			},
		},
		TargetBytes: 100,
	}

	opts := PruneOptions{
		Execute: true,
		NoHash:  true, // Use mtime verification for this test
	}

	// Override state manager for this test
	origNewStateManagerFn := newStateManagerFn
	defer func() { newStateManagerFn = origNewStateManagerFn }()
	newStateManagerFn = func() *StateManager { return sm }

	err := ExecutePrune(state, result, opts, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify project was deleted
	if _, err := os.Stat(projectPath); !os.IsNotExist(err) {
		t.Error("expected project directory to be deleted")
	}

	if len(result.Deleted) != 1 {
		t.Errorf("expected 1 deleted project, got %d", len(result.Deleted))
	}

	// Verify state was updated
	if state.Projects["test-project"].IsGrabbed {
		t.Error("expected IsGrabbed to be false after deletion")
	}
}

func TestExecutePrune_SkipsModifiedProjects(t *testing.T) {
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "modified-project")
	os.MkdirAll(projectPath, 0755)
	testFile := filepath.Join(projectPath, "test.txt")
	os.WriteFile(testFile, []byte("test content"), 0644)

	// Park time is in the past, but file is modified after
	parkTime := time.Now().Add(-2 * time.Hour)
	parkMtime := parkTime.Add(-time.Minute)

	// File is modified after park (simulating uncommitted changes)
	newTime := time.Now().Add(-30 * time.Minute)
	os.Chtimes(testFile, newTime, newTime)

	stateDir := filepath.Join(tmpDir, ".parkr")
	os.MkdirAll(stateDir, 0755)
	sm := &StateManager{statePath: filepath.Join(stateDir, "state.json")}

	state := &State{
		Projects: map[string]*Project{
			"modified-project": {
				LocalPath:     projectPath,
				IsGrabbed:     true,
				LastParkAt:    &parkTime,
				LastParkMtime: &parkMtime,
			},
		},
	}
	sm.Save(state)

	result := &PruneResult{
		SelectedProjects: []ProjectReport{
			{
				Name:      "modified-project",
				LocalPath: projectPath,
				LocalSize: 12,
			},
		},
		TargetBytes: 100,
	}

	opts := PruneOptions{
		Execute: true,
		NoHash:  true, // Use mtime check
	}

	origNewStateManagerFn := newStateManagerFn
	defer func() { newStateManagerFn = origNewStateManagerFn }()
	newStateManagerFn = func() *StateManager { return sm }

	err := ExecutePrune(state, result, opts, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify project was NOT deleted (it was modified after park)
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		t.Error("expected project directory to still exist")
	}

	if len(result.FailedDeletions) != 1 {
		t.Errorf("expected 1 failed deletion, got %d", len(result.FailedDeletions))
	}
}

func TestVerifyBeforeDeletion_NeverParked(t *testing.T) {
	project := &Project{
		LastParkAt: nil,
	}

	safe, _ := verifyBeforeDeletion(project, false)
	if safe {
		t.Error("never-parked project should not be safe")
	}
}

func TestVerifyBeforeDeletion_SafeWithMtime(t *testing.T) {
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "test")
	os.MkdirAll(projectPath, 0755)
	testFile := filepath.Join(projectPath, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	oldTime := time.Now().Add(-time.Hour)
	os.Chtimes(testFile, oldTime, oldTime)
	parkTime := oldTime.Add(time.Minute)

	project := &Project{
		LocalPath:     projectPath,
		LastParkAt:    &parkTime,
		LastParkMtime: &oldTime,
	}

	safe, status := verifyBeforeDeletion(project, true)
	if !safe {
		t.Errorf("expected safe, got status: %s", status)
	}
}

func TestVerifyBeforeDeletion_UnsafeWithModifiedFile(t *testing.T) {
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "test")
	os.MkdirAll(projectPath, 0755)
	testFile := filepath.Join(projectPath, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	// Park time is in past
	parkTime := time.Now().Add(-time.Hour)
	parkMtime := parkTime.Add(-time.Minute)

	// File is newer than park time
	newTime := time.Now().Add(-30 * time.Minute)
	os.Chtimes(testFile, newTime, newTime)

	project := &Project{
		LocalPath:     projectPath,
		LastParkAt:    &parkTime,
		LastParkMtime: &parkMtime,
	}

	safe, status := verifyBeforeDeletion(project, true)
	if safe {
		t.Error("expected unsafe for modified file")
	}
	if status != "Has uncommitted work" {
		t.Errorf("expected 'Has uncommitted work', got '%s'", status)
	}
}
