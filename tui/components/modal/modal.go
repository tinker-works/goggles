// Package modal renders a centered, dismissible presentation panel.
package modal

import (
	"strings"

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
	Checked       []bool
	OptionsPrompt string
	Choices       []string
	Selected      int
	Cycle         []string
	Forced        bool
	Busy          bool
	Errors        map[int]string
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

	spec    Spec
	form    bool
	focus   int
	values  []string
	checked []bool
}

// New creates a hidden content modal.
func New(title, content string) Model {
	t := theme.Default()
	return Model{Title: title, Content: content, Visible: false, Style: t.Panel, TitleStyle: t.Title}
}

// NewForm creates a form model owned by a screen.
func NewForm(spec Spec, t theme.Theme, width int) Model {
	values := make([]string, len(spec.Fields))
	for i, field := range spec.Fields {
		values[i] = field.Value
	}
	return Model{Width: width, Style: t.Panel, TitleStyle: t.Title,
		spec: spec, form: true, values: values, checked: append([]bool(nil), spec.Checked...)}
}

// Show opens the modal with content.
func (m *Model) Show(content string) { m.Content, m.Visible = content, true }

// Hide closes the modal.
func (m *Model) Hide() { m.Visible = false }

// SetSize sets the modal dimensions.
func (m *Model) SetSize(width, height int) { m.Width, m.Height = width, height }

// Update handles escape and enter as close requests by hiding the modal.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if m.form {
		return m.updateForm(msg)
	}
	if !m.Visible {
		return m, nil
	}
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok && (key.Matches(keyMsg, key.NewBinding(key.WithKeys("esc", "enter")))) {
		m.Visible = false
	}
	return m, nil
}

func (m Model) updateForm(msg tea.Msg) (Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	if key.Matches(keyMsg, key.NewBinding(key.WithKeys("esc"))) {
		if !m.spec.Forced {
			return m, func() tea.Msg { return nil }
		}
		return m, nil
	}
	if key.Matches(keyMsg, key.NewBinding(key.WithKeys("tab"))) {
		focusable := len(m.values) + len(m.checked)
		if focusable > 0 {
			m.focus = (m.focus + 1) % focusable
		}
		return m, nil
	}
	if key.Matches(keyMsg, key.NewBinding(key.WithKeys("shift+tab"))) {
		focusable := len(m.values) + len(m.checked)
		if focusable > 0 {
			m.focus = (m.focus + focusable - 1) % focusable
		}
		return m, nil
	}
	if key.Matches(keyMsg, key.NewBinding(key.WithKeys("enter"))) {
		values := append([]string(nil), m.values...)
		options := make([]string, 0, len(m.checked))
		for i, checked := range m.checked {
			if checked && i < len(m.spec.Options) {
				options = append(options, m.spec.Options[i])
			}
		}
		return m, func() tea.Msg {
			return SubmittedMsg{ID: m.spec.ID, Values: values, Options: options}
		}
	}
	if key.Matches(keyMsg, key.NewBinding(key.WithKeys("space"))) && len(m.checked) > 0 {
		if m.focus >= len(m.values) {
			index := m.focus - len(m.values)
			if index >= 0 && index < len(m.checked) {
				m.checked[index] = !m.checked[index]
			}
		}
		return m, nil
	}
	if m.focus >= 0 && m.focus < len(m.values) {
		if keyMsg.Text == "" {
			if keyMsg.Code == tea.KeyBackspace && len(m.values[m.focus]) > 0 {
				m.values[m.focus] = m.values[m.focus][:len(m.values[m.focus])-1]
			}
			return m, nil
		}
		m.values[m.focus] += keyMsg.Text
	}
	return m, nil
}

// Resize changes the form width.
func (m *Model) Resize(width int) { m.Width = width }

// Values returns the current form fields.
func (m Model) Values() []string { return append([]string(nil), m.values...) }

// Checked returns the current checklist state.
func (m Model) Checked() []bool { return append([]bool(nil), m.checked...) }

// View renders the modal or an empty string when hidden.
func (m Model) View() string {
	if m.form {
		return m.formView()
	}
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

func (m Model) formView() string {
	lines := []string{}
	if m.spec.Title != "" {
		lines = append(lines, m.TitleStyle.Render(m.spec.Title))
	}
	if m.spec.Explain != "" {
		lines = append(lines, m.spec.Explain)
	}
	if m.spec.Message != "" {
		lines = append(lines, m.spec.Message)
	}
	for i, field := range m.spec.Fields {
		value := ""
		if i < len(m.values) {
			value = m.values[i]
		}
		lines = append(lines, field.Prompt+": "+value)
		if errorText := m.spec.FieldError(i); errorText != "" {
			lines = append(lines, errorText)
		}
	}
	for i, option := range m.spec.Options {
		mark := " "
		if i < len(m.checked) && m.checked[i] {
			mark = "x"
		}
		lines = append(lines, "["+mark+"] "+option)
	}
	if errorText := m.spec.FormError(); errorText != "" {
		lines = append(lines, errorText)
	}
	if m.spec.Submit != "" {
		submit := m.spec.Submit
		if m.spec.Busy {
			submit += "…"
		}
		lines = append(lines, submit)
	}
	return strings.Join(lines, "\n")
}

// WithError returns a copy with an error attached to a field or the form.
func (s Spec) WithError(index int, value string) Spec {
	copy := s
	copy.Errors = make(map[int]string, len(s.Errors)+1)
	for key, errorText := range s.Errors {
		copy.Errors[key] = errorText
	}
	copy.Errors[index] = value
	copy.Busy = false
	return copy
}

// FieldError returns an error attached to index.
func (s Spec) FieldError(index int) string {
	if s.Errors == nil {
		return ""
	}
	return s.Errors[index]
}

// FormError returns an error attached after the fields.
func (s Spec) FormError() string { return s.FieldError(len(s.Fields)) }
