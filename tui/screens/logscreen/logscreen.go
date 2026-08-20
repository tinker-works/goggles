// Package logscreen renders the daemon log returned by bounded offset polls.
package logscreen

import (
	"strings"

	"github.com/tinker-works/goggles/tui/components/panel"
	"github.com/tinker-works/goggles/tui/components/statusline"
	"github.com/tinker-works/goggles/tui/components/text"
	"github.com/tinker-works/goggles/tui/theme"
)

type Model struct {
	Lines    []string
	Offset   int64
	Err      error
	top      int
	detached bool
}

// maxLogLines bounds the local screen buffer; each daemon response is already
// bounded independently by actions.ReadLog.
const maxLogLines = 10000

func New() Model { return Model{} }

func (m Model) Append(lines []string, next int64, err error) Model {
	m.Err = err
	if err != nil {
		return m
	}
	m.Lines = append(m.Lines, lines...)
	if dropped := len(m.Lines) - maxLogLines; dropped > 0 {
		m.Lines = m.Lines[dropped:]
		if m.detached {
			m.top = max(0, m.top-dropped)
		}
	}
	m.Offset = next
	return m
}

func (m Model) Reload() Model {
	m.Lines, m.Offset, m.Err = nil, 0, nil
	m.top, m.detached = 0, false
	return m
}

func (m Model) Scroll(rows, total, window int) Model {
	end := max(0, total-window)
	next := max(0, min(m.logTop(total, window)+rows, end))
	m.top, m.detached = next, next < end
	return m
}

func (m Model) logTop(total, window int) int {
	end := max(0, total-window)
	if !m.detached {
		return end
	}
	return min(m.top, end)
}

func (m Model) Rows(width int) []string {
	inner := panel.ContentWidth(width)
	var rows []string
	for _, line := range m.Lines {
		rows = append(rows, text.Lines(line, inner)...)
	}
	return rows
}

func (m Model) Window(height int) int { return max(1, panel.ContentHeight(height)) }

func (m Model) View(th theme.Theme, status statusline.Model, width, height int) string {
	rows := m.Rows(width)
	window := m.Window(height)
	var content []string
	switch {
	case m.Err != nil:
		content = []string{th.Error.Render(text.Truncate("! "+m.Err.Error(), panel.ContentWidth(width)))}
	case len(rows) == 0:
		content = []string{th.Muted.Render("No log entries yet.")}
	default:
		top := m.logTop(len(rows), window)
		content = rows[top:min(len(rows), top+window)]
	}
	return panel.Render(th, "donsy.log"+status.TitleSuffix(), strings.Join(content, "\n"), width, height, false)
}

func (m Model) Footer(th theme.Theme, width int) string {
	return text.Truncate(th.Muted.Render("j/k scroll   r reload   esc back"), width)
}

func (m Model) Following() bool { return !m.detached }
func (m Model) Top() int        { return m.top }
