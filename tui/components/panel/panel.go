// Package panel provides consistent bordered containers.
package panel

import (
	"strings"

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

// ContentWidth returns the usable width inside a bordered panel.
func ContentWidth(width int) int { return max(1, width-4) }

// ContentHeight returns the usable height inside a bordered panel.
func ContentHeight(height int) int { return max(1, height-2) }

// Render is the value-oriented form used by screens that render a panel for one
// frame rather than retaining a panel model.
func Render(th theme.Theme, title, content string, width, height int, focused bool) string {
	style := th.Panel
	if focused {
		style = style.BorderForeground(th.Palette.Accent)
	}
	m := Model{Title: title, Content: content, Style: style, TitleStyle: th.Title}
	m.SetSize(width, height)
	view := m.View()
	if height <= 0 {
		return view
	}
	rows := strings.Split(view, "\n")
	if len(rows) > height {
		rows = rows[:height]
	}
	for len(rows) < height {
		rows = append(rows, "")
	}
	return strings.Join(rows, "\n")
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
