// Package commentbox renders a bordered block of user-provided text.
package commentbox

import (
	"strings"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tinker-works/goggles/tui/components/text"
	"github.com/tinker-works/goggles/tui/theme"
	"github.com/tinker-works/goggles/tui/zones"
)

// Model supports both a static comment card and the screen's public comment
// editor. It deliberately accepts text rather than a domain comment type.
type Model struct {
	Author    string
	Timestamp string
	Body      string
	Width     int
	Style     lipgloss.Style
	Header    lipgloss.Style
	area      textarea.Model
	focused   bool
	ready     bool
}

// New creates a comment box using the default theme.
func New() Model {
	t := theme.Default()
	area := textarea.New()
	area.Prompt = "▏"
	area.CharLimit = 5000
	area.SetHeight(5)
	area.ShowLineNumbers = false
	return Model{
		Style:  t.Panel,
		Header: t.Subtitle.Bold(true),
		area:   area,
		ready:  true,
	}
}

func (m Model) initialized() Model {
	if m.ready {
		return m
	}
	fresh := New()
	m.area, m.ready = fresh.area, true
	return m
}

func (m Model) Focus() (Model, tea.Cmd) {
	m = m.initialized()
	m.focused = true
	return m, m.area.Focus()
}

func (m Model) Blur(keep bool) Model {
	m = m.initialized()
	m.focused = false
	m.area.Blur()
	if !keep {
		m.area.SetValue("")
	}
	return m
}

func (m Model) Focused() bool { return m.focused }
func (m Model) Value() string { return m.area.Value() }
func (m Model) Empty() bool   { return strings.TrimSpace(m.area.Value()) == "" }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	m = m.initialized()
	area, cmd := m.area.Update(msg)
	m.area = area
	return m, cmd
}

func (m *Model) Resize(width int) {
	value := m.initialized()
	value.area.SetWidth(max(10, width))
	*m = value
}

// SetContent sets the presentation fields.
func (m *Model) SetContent(author, timestamp, body string) {
	m.Author, m.Timestamp, m.Body = author, timestamp, body
}

// SetWidth sets the outer width. A non-positive value leaves width automatic.
func (m *Model) SetWidth(width int) { m.Width = width }

// View renders the comment.
func (m Model) View(args ...any) string {
	if len(args) > 0 {
		m = m.initialized()
		th, ok := args[0].(theme.Theme)
		if !ok {
			th = theme.Default()
		}
		width := 0
		if len(args) > 1 {
			width, _ = args[1].(int)
		}
		if !m.focused {
			return zones.Mark(zones.CommentBox,
				th.Accent.Render("c")+th.Muted.Render("  write a comment"))
		}
		rows := []string{
			th.Accent.Render("Comment") + th.Muted.Render("  (ctrl+s post, esc cancel)"),
			m.area.View(),
		}
		content := strings.Join(rows, "\n")
		if width > 0 {
			content = text.Fit(content, width)
		}
		return zones.Mark(zones.CommentBox, content)
	}
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
