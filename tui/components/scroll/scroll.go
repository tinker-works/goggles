// Package scroll wraps bubbles' viewport with predictable size and content APIs.
package scroll

import (
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
)

// Clamp keeps a scroll offset within the content and viewport bounds.
func Clamp(total, window, top int) int {
	if total <= 0 || window <= 0 {
		return 0
	}
	return max(0, min(top, max(0, total-window)))
}

// Follow keeps selected at least inside the viewport, preferring the tail when
// there is no selection.
func Follow(total, window, selected int) int {
	if window <= 0 || total <= window {
		return 0
	}
	return Clamp(total, window, selected-window/2)
}

// Window returns a bounded slice of rows.
func Window(rows []string, window, top int) []string {
	if window <= 0 || len(rows) == 0 {
		return nil
	}
	top = max(0, min(top, max(0, len(rows)-1)))
	return append([]string(nil), rows[top:min(len(rows), top+window)]...)
}

// Mark indicates whether content continues above or below the viewport.
func Mark(total, window, top int) string {
	if total <= window {
		return ""
	}
	end := total - window
	switch {
	case top == 0:
		return " ↓"
	case top >= end:
		return " ↑"
	case end > 0:
		return " ↕"
	default:
		return ""
	}
}

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
