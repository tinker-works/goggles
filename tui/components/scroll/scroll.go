// Package scroll wraps bubbles' viewport with predictable size and content APIs.
package scroll

import (
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
)

// Model is a scrollable text view.
type Model struct {
	Viewport viewport.Model
}

// New creates a viewport with dimensions.
func New(width, height int) Model {
	return Model{Viewport: viewport.New(viewport.WithWidth(width), viewport.WithHeight(height))}
}

// SetSize changes viewport dimensions.
func (m *Model) SetSize(width, height int) {
	m.Viewport.SetWidth(width)
	m.Viewport.SetHeight(height)
}

// SetContent replaces the scrollable content.
func (m *Model) SetContent(content string) { m.Viewport.SetContent(content) }

// ScrollPercent returns the current vertical scroll percentage.
func (m Model) ScrollPercent() float64 { return m.Viewport.ScrollPercent() }

// Update handles keyboard and mouse scrolling.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.Viewport, cmd = m.Viewport.Update(msg)
	return m, cmd
}

// View renders the viewport.
func (m Model) View() string { return m.Viewport.View() }

// Clamp keeps an offset inside a window over total rows.
func Clamp(total, window, offset int) int {
	if total <= 0 || window <= 0 {
		return 0
	}
	return max(0, min(offset, max(0, total-window)))
}

// Window returns the rows visible from offset.
func Window(rows []string, window, offset int) []string {
	if window <= 0 || len(rows) == 0 {
		return nil
	}
	start := max(0, min(offset, max(0, len(rows)-1)))
	end := min(len(rows), start+window)
	return append([]string(nil), rows[start:end]...)
}

// Follow derives an offset that keeps index visible in the window.
func Follow(total, window, index int) int {
	if window <= 0 || total <= window {
		return 0
	}
	return Clamp(total, window, index-window/2)
}

// Mark describes the direction of rows hidden by a window.
func Mark(total, window, offset int) string {
	if total <= window {
		return ""
	}
	switch {
	case offset == 0:
		return " ↓"
	case offset >= total-window:
		return " ↑"
	default:
		return " ↕"
	}
}
