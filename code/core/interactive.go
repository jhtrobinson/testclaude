package core

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

// InteractiveSelector allows users to interactively select projects for pruning
type InteractiveSelector struct {
	candidates    []PruneCandidate
	cursor        int
	selected      map[int]bool
	targetBytes   int64
	totalSelected int64
	quitting      bool
	confirmed     bool
}

// NewInteractiveSelector creates a new interactive selector
func NewInteractiveSelector(candidates []PruneCandidate, targetBytes int64) *InteractiveSelector {
	selected := make(map[int]bool)
	var totalSelected int64

	// Pre-select candidates that were auto-selected
	for i, c := range candidates {
		if c.Selected {
			selected[i] = true
			totalSelected += c.LocalSize
		}
	}

	return &InteractiveSelector{
		candidates:    candidates,
		cursor:        0,
		selected:      selected,
		targetBytes:   targetBytes,
		totalSelected: totalSelected,
	}
}

// termios structure for terminal settings
type termios struct {
	Iflag  uint32
	Oflag  uint32
	Cflag  uint32
	Lflag  uint32
	Cc     [20]byte
	Ispeed uint32
	Ospeed uint32
}

// getTermios gets the current terminal settings
func getTermios(fd int) (*termios, error) {
	var t termios
	// Use TIOCGETA on macOS (Darwin), TCGETS on Linux
	const TIOCGETA = 0x40487413 // macOS
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), TIOCGETA, uintptr(unsafe.Pointer(&t)))
	if err != 0 {
		return nil, err
	}
	return &t, nil
}

// setTermios sets the terminal settings
func setTermios(fd int, t *termios) error {
	// Use TIOCSETA on macOS (Darwin), TCSETS on Linux
	const TIOCSETA = 0x80487414 // macOS
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), TIOCSETA, uintptr(unsafe.Pointer(t)))
	if err != 0 {
		return err
	}
	return nil
}

// makeRaw puts the terminal into raw mode
func makeRaw(fd int) (*termios, error) {
	old, err := getTermios(fd)
	if err != nil {
		return nil, err
	}

	newT := *old
	// Turn off echo and canonical mode
	newT.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	newT.Iflag &^= syscall.BRKINT | syscall.ICRNL | syscall.INPCK | syscall.ISTRIP | syscall.IXON
	newT.Cflag &^= syscall.CSIZE | syscall.PARENB
	newT.Cflag |= syscall.CS8
	newT.Oflag &^= syscall.OPOST
	newT.Cc[syscall.VMIN] = 1
	newT.Cc[syscall.VTIME] = 0

	if err := setTermios(fd, &newT); err != nil {
		return nil, err
	}

	return old, nil
}

// isTerminal checks if fd is a terminal
func isTerminal(fd int) bool {
	_, err := getTermios(fd)
	return err == nil
}

// clearScreen clears the terminal screen
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

// render displays the current state of the selector
func (m *InteractiveSelector) render() {
	var b strings.Builder

	// Header
	b.WriteString(fmt.Sprintf("Need to free up %s. Select projects to delete:\n\n", FormatSize(m.targetBytes)))

	// List candidates
	for i, c := range m.candidates {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := "[ ]"
		if m.selected[i] {
			checked = "[x]"
		}

		sizeStr := FormatSize(c.LocalSize)
		ageStr := formatAge(c.LastModified)

		b.WriteString(fmt.Sprintf("%s %s %s (%s) - %s\n", cursor, checked, c.Name, sizeStr, ageStr))
	}

	// Footer with running total
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Selected: %s / Target: %s", FormatSize(m.totalSelected), FormatSize(m.targetBytes)))

	if m.totalSelected >= m.targetBytes {
		b.WriteString(" (target reached)")
	} else if m.totalSelected > 0 {
		remaining := m.targetBytes - m.totalSelected
		b.WriteString(fmt.Sprintf(" (need %s more)", FormatSize(remaining)))
	}

	b.WriteString("\n\n")
	b.WriteString("Controls: space=toggle  a=select all  enter=confirm  q=quit\n")

	fmt.Print(b.String())
}

