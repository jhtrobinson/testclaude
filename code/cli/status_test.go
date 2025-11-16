package cli

import (
	"testing"
	"time"

	"github.com/jamespark/parkr/core"
)

func TestFormatTimeAgo(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"zero time", 0, "never"},
		{"just now", 30 * time.Second, "just now"},
		{"1 minute", 1 * time.Minute, "1 min ago"},
		{"5 minutes", 5 * time.Minute, "5 mins ago"},
		{"59 minutes", 59 * time.Minute, "59 mins ago"},
		{"1 hour", 1 * time.Hour, "1 hour ago"},
		{"2 hours", 2 * time.Hour, "2 hours ago"},
		{"23 hours", 23 * time.Hour, "23 hours ago"},
		{"1 day", 24 * time.Hour, "1 day ago"},
		{"2 days", 48 * time.Hour, "2 days ago"},
		{"6 days", 6 * 24 * time.Hour, "6 days ago"},
		{"1 week", 7 * 24 * time.Hour, "1 week ago"},
		{"2 weeks", 14 * 24 * time.Hour, "2 weeks ago"},
		{"4 weeks", 28 * 24 * time.Hour, "4 weeks ago"},
		{"1 month", 30 * 24 * time.Hour, "1 month ago"},
		{"2 months", 60 * 24 * time.Hour, "2 months ago"},
		{"6 months", 180 * 24 * time.Hour, "6 months ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var testTime time.Time
			if tt.expected == "never" {
				testTime = time.Time{} // zero time
			} else {
				testTime = time.Now().Add(-tt.duration)
			}

			result := formatTimeAgo(testTime)
			if result != tt.expected {
				t.Errorf("formatTimeAgo() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestDetermineStatusInfo(t *testing.T) {
	now := time.Now()
	pastHour := now.Add(-1 * time.Hour)
	pastDay := now.Add(-24 * time.Hour)

	tests := []struct {
		name         string
		project      *core.Project
		lastModified time.Time
		expectedText string
	}{
		{
			name: "never checked in",
			project: &core.Project{
				IsGrabbed:  true,
				LastParkAt: nil,
			},
			lastModified: now,
			expectedText: "Never checked in",
		},
		{
			name: "has uncommitted work - modified after park mtime",
			project: &core.Project{
				IsGrabbed:     true,
				LastParkAt:    &pastDay,
				LastParkMtime: &pastDay,
			},
			lastModified: pastHour,
			expectedText: "Has uncommitted work",
		},
		{
			name: "safe to delete - not modified after park mtime",
			project: &core.Project{
				IsGrabbed:     true,
				LastParkAt:    &now,
				LastParkMtime: &now,
			},
			lastModified: pastHour,
			expectedText: "Safe to delete",
		},
		{
			name: "has uncommitted work - fallback to LastParkAt",
			project: &core.Project{
				IsGrabbed:     true,
				LastParkAt:    &pastDay,
				LastParkMtime: nil, // no mtime set
			},
			lastModified: pastHour,
			expectedText: "Has uncommitted work",
		},
		{
			name: "safe to delete - fallback to LastParkAt",
			project: &core.Project{
				IsGrabbed:     true,
				LastParkAt:    &now,
				LastParkMtime: nil,
			},
			lastModified: pastHour,
			expectedText: "Safe to delete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineStatusInfo(tt.project, tt.lastModified)
			if result.Text != tt.expectedText {
				t.Errorf("determineStatusInfo().Text = %q, want %q", result.Text, tt.expectedText)
			}

			// Also verify emoji is set
			if result.Emoji == "" {
				t.Error("determineStatusInfo().Emoji should not be empty")
			}

			// Verify String() method works correctly
			fullStatus := result.String()
			if fullStatus == "" || len(fullStatus) < len(result.Text) {
				t.Errorf("StatusInfo.String() = %q, should contain emoji and text", fullStatus)
			}
		})
	}
}

func TestDetermineStatus(t *testing.T) {
	now := time.Now()
	pastDay := now.Add(-24 * time.Hour)

	project := &core.Project{
		IsGrabbed:     true,
		LastParkAt:    &now,
		LastParkMtime: &now,
	}

	result := determineStatus(project, pastDay)

	// Should contain both emoji and text
	if result == "" {
		t.Error("determineStatus() should not return empty string")
	}

	// Should contain "Safe to delete" since lastModified is before LastParkMtime
	expected := SymbolCheck + " Safe to delete"
	if result != expected {
		t.Errorf("determineStatus() = %q, want %q", result, expected)
	}
}

func TestStatusInfoString(t *testing.T) {
	tests := []struct {
		name     string
		info     StatusInfo
		expected string
	}{
		{
			name:     "check mark with safe",
			info:     StatusInfo{Emoji: SymbolCheck, Text: "Safe to delete"},
			expected: SymbolCheck + " Safe to delete",
		},
		{
			name:     "warning with uncommitted",
			info:     StatusInfo{Emoji: SymbolWarning, Text: "Has uncommitted work"},
			expected: SymbolWarning + " Has uncommitted work",
		},
		{
			name:     "cross with never checked in",
			info:     StatusInfo{Emoji: SymbolCross, Text: "Never checked in"},
			expected: SymbolCross + " Never checked in",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.info.String()
			if result != tt.expected {
				t.Errorf("StatusInfo.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTimeConstants(t *testing.T) {
	// Verify time constants are set correctly
	if hoursPerDay != 24 {
		t.Errorf("hoursPerDay = %d, want 24", hoursPerDay)
	}
	if daysPerWeek != 7 {
		t.Errorf("daysPerWeek = %d, want 7", daysPerWeek)
	}
	if daysPerMonth != 30 {
		t.Errorf("daysPerMonth = %d, want 30", daysPerMonth)
	}
	if hoursPerWeek != 168 {
		t.Errorf("hoursPerWeek = %d, want 168", hoursPerWeek)
	}
	if hoursPerMonth != 720 {
		t.Errorf("hoursPerMonth = %d, want 720", hoursPerMonth)
	}
}
