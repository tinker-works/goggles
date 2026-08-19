// Package markdown renders markdown into terminal-friendly ANSI text.
package markdown

import (
	"strings"
)

// Model stores markdown source and rendering options.
type Model struct {
	Source string
	Width  int
	Style  string
}

// New creates a markdown model.
func New(source string) Model { return Model{Source: source, Style: "dark"} }

// SetWidth sets the markdown word-wrap width.
func (m *Model) SetWidth(width int) { m.Width = width }

// SetSource replaces the markdown source.
func (m *Model) SetSource(source string) { m.Source = source }

// Render returns a terminal-friendly markdown representation. The model keeps
// rendering local so it remains safe to use in a daemon-independent package.
func (m Model) Render() string {
	if strings.TrimSpace(m.Source) == "" {
		return ""
	}
	content := strings.TrimSuffix(m.Source, "\n")
	if m.Width > 0 {
		content = wrap(content, m.Width)
	}
	return content
}

// View implements the conventional component view method.
func (m Model) View() string { return m.Render() }

func wrap(content string, width int) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if len([]rune(line)) <= width {
			continue
		}
		runes := []rune(line)
		parts := make([]string, 0, len(runes)/width+1)
		for len(runes) > width {
			parts = append(parts, string(runes[:width]))
			runes = runes[width:]
		}
		parts = append(parts, string(runes))
		lines[i] = strings.Join(parts, "\n")
	}
	return strings.Join(lines, "\n")
}
