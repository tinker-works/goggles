// Package commentbox renders a bordered block of user-provided text.
package commentbox

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/tinker-works/goggles/tui/theme"
)

// Model is a static comment presentation primitive. It deliberately accepts
// text rather than a domain comment type.
type Model struct {
	Author    string
	Timestamp string
	Body      string
	Width     int
	Style     lipgloss.Style
	Header    lipgloss.Style
}

// New creates a comment box using the default theme.
func New() Model {
	t := theme.Default
	return Model{
		Style:  t.Panel,
		Header: t.Subtitle.Bold(true),
	}
}

// SetContent sets the presentation fields.
func (m *Model) SetContent(author, timestamp, body string) {
	m.Author, m.Timestamp, m.Body = author, timestamp, body
}

// SetWidth sets the outer width. A non-positive value leaves width automatic.
func (m *Model) SetWidth(width int) { m.Width = width }

// View renders the comment.
func (m Model) View() string {
	header := strings.TrimSpace(strings.Join([]string{m.Author, m.Timestamp}, "  "))
	content := m.Body
	if header != "" {
		content = m.Header.Render(header) + "\n" + content
	}
	style := m.Style
	if m.Width > 0 {
		style = style.Width(m.Width)
	}
	return style.Render(content)
}
