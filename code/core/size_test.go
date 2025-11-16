package core

import (
	"strings"
	"testing"
)

func TestParseSize_ValidInputs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		// Kilobytes
		{"simple K", "1024K", 1024 * Kilobyte},
		{"simple KB", "1024KB", 1024 * Kilobyte},
		{"lowercase k", "100k", 100 * Kilobyte},
		{"lowercase kb", "100kb", 100 * Kilobyte},
		{"decimal K", "1.5K", int64(1.5 * float64(Kilobyte))},

		// Megabytes
		{"simple M", "500M", 500 * Megabyte},
		{"simple MB", "100MB", 100 * Megabyte},
		{"lowercase m", "250m", 250 * Megabyte},
		{"lowercase mb", "250mb", 250 * Megabyte},
		{"decimal M", "2.5M", int64(2.5 * float64(Megabyte))},

		// Gigabytes
		{"simple G", "10G", 10 * Gigabyte},
		{"simple GB", "1.5GB", int64(1.5 * float64(Gigabyte))},
		{"lowercase g", "5g", 5 * Gigabyte},
		{"lowercase gb", "5gb", 5 * Gigabyte},
		{"decimal G", "1.5G", int64(1.5 * float64(Gigabyte))},
		{"large G", "100G", 100 * Gigabyte},

		// Terabytes
		{"simple T", "2T", 2 * Terabyte},
		{"simple TB", "1TB", 1 * Terabyte},
		{"lowercase t", "1t", 1 * Terabyte},
		{"lowercase tb", "1tb", 1 * Terabyte},
		{"decimal T", "1.5T", int64(1.5 * float64(Terabyte))},

		// Edge cases
		{"minimum K", "1K", 1 * Kilobyte},
		{"small decimal", "0.5G", int64(0.5 * float64(Gigabyte))},
		{"with spaces", "  10G  ", 10 * Gigabyte},
		{"space between", "10 G", 10 * Gigabyte},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseSize(tt.input)
			if err != nil {
				t.Errorf("ParseSize(%q) returned unexpected error: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("ParseSize(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseSize_InvalidFormats(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		errContains string
	}{
		{"empty string", "", "empty"},
		{"just number", "100", "invalid size format"},
		{"just unit", "GB", "invalid size format"},
		{"invalid unit", "100X", "invalid size format"},
		{"invalid unit bytes", "100B", "invalid size format"},
		{"negative value", "-10G", "invalid size format"},
		{"double negative", "--10G", "invalid size format"},
		{"letters in number", "10aG", "invalid size format"},
		{"multiple decimals", "1.2.3G", "invalid size format"},
		{"no number", "G", "invalid size format"},
		{"special chars", "10@G", "invalid size format"},
		{"percent", "10%G", "invalid size format"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseSize(tt.input)
			if err == nil {
				t.Errorf("ParseSize(%q) expected error but got none", tt.input)
				return
			}
			if !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("ParseSize(%q) error = %q, want error containing %q", tt.input, err.Error(), tt.errContains)
			}
		})
	}
}

func TestParseSize_ZeroValues(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"zero G", "0G"},
		{"zero GB", "0GB"},
		{"zero M", "0M"},
		{"zero MB", "0MB"},
		{"zero K", "0K"},
		{"zero KB", "0KB"},
		{"zero T", "0T"},
		{"zero TB", "0TB"},
		{"zero decimal", "0.0G"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseSize(tt.input)
			if err == nil {
				t.Errorf("ParseSize(%q) expected error for zero value but got none", tt.input)
				return
			}
			if !strings.Contains(err.Error(), "positive") {
				t.Errorf("ParseSize(%q) error = %q, want error about positive value", tt.input, err.Error())
			}
		})
	}
}

