package cli

import "testing"

func TestSetASCIIMode(t *testing.T) {
	// Save original values
	origCheck := SymbolCheck
	origWarning := SymbolWarning
	origCross := SymbolCross
	defer func() {
		// Restore original values after test
		SymbolCheck = origCheck
		SymbolWarning = origWarning
		SymbolCross = origCross
	}()

	t.Run("ASCII mode sets ASCII symbols", func(t *testing.T) {
		SetASCIIMode(true)

		if SymbolCheck != "[OK]" {
			t.Errorf("SymbolCheck = %q, want %q", SymbolCheck, "[OK]")
		}
		if SymbolWarning != "[!]" {
			t.Errorf("SymbolWarning = %q, want %q", SymbolWarning, "[!]")
		}
		if SymbolCross != "[X]" {
			t.Errorf("SymbolCross = %q, want %q", SymbolCross, "[X]")
		}
	})

	t.Run("Unicode mode restores Unicode symbols", func(t *testing.T) {
		SetASCIIMode(false)

		if SymbolCheck != "✓" {
			t.Errorf("SymbolCheck = %q, want %q", SymbolCheck, "✓")
		}
		if SymbolWarning != "⚠" {
			t.Errorf("SymbolWarning = %q, want %q", SymbolWarning, "⚠")
		}
		if SymbolCross != "✗" {
			t.Errorf("SymbolCross = %q, want %q", SymbolCross, "✗")
		}
	})
}

func TestUseASCIISymbols(t *testing.T) {
	// Save original values
	origCheck := SymbolCheck
	origWarning := SymbolWarning
	origCross := SymbolCross
	defer func() {
		// Restore original values after test
		SymbolCheck = origCheck
		SymbolWarning = origWarning
		SymbolCross = origCross
	}()

	useASCIISymbols()

	if SymbolCheck != "[OK]" {
		t.Errorf("SymbolCheck = %q, want %q", SymbolCheck, "[OK]")
	}
	if SymbolWarning != "[!]" {
		t.Errorf("SymbolWarning = %q, want %q", SymbolWarning, "[!]")
	}
	if SymbolCross != "[X]" {
		t.Errorf("SymbolCross = %q, want %q", SymbolCross, "[X]")
	}
}

func TestSymbolsIntegration(t *testing.T) {
	// Save original values
	origCheck := SymbolCheck
	origWarning := SymbolWarning
	origCross := SymbolCross
	defer func() {
		// Restore original values after test
		SymbolCheck = origCheck
		SymbolWarning = origWarning
		SymbolCross = origCross
	}()

	t.Run("StatusInfo uses configurable symbols in ASCII mode", func(t *testing.T) {
		SetASCIIMode(true)

		info := StatusInfo{Emoji: SymbolCheck, Text: "Safe to delete"}
		result := info.String()
		expected := "[OK] Safe to delete"

		if result != expected {
			t.Errorf("StatusInfo.String() = %q, want %q", result, expected)
		}
	})

	t.Run("StatusInfo uses configurable symbols in Unicode mode", func(t *testing.T) {
		SetASCIIMode(false)

		info := StatusInfo{Emoji: SymbolCheck, Text: "Safe to delete"}
		result := info.String()
		expected := "✓ Safe to delete"

		if result != expected {
			t.Errorf("StatusInfo.String() = %q, want %q", result, expected)
		}
	})
}
