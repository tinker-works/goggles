// Package markdown renders markdown into terminal-friendly ANSI text.
package markdown

import (
	"strings"

	"github.com/charmbracelet/glamour"
)

// Model stores markdown source and rendering options.
type Model struct {
	Source string
	Width  int
	Style  string
}

// New creates a markdown model. An omitted source creates an empty reusable
// renderer for callers that set each body before rendering.
func New(sources ...string) Model {
	source := ""
	if len(sources) > 0 {
		source = sources[0]
	}
	return Model{Source: source, Style: "dark"}
}

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

	style := m.Style
	if style == "" {
		style = "dark"
	}
	options := []glamour.TermRendererOption{glamour.WithStandardStyle(style)}
	if m.Width > 0 {
		options = append(options, glamour.WithWordWrap(m.Width))
	}
	renderer, err := glamour.NewTermRenderer(options...)
	if err != nil {
		return strings.TrimSuffix(m.Source, "\n")
	}
	content, err := renderer.Render(m.Source)
	if err != nil {
		return strings.TrimSuffix(m.Source, "\n")
	}
	return content
}

// View implements the conventional component view method.
func (m Model) View() string { return m.Render() }
