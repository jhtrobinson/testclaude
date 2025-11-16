package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Project represents a single project's state
type Project struct {
	LocalPath           string     `json:"local_path"`
	Master              string     `json:"master"`
	ArchiveCategory     string     `json:"archive_category"`
	GrabbedAt           *time.Time `json:"grabbed_at"`
	LastParkAt          *time.Time `json:"last_park_at"`
	ArchiveContentHash  *string    `json:"archive_content_hash"`
	LocalContentHash    *string    `json:"local_content_hash"`
	LocalHashComputedAt *time.Time `json:"local_hash_computed_at"`
	LastParkMtime       *time.Time `json:"last_park_mtime"`
	NoHashMode          bool       `json:"no_hash_mode"`
	IsGrabbed           bool       `json:"is_grabbed"`
}

// State represents the entire parkr state file
type State struct {
	Masters          map[string]map[string]string `json:"masters"`
	DefaultMaster    string                       `json:"default_master"`
	Projects         map[string]*Project          `json:"projects"`
	LocalDirectories []string                     `json:"local_directories,omitempty"`
}

// StateManager handles reading and writing state
type StateManager struct {
	statePath string
}

// NewStateManager creates a state manager with default path
func NewStateManager() *StateManager {
	homeDir, _ := os.UserHomeDir()
	return &StateManager{
		statePath: filepath.Join(homeDir, ".parkr", "state.json"),
	}
}

// StatePath returns the path to the state file
func (sm *StateManager) StatePath() string {
	return sm.statePath
}

// Load reads the state file from disk
func (sm *StateManager) Load() (*State, error) {
	data, err := os.ReadFile(sm.statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("state file not found at %s - run 'parkr init' first", sm.statePath)
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	// Initialize maps if nil
	if state.Projects == nil {
		state.Projects = make(map[string]*Project)
	}
	if state.Masters == nil {
		state.Masters = make(map[string]map[string]string)
	}

	return &state, nil
}

// Save writes the state file to disk
func (sm *StateManager) Save(state *State) error {
	// Ensure directory exists
	dir := filepath.Dir(sm.statePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize state: %w", err)
	}

	// Write to temp file first, then rename (atomic)
	tmpPath := sm.statePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	if err := os.Rename(tmpPath, sm.statePath); err != nil {
		os.Remove(tmpPath) // Clean up on failure
		return fmt.Errorf("failed to save state file: %w", err)
	}

	return nil
}

// Exists checks if the state file exists
func (sm *StateManager) Exists() bool {
	_, err := os.Stat(sm.statePath)
	return err == nil
}

// CreateDefault creates a new state file with default configuration
func (sm *StateManager) CreateDefault() error {
	return sm.CreateWithRoot("/tmp/parkr-archive")
}

// CreateWithRoot creates a new state file with the given archive root
func (sm *StateManager) CreateWithRoot(archiveRoot string) error {
	state := &State{
		Masters: map[string]map[string]string{
			"primary": {
				"code":    filepath.Join(archiveRoot, "code"),
				"pycharm": filepath.Join(archiveRoot, "pycharm"),
				"rstudio": filepath.Join(archiveRoot, "rstudio"),
				"misc":    filepath.Join(archiveRoot, "misc"),
			},
		},
		DefaultMaster: "primary",
		Projects:      make(map[string]*Project),
	}

	return sm.Save(state)
}

// GetArchivePath returns the full archive path for a project
func (s *State) GetArchivePath(projectName string) (string, error) {
	project, exists := s.Projects[projectName]
	if !exists {
		return "", fmt.Errorf("project '%s' not found in state", projectName)
	}

	master, exists := s.Masters[project.Master]
	if !exists {
		return "", fmt.Errorf("master '%s' not found", project.Master)
	}

	categoryPath, exists := master[project.ArchiveCategory]
	if !exists {
		return "", fmt.Errorf("category '%s' not found in master '%s'", project.ArchiveCategory, project.Master)
	}

	return filepath.Join(categoryPath, projectName), nil
}

// GetDefaultLocalPath returns the default local path for a category
func GetDefaultLocalPath(category string) string {
	homeDir, _ := os.UserHomeDir()

	switch category {
	case "pycharm":
		return filepath.Join(homeDir, "PycharmProjects")
	case "rstudio":
		return filepath.Join(homeDir, "RStudioProjects")
	default:
		return filepath.Join(homeDir, "code")
	}
}