func TestFormatSizeCompact(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		// Bytes
		{"zero bytes", 0, "0B"},
		{"single byte", 1, "1B"},
		{"few bytes", 100, "100B"},
		{"max bytes before K", 1023, "1023B"},

		// Kilobytes
		{"exact 1K", Kilobyte, "1K"},
		{"exact 10K", 10 * Kilobyte, "10K"},
		{"exact 1023K", 1023 * Kilobyte, "1023K"},
		{"decimal K", int64(1.5 * float64(Kilobyte)), "1.5K"},

		// Megabytes
		{"exact 1M", Megabyte, "1M"},
		{"exact 100M", 100 * Megabyte, "100M"},
		{"exact 500M", 500 * Megabyte, "500M"},
		{"decimal M", int64(2.5 * float64(Megabyte)), "2.5M"},

		// Gigabytes
		{"exact 1G", Gigabyte, "1G"},
		{"exact 10G", 10 * Gigabyte, "10G"},
		{"exact 100G", 100 * Gigabyte, "100G"},
		{"decimal G", int64(1.5 * float64(Gigabyte)), "1.5G"},

		// Terabytes
		{"exact 1T", Terabyte, "1T"},
		{"exact 2T", 2 * Terabyte, "2T"},
		{"decimal T", int64(1.5 * float64(Terabyte)), "1.5T"},

		// Negative (edge case - shouldn't normally occur)
		{"negative", -100, "-100B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatSizeCompact(tt.input)
			if result != tt.expected {
				t.Errorf("FormatSizeCompact(%d) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseSizeAndFormatSizeCompact_RoundTrip(t *testing.T) {
	// Test that parsing and formatting are consistent
	tests := []struct {
		input    string
		expected string
	}{
		{"1K", "1K"},
		{"10K", "10K"},
		{"1M", "1M"},
		{"100M", "100M"},
		{"1G", "1G"},
		{"10G", "10G"},
		{"1T", "1T"},
		{"2T", "2T"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			bytes, err := ParseSize(tt.input)
			if err != nil {
				t.Errorf("ParseSize(%q) returned error: %v", tt.input, err)
				return
			}
			result := FormatSizeCompact(bytes)
			if result != tt.expected {
				t.Errorf("FormatSizeCompact(ParseSize(%q)) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMustParseSize_Success(t *testing.T) {
	// Should not panic for valid input
	result := MustParseSize("10G")
	expected := int64(10 * Gigabyte)
	if result != expected {
		t.Errorf("MustParseSize(\"10G\") = %d, want %d", result, expected)
	}
}

func TestMustParseSize_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParseSize with invalid input should panic")
		}
	}()
	MustParseSize("invalid")
}

func TestSizeConstants(t *testing.T) {
	// Verify size constants are correct
	if Kilobyte != 1024 {
		t.Errorf("Kilobyte = %d, want 1024", Kilobyte)
	}
	if Megabyte != 1024*1024 {
		t.Errorf("Megabyte = %d, want %d", Megabyte, 1024*1024)
	}
	if Gigabyte != 1024*1024*1024 {
		t.Errorf("Gigabyte = %d, want %d", Gigabyte, 1024*1024*1024)
	}
	if Terabyte != 1024*1024*1024*1024 {
		t.Errorf("Terabyte = %d, want %d", Terabyte, 1024*1024*1024*1024)
	}
}

func TestParseSize_CaseInsensitivity(t *testing.T) {
	// All these should parse to the same value
	expected := int64(10 * Gigabyte)
	inputs := []string{"10g", "10G", "10gb", "10GB", "10Gb", "10gB"}

	for _, input := range inputs {
		result, err := ParseSize(input)
		if err != nil {
			t.Errorf("ParseSize(%q) returned error: %v", input, err)
			continue
		}
		if result != expected {
			t.Errorf("ParseSize(%q) = %d, want %d", input, result, expected)
		}
	}
}

func TestFormatSizeCompact_Precision(t *testing.T) {
	// Test that formatting handles precision correctly
	// Use runtime computation to avoid compile-time constant truncation issues
	gb := float64(Gigabyte)
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		// Values that result in clean decimals
		{"1.25G", int64(gb * 1.25), "1.25G"},
		{"1.1G", int64(gb * 1.1), "1.1G"},
		{"1.01G", int64(gb * 1.01), "1.01G"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatSizeCompact(tt.bytes)
			if result != tt.expected {
				t.Errorf("FormatSizeCompact(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestParseSize_Overflow(t *testing.T) {
	// Test that very large values are rejected to prevent overflow
	tests := []struct {
		name        string
		input       string
		errContains string
	}{
		{"overflow T", "9999999999999999T", "overflow"},
		{"overflow G", "9999999999999999G", "overflow"},
		{"overflow M", "9999999999999999999M", "overflow"},
		{"overflow K", "9999999999999999999999K", "overflow"},
		// Max int64 is 9,223,372,036,854,775,807
		// Max TB is ~8388607 (8.3 million TB)
		{"near max T", "10000000T", "overflow"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseSize(tt.input)
			if err == nil {
				t.Errorf("ParseSize(%q) expected overflow error but got none", tt.input)
				return
			}
			if !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("ParseSize(%q) error = %q, want error containing %q", tt.input, err.Error(), tt.errContains)
			}
		})
	}
}

func TestParseSize_LargeValidValues(t *testing.T) {
	// Test that large but valid values work correctly
	tests := []struct {
		name  string
		input string
	}{
		{"8000T", "8000T"},
		{"8000000G", "8000000G"},
		{"8000000000M", "8000000000M"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseSize(tt.input)
			if err != nil {
				t.Errorf("ParseSize(%q) returned unexpected error: %v", tt.input, err)
				return
			}
			if result <= 0 {
				t.Errorf("ParseSize(%q) = %d, want positive value", tt.input, result)
			}
		})
	}
}
