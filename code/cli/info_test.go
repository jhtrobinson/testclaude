package cli

import "testing"

func TestBoolToYesNo(t *testing.T) {
	tests := []struct {
		name     string
		input    bool
		expected string
	}{
		{"true returns Yes", true, "Yes"},
		{"false returns No", false, "No"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := boolToYesNo(tt.input)
			if result != tt.expected {
				t.Errorf("boolToYesNo(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
