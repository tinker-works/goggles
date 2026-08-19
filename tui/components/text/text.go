// Package text provides width-aware plain text presentation helpers.
package text

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// Model stores text and rendering options.
type Model struct {
	Content string
	Width   int
	Height  int
	Style   lipgloss.Style
}

// New creates a text model.
func New(content string) Model { return Model{Content: content} }

// SetContent replaces the text.
func (m *Model) SetContent(content string) { m.Content = content }

// SetSize sets optional bounds.
func (m *Model) SetSize(width, height int) { m.Width, m.Height = width, height }

// View renders text with optional wrapping and clipping.
func (m Model) View() string {
	style := m.Style
	if m.Width > 0 {
		style = style.Width(m.Width)
	}
	if m.Height > 0 {
		style = style.MaxHeight(m.Height)
	}
	return style.Render(strings.TrimSuffix(m.Content, "\n"))
}

// Wrap wraps each line to width without changing explicit line breaks.
func Wrap(content string, width int) string {
	if width <= 0 {
		return content
	}
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if lipgloss.Width(line) <= width {
			continue
		}
		var wrapped []string
		runes := []rune(line)
		for len(runes) > width {
			wrapped = append(wrapped, string(runes[:width]))
			runes = runes[width:]
		}
		wrapped = append(wrapped, string(runes))
		lines[i] = strings.Join(wrapped, "\n")
	}
	return strings.Join(lines, "\n")
}
