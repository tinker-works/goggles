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
