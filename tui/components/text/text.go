// Package text provides width-aware plain text presentation helpers.
package text

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
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
		wrapped := ansi.Hardwrap(line, width, true)
		if newline := strings.IndexByte(wrapped, '\n'); newline >= 0 && lipgloss.Width(wrapped[:newline]) == 0 {
			wrapped = wrapped[:newline] + wrapped[newline+1:]
		}
		lines[i] = wrapped
	}
	return strings.Join(lines, "\n")
}

// Truncate shortens content to a display width, adding tail when needed.
func Truncate(content string, width int) string {
	if width <= 0 {
		return ""
	}
	return ansi.Truncate(content, width, "…")
}

// Lines wraps content into display rows without styling it.
func Lines(content string, width int) []string {
	return strings.Split(Fit(content, width), "\n")
}

// Pad truncates content to width and right-pads it when it is shorter.
func Pad(content string, width int) string {
	content = Truncate(content, width)
	return content + strings.Repeat(" ", max(0, width-lipgloss.Width(content)))
}

// Justify places left and right content on one width-limited row.
func Justify(left, right string, width int) string {
	space := max(1, width-lipgloss.Width(left)-lipgloss.Width(right))
	return left + strings.Repeat(" ", space) + right
}

// Fit keeps every row within width while retaining explicit line breaks.
func Fit(content string, width int) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = Pad(line, width)
	}
	return strings.Join(lines, "\n")
}
