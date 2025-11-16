package core

import (
	"testing"
	"time"
)

func TestNewInteractiveSelector(t *testing.T) {
	candidates := []PruneCandidate{
		{
			ProjectReport: ProjectReport{
				Name:      "project1",
				LocalSize: 1000000000, // 1GB
			},
			Selected: true,
		},
		{
			ProjectReport: ProjectReport{
				Name:      "project2",
				LocalSize: 2000000000, // 2GB
			},
			Selected: false,
		},
		{
			ProjectReport: ProjectReport{
				Name:      "project3",
				LocalSize: 500000000, // 500MB
			},
			Selected: true,
		},
	}

	targetBytes := int64(3000000000) // 3GB

	selector := NewInteractiveSelector(candidates, targetBytes)

	// Check that pre-selected candidates are selected
	if !selector.selected[0] {
		t.Error("Expected project1 to be pre-selected")
	}
	if selector.selected[1] {
		t.Error("Expected project2 to not be pre-selected")
	}
	if !selector.selected[2] {
		t.Error("Expected project3 to be pre-selected")
	}

	// Check total selected
	expectedTotal := int64(1500000000) // 1.5GB
	if selector.totalSelected != expectedTotal {
		t.Errorf("Expected total selected %d, got %d", expectedTotal, selector.totalSelected)
	}

	// Check initial state
	if selector.cursor != 0 {
		t.Errorf("Expected cursor at 0, got %d", selector.cursor)
	}
	if selector.quitting {
		t.Error("Expected quitting to be false")
	}
	if selector.confirmed {
		t.Error("Expected confirmed to be false")
	}
}

func TestInteractiveSelector_HandleInput_Navigation(t *testing.T) {
	candidates := makeCandidates(3)
	selector := NewInteractiveSelector(candidates, 1000)

	// Test moving down with 'j'
	selector.handleInput('j')
	if selector.cursor != 1 {
		t.Errorf("Expected cursor at 1 after 'j', got %d", selector.cursor)
	}

	// Test moving down with arrow key
	selector.handleInput('B')
	if selector.cursor != 2 {
		t.Errorf("Expected cursor at 2 after down arrow, got %d", selector.cursor)
	}

	// Test moving down at bottom (should stay)
	selector.handleInput('j')
	if selector.cursor != 2 {
		t.Errorf("Expected cursor to stay at 2, got %d", selector.cursor)
	}

	// Test moving up with 'k'
	selector.handleInput('k')
	if selector.cursor != 1 {
		t.Errorf("Expected cursor at 1 after 'k', got %d", selector.cursor)
	}

	// Test moving up with arrow key
	selector.handleInput('A')
	if selector.cursor != 0 {
		t.Errorf("Expected cursor at 0 after up arrow, got %d", selector.cursor)
	}

	// Test moving up at top (should stay)
	selector.handleInput('k')
	if selector.cursor != 0 {
		t.Errorf("Expected cursor to stay at 0, got %d", selector.cursor)
	}
}

func TestInteractiveSelector_HandleInput_Toggle(t *testing.T) {
	candidates := makeCandidates(3)
	selector := NewInteractiveSelector(candidates, 1000)

	// Initially nothing is selected
	if selector.totalSelected != 0 {
		t.Errorf("Expected total selected 0, got %d", selector.totalSelected)
	}

	// Toggle selection with space
	selector.handleInput(' ')
	if !selector.selected[0] {
		t.Error("Expected project 0 to be selected after toggle")
	}
	if selector.totalSelected != 100 {
		t.Errorf("Expected total selected 100, got %d", selector.totalSelected)
	}

	// Toggle again to deselect
	selector.handleInput(' ')
	if selector.selected[0] {
		t.Error("Expected project 0 to be deselected after second toggle")
	}
	if selector.totalSelected != 0 {
		t.Errorf("Expected total selected 0, got %d", selector.totalSelected)
	}
}

func TestInteractiveSelector_HandleInput_SelectAll(t *testing.T) {
	candidates := makeCandidates(3)
	selector := NewInteractiveSelector(candidates, 1000)

	// Select all with 'a'
	selector.handleInput('a')
	if len(selector.selected) != 3 {
		t.Errorf("Expected all 3 projects selected, got %d", len(selector.selected))
	}
	expectedTotal := int64(300) // 100 * 3
	if selector.totalSelected != expectedTotal {
		t.Errorf("Expected total selected %d, got %d", expectedTotal, selector.totalSelected)
	}

	// Toggle all (deselect) with 'a'
	selector.handleInput('a')
	if len(selector.selected) != 0 {
		t.Errorf("Expected no projects selected, got %d", len(selector.selected))
	}
	if selector.totalSelected != 0 {
		t.Errorf("Expected total selected 0, got %d", selector.totalSelected)
	}
}

