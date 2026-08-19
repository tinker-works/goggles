// Package statusline renders a one-line left/right status bar.
package statusline

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// Model stores statusline content.
type Model struct {
	Left       string
	Right      string
	Width      int
	LeftStyle  lipgloss.Style
	RightStyle lipgloss.Style
}

// New creates a statusline of width.
func New(width int) Model { return Model{Width: width} }

// Set updates the two sides.
func (m *Model) Set(left, right string) { m.Left, m.Right = left, right }

// View renders both sides without allowing the right side to overlap the left.
func (m Model) View() string {
	left, right := m.LeftStyle.Render(m.Left), m.RightStyle.Render(m.Right)
	if m.Width <= 0 {
		return strings.TrimSpace(left + " " + right)
	}
	available := m.Width - lipgloss.Width(left) - lipgloss.Width(right)
	if available < 1 {
		return lipgloss.NewStyle().Width(m.Width).MaxWidth(m.Width).Render(left)
	}
	return left + strings.Repeat(" ", available) + right
}
