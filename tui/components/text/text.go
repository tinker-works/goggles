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

// Pad truncates and right-pads content to a display width.
func Pad(content string, width int) string {
	if width <= 0 {
		return ""
	}
	content = Truncate(content, width)
	if gap := width - lipgloss.Width(content); gap > 0 {
		return content + strings.Repeat(" ", gap)
	}
	return content
}

// Fit makes every line the same display width for horizontal composition.
func Fit(content string, width int) string {
	lines := strings.Split(content, "\n")
	for i := range lines {
		lines[i] = Pad(lines[i], width)
	}
	return strings.Join(lines, "\n")
}

// FitHeight returns exactly height rows, clipping only at the bottom.
func FitHeight(content string, height int) string {
	if height <= 0 {
		return ""
	}
	lines := strings.Split(content, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	for len(lines) < height {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

// Lines splits content and pads each row for a rectangular panel body.
func Lines(content string, width int) []string {
	return strings.Split(Fit(content, width), "\n")
}

// Justify puts left and right at opposite ends of a row.
func Justify(left, right string, width int) string {
	return left + strings.Repeat(" ", max(1, width-lipgloss.Width(left)-lipgloss.Width(right))) + right
}
