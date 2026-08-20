// Package modal renders a centered, dismissible presentation panel.
package modal

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tinker-works/goggles/tui/theme"
)

// Field is one text input in a modal specification.
type Field struct {
	Prompt string
	Value  string
}

// Spec describes a form or confirmation dialog. Screens own the submitted
// values; this package only carries the presentation contract.
type Spec struct {
	ID            string
	Title         string
	Explain       string
	Message       string
	Fields        []Field
	Body          bool
	Submit        string
	Options       []string
	OptionsPrompt string
	Choices       []string
	Selected      int
	Cycle         []string
}

// SubmittedMsg is emitted by a form implementation after the user submits it.
type SubmittedMsg struct {
	ID      string
	Values  []string
	Body    string
	Options []string
	Choice  int
	Cycle   int
}

// OpenMsg asks the root model to replace the current modal with Spec.
type OpenMsg struct{ Spec Spec }

// Open returns a command that opens a modal specification.
func Open(spec Spec) tea.Cmd {
	return func() tea.Msg { return OpenMsg{Spec: spec} }
}

// Model is a modal panel. Parent models decide what a close event means.
type Model struct {
	Title      string
	Content    string
	Width      int
	Height     int
	Visible    bool
	Style      lipgloss.Style
	TitleStyle lipgloss.Style
}

// New creates a hidden modal.
func New(title, content string) Model {
	t := theme.Default()
	return Model{Title: title, Content: content, Visible: false, Style: t.Panel, TitleStyle: t.Title}
}

// Show opens the modal with content.
func (m *Model) Show(content string) { m.Content, m.Visible = content, true }

// Hide closes the modal.
func (m *Model) Hide() { m.Visible = false }

// SetSize sets the modal dimensions.
func (m *Model) SetSize(width, height int) { m.Width, m.Height = width, height }

// Update handles escape and enter as close requests by hiding the modal.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.Visible {
		return m, nil
	}
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok && (key.Matches(keyMsg, key.NewBinding(key.WithKeys("esc", "enter")))) {
		m.Visible = false
	}
	return m, nil
}

// View renders the modal or an empty string when hidden.
func (m Model) View() string {
	if !m.Visible {
		return ""
	}
	content := m.Content
	if m.Title != "" {
		content = m.TitleStyle.Render(m.Title) + "\n\n" + content
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
