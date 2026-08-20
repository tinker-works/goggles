// Package panel provides consistent bordered containers.
package panel

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/tinker-works/goggles/tui/components/text"
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

// ContentWidth is the usable width inside a bordered panel.
func ContentWidth(width int) int { return max(1, width-4) }

// ContentHeight is the usable height inside a bordered panel.
func ContentHeight(height int) int { return max(1, height-2) }

// Render draws a fixed-size rounded panel used by the migrated screens.
func Render(th theme.Theme, title, content string, width, height int, focused bool) string {
	width = max(6, width)
	inner := ContentWidth(width)
	border := th.Border
	if focused {
		border = th.Selected
	}
	rows := text.Lines(content, inner)
	if height > 2 {
		for len(rows) < height-2 {
			rows = append(rows, strings.Repeat(" ", inner))
		}
	}
	top := "╭" + strings.Repeat("─", width-2) + "╮"
	if title != "" {
		caption := text.Truncate(title, max(1, width-6))
		trailing := max(0, width-3-lipgloss.Width(caption))
		top = "╭─ " + th.Accent.Render(caption) + " " + strings.Repeat("─", trailing) + "╮"
	}
	lines := make([]string, 0, len(rows)+2)
	lines = append(lines, border.Render(top))
	for _, row := range rows {
		lines = append(lines, border.Render("│")+" "+row+" "+border.Render("│"))
	}
	lines = append(lines, border.Render("╰"+strings.Repeat("─", width-2)+"╯"))
	return strings.Join(lines, "\n")
}
