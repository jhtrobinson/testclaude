package core

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Size constants for unit conversions
const (
	Byte     int64 = 1
	Kilobyte       = 1024 * Byte
	Megabyte       = 1024 * Kilobyte
	Gigabyte       = 1024 * Megabyte
	Terabyte       = 1024 * Gigabyte
)

// sizePattern matches human-readable size strings like "10G", "1.5GB", "500M", etc.
var sizePattern = regexp.MustCompile(`(?i)^(\d+(?:\.\d+)?)\s*([KMGT]B?)$`)

// ParseSize converts a human-readable size string to bytes.
// Supported formats: 10G, 500M, 2T, 1.5GB, 100MB, 1024K, etc.
// Units are case insensitive: G/GB, M/MB, K/KB, T/TB
// Returns an error for invalid formats, negative, or zero values.
func ParseSize(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(sizeStr)
	if sizeStr == "" {
		return 0, fmt.Errorf("empty size string")
	}

	matches := sizePattern.FindStringSubmatch(sizeStr)
	if matches == nil {
		return 0, fmt.Errorf("invalid size format: %q (expected format like 10G, 500M, 1.5GB)", sizeStr)
	}

	valueStr := matches[1]
	unit := strings.ToUpper(matches[2])

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric value: %q", valueStr)
	}

	if value <= 0 {
		return 0, fmt.Errorf("size must be positive: %v", value)
	}

	var multiplier int64
	switch unit {
	case "K", "KB":
		multiplier = Kilobyte
	case "M", "MB":
		multiplier = Megabyte
	case "G", "GB":
		multiplier = Gigabyte
	case "T", "TB":
		multiplier = Terabyte
	default:
		return 0, fmt.Errorf("unsupported unit: %q", unit)
	}

	bytes := int64(value * float64(multiplier))
	if bytes <= 0 {
		return 0, fmt.Errorf("calculated size must be positive")
	}

	return bytes, nil
}

// FormatSizeCompact converts bytes to a compact human-readable size string.
// Uses the largest appropriate unit (TB, GB, MB, KB, or bytes).
// The output uses the short form (G, M, K, T) for consistency with ParseSize.
func FormatSizeCompact(bytes int64) string {
	if bytes < 0 {
		return fmt.Sprintf("%dB", bytes)
	}

	if bytes == 0 {
		return "0B"
	}

	// Use float for calculation to handle decimal values
	size := float64(bytes)

	switch {
	case bytes >= Terabyte:
		value := size / float64(Terabyte)
		return formatValue(value, "T")
	case bytes >= Gigabyte:
		value := size / float64(Gigabyte)
		return formatValue(value, "G")
	case bytes >= Megabyte:
		value := size / float64(Megabyte)
		return formatValue(value, "M")
	case bytes >= Kilobyte:
		value := size / float64(Kilobyte)
		return formatValue(value, "K")
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

// formatValue formats a float value with its unit, removing unnecessary decimal places.
func formatValue(value float64, unit string) string {
	// If it's a whole number, format without decimals
	if value == float64(int64(value)) {
		return fmt.Sprintf("%d%s", int64(value), unit)
	}

	// Otherwise, format with up to 2 decimal places, trimming trailing zeros
	formatted := fmt.Sprintf("%.2f", value)
	formatted = strings.TrimRight(formatted, "0")
	formatted = strings.TrimRight(formatted, ".")
	return formatted + unit
}

// MustParseSize is like ParseSize but panics on error.
// Useful for initializing constants or in tests.
func MustParseSize(sizeStr string) int64 {
	bytes, err := ParseSize(sizeStr)
	if err != nil {
		panic(fmt.Sprintf("failed to parse size %q: %v", sizeStr, err))
	}
	return bytes
}