func TestInteractiveSelector_HandleInput_Quit(t *testing.T) {
	candidates := makeCandidates(3)
	selector := NewInteractiveSelector(candidates, 1000)

	// Test quit with 'q'
	result := selector.handleInput('q')
	if result {
		t.Error("Expected handleInput to return false for quit")
	}
	if !selector.quitting {
		t.Error("Expected quitting to be true")
	}

	// Reset and test ESC
	selector = NewInteractiveSelector(candidates, 1000)
	result = selector.handleInput(27)
	if result {
		t.Error("Expected handleInput to return false for ESC")
	}
	if !selector.quitting {
		t.Error("Expected quitting to be true after ESC")
	}
}

func TestInteractiveSelector_HandleInput_Confirm(t *testing.T) {
	candidates := makeCandidates(3)
	selector := NewInteractiveSelector(candidates, 1000)

	// Test confirm with Enter
	result := selector.handleInput('\n')
	if result {
		t.Error("Expected handleInput to return false for Enter")
	}
	if !selector.confirmed {
		t.Error("Expected confirmed to be true")
	}

	// Reset and test carriage return
	selector = NewInteractiveSelector(candidates, 1000)
	result = selector.handleInput('\r')
	if result {
		t.Error("Expected handleInput to return false for carriage return")
	}
	if !selector.confirmed {
		t.Error("Expected confirmed to be true after carriage return")
	}
}

func TestInteractiveSelector_GetSelected(t *testing.T) {
	candidates := makeCandidates(3)
	selector := NewInteractiveSelector(candidates, 1000)

	// Select first and third
	selector.selected[0] = true
	selector.selected[2] = true

	selected := selector.GetSelected()
	if len(selected) != 2 {
		t.Errorf("Expected 2 selected, got %d", len(selected))
	}

	// Check order preservation
	if selected[0].Name != "project-0" {
		t.Errorf("Expected first selected to be project-0, got %s", selected[0].Name)
	}
	if selected[1].Name != "project-2" {
		t.Errorf("Expected second selected to be project-2, got %s", selected[1].Name)
	}
}

func TestInteractiveSelector_StatusMethods(t *testing.T) {
	candidates := makeCandidates(2)
	selector := NewInteractiveSelector(candidates, 1000)

	// Initial state
	if selector.WasConfirmed() {
		t.Error("Expected WasConfirmed to be false initially")
	}
	if selector.WasQuit() {
		t.Error("Expected WasQuit to be false initially")
	}
	if selector.TotalSelected() != 0 {
		t.Errorf("Expected TotalSelected to be 0 initially, got %d", selector.TotalSelected())
	}

	// After selecting
	selector.selected[0] = true
	selector.totalSelected = 100
	if selector.TotalSelected() != 100 {
		t.Errorf("Expected TotalSelected to be 100, got %d", selector.TotalSelected())
	}

	// After confirming
	selector.confirmed = true
	if !selector.WasConfirmed() {
		t.Error("Expected WasConfirmed to be true after confirmation")
	}

	// After quitting
	selector.quitting = true
	if !selector.WasQuit() {
		t.Error("Expected WasQuit to be true after quit")
	}
}

func TestFormatAge(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{
			name:     "zero time",
			time:     time.Time{},
			expected: "never",
		},
		{
			name:     "just now",
			time:     time.Now().Add(-30 * time.Second),
			expected: "just now",
		},
		{
			name:     "1 min ago",
			time:     time.Now().Add(-1 * time.Minute),
			expected: "1 min ago",
		},
		{
			name:     "multiple mins ago",
			time:     time.Now().Add(-5 * time.Minute),
			expected: "5 mins ago",
		},
		{
			name:     "1 hour ago",
			time:     time.Now().Add(-1 * time.Hour),
			expected: "1 hour ago",
		},
		{
			name:     "multiple hours ago",
			time:     time.Now().Add(-3 * time.Hour),
			expected: "3 hours ago",
		},
		{
			name:     "1 day ago",
			time:     time.Now().Add(-24 * time.Hour),
			expected: "1 day ago",
		},
		{
			name:     "multiple days ago",
			time:     time.Now().Add(-3 * 24 * time.Hour),
			expected: "3 days ago",
		},
		{
			name:     "1 week ago",
			time:     time.Now().Add(-7 * 24 * time.Hour),
			expected: "1 week ago",
		},
		{
			name:     "multiple weeks ago",
			time:     time.Now().Add(-21 * 24 * time.Hour),
			expected: "3 weeks ago",
		},
		{
			name:     "1 month ago",
			time:     time.Now().Add(-30 * 24 * time.Hour),
			expected: "1 month ago",
		},
		{
			name:     "multiple months ago",
			time:     time.Now().Add(-90 * 24 * time.Hour),
			expected: "3 months ago",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatAge(tc.time)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

// Helper function to create test candidates
func makeCandidates(n int) []PruneCandidate {
	candidates := make([]PruneCandidate, n)
	for i := 0; i < n; i++ {
		candidates[i] = PruneCandidate{
			ProjectReport: ProjectReport{
				Name:         "project-" + string(rune('0'+i)),
				LocalSize:    100,
				LastModified: time.Now().Add(-time.Duration(i) * 24 * time.Hour),
			},
			Selected: false,
		}
	}
	return candidates
}
