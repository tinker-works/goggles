// Package panel provides consistent bordered containers.
package panel

import (
	"charm.land/lipgloss/v2"
	"github.com/tinker-works/goggles/tui/theme"
)

// Model is a bordered content panel.
type Model struct {
	Content    string
	Title      string
	Width      int
	Height     int
	Style      lipgloss.Style
	TitleStyle lipgloss.Style
}

// New creates a panel with default theme styles.
func New(title string) Model {
	t := theme.Default()
	return Model{Title: title, Style: t.Panel, TitleStyle: t.Title}
}

// SetContent updates the panel body.
func (m *Model) SetContent(content string) { m.Content = content }

// SetSize updates the panel size.
func (m *Model) SetSize(width, height int) { m.Width, m.Height = width, height }

// View renders the panel.
func (m Model) View() string {
	content := m.Content
	if m.Title != "" {
		content = m.TitleStyle.Render(m.Title) + "\n" + content
	}
	style := m.Style
	if m.Width > 0 {
		style = style.Width(m.Width)
	}
	if m.Height > 0 {
		style = style.Height(m.Height)
	}
	return style.Render(content)
}
