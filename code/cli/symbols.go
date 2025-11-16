package cli

import (
	"os"
	"strings"
)

// Symbol constants with Unicode and ASCII fallbacks
var (
	SymbolCheck   = "✓"
	SymbolWarning = "⚠"
	SymbolCross   = "✗"
)

func init() {
	// Check if ASCII mode is requested via environment variable
	if os.Getenv("PARKR_ASCII") == "1" || os.Getenv("PARKR_ASCII") == "true" {
		useASCIISymbols()
		return
	}

	// Check if terminal might not support Unicode
	// Common indicators: TERM=dumb, or certain legacy terminals
	term := os.Getenv("TERM")
	if term == "dumb" || term == "" {
		useASCIISymbols()
		return
	}

	// Check for LC_ALL or LANG - if not set to UTF-8, use ASCII
	lang := os.Getenv("LANG")
	lcAll := os.Getenv("LC_ALL")

	// If explicitly set to non-UTF8 locale, use ASCII
	if lcAll != "" && !strings.Contains(strings.ToUpper(lcAll), "UTF") {
		useASCIISymbols()
		return
	}

	if lang != "" && !strings.Contains(strings.ToUpper(lang), "UTF") && lcAll == "" {
		useASCIISymbols()
		return
	}
}

// useASCIISymbols switches to ASCII-compatible symbols
func useASCIISymbols() {
	SymbolCheck = "[OK]"
	SymbolWarning = "[!]"
	SymbolCross = "[X]"
}

// SetASCIIMode allows programmatic switching to ASCII mode (useful for testing)
func SetASCIIMode(ascii bool) {
	if ascii {
		useASCIISymbols()
	} else {
		SymbolCheck = "✓"
		SymbolWarning = "⚠"
		SymbolCross = "✗"
	}
}