// handleInput processes a single keypress
func (m *InteractiveSelector) handleInput(key byte) bool {
	switch key {
	case 'q', 27: // q or ESC
		m.quitting = true
		return false

	case 'k', 'A': // k or up arrow (part of escape sequence)
		if m.cursor > 0 {
			m.cursor--
		}

	case 'j', 'B': // j or down arrow (part of escape sequence)
		if m.cursor < len(m.candidates)-1 {
			m.cursor++
		}

	case ' ': // Space to toggle
		if m.cursor < len(m.candidates) {
			size := m.candidates[m.cursor].LocalSize
			if m.selected[m.cursor] {
				delete(m.selected, m.cursor)
				m.totalSelected -= size
			} else {
				m.selected[m.cursor] = true
				m.totalSelected += size
			}
		}

	case 'a': // Select all / deselect all
		if len(m.selected) == len(m.candidates) {
			// Deselect all
			m.selected = make(map[int]bool)
			m.totalSelected = 0
		} else {
			// Select all
			m.selected = make(map[int]bool)
			m.totalSelected = 0
			for i, c := range m.candidates {
				m.selected[i] = true
				m.totalSelected += c.LocalSize
			}
		}

	case '\r', '\n': // Enter
		m.confirmed = true
		return false
	}

	return true
}

// GetSelected returns the selected candidates
func (m *InteractiveSelector) GetSelected() []PruneCandidate {
	var selected []PruneCandidate
	for i, c := range m.candidates {
		if m.selected[i] {
			selected = append(selected, c)
		}
	}
	return selected
}

// WasConfirmed returns true if the user confirmed their selection
func (m *InteractiveSelector) WasConfirmed() bool {
	return m.confirmed
}

// WasQuit returns true if the user quit without confirming
func (m *InteractiveSelector) WasQuit() bool {
	return m.quitting
}

// TotalSelected returns the total size of selected projects
func (m *InteractiveSelector) TotalSelected() int64 {
	return m.totalSelected
}

// RunInteractiveSelection runs the interactive selection UI
func RunInteractiveSelection(candidates []PruneCandidate, targetBytes int64) (*InteractiveSelector, error) {
	selector := NewInteractiveSelector(candidates, targetBytes)

	// Check if stdin is a terminal
	if !isTerminal(int(os.Stdin.Fd())) {
		return nil, fmt.Errorf("interactive mode requires a terminal")
	}

	// Set terminal to raw mode
	oldState, err := makeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return nil, fmt.Errorf("failed to set raw mode: %w", err)
	}
	defer setTermios(int(os.Stdin.Fd()), oldState)

	// Clear screen and hide cursor
	clearScreen()
	fmt.Print("\033[?25l") // Hide cursor
	defer fmt.Print("\033[?25h") // Show cursor on exit

	reader := bufio.NewReader(os.Stdin)

	for {
		// Clear screen and render
		clearScreen()
		selector.render()

		// Read single character
		char, err := reader.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("failed to read input: %w", err)
		}

		// Handle escape sequences (arrow keys)
		if char == 27 { // ESC
			// Check if there are more bytes (arrow key sequence)
			if reader.Buffered() > 0 {
				next, _ := reader.ReadByte()
				if next == '[' {
					// Arrow key sequence
					arrow, _ := reader.ReadByte()
					if !selector.handleInput(arrow) {
						break
					}
					continue
				}
			}
			// Plain ESC key - quit
			if !selector.handleInput(char) {
				break
			}
			continue
		}

		if !selector.handleInput(char) {
			break
		}
	}

	// Clear the selection UI
	clearScreen()

	return selector, nil
}

// formatAge formats the age of a project relative to now
func formatAge(t time.Time) string {
	if t.IsZero() {
		return "never"
	}

	duration := time.Since(t)

	const (
		hoursPerDay   = 24
		hoursPerWeek  = 24 * 7
		hoursPerMonth = 24 * 30
	)

	if duration < time.Minute {
		return "just now"
	}
	if duration < time.Hour {
		mins := int(duration.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", mins)
	}
	if duration < hoursPerDay*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	if duration < hoursPerWeek*time.Hour {
		days := int(duration.Hours() / hoursPerDay)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
	if duration < hoursPerMonth*time.Hour {
		weeks := int(duration.Hours() / hoursPerWeek)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	}

	months := int(duration.Hours() / hoursPerMonth)
	if months == 1 {
		return "1 month ago"
	}
	return fmt.Sprintf("%d months ago", months)
}
