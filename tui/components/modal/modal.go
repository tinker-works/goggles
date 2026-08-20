// Package modal renders a centered, dismissible presentation panel.
package modal

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tinker-works/goggles/tui/theme"
	"github.com/tinker-works/goggles/tui/zones"
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

// CancelledMsg is emitted when the user dismisses a cancellable form.
type CancelledMsg struct{ ID string }

// OpenMsg asks the root model to replace the current modal with Spec.
type OpenMsg struct{ Spec Spec }

// Open returns a command that opens a modal specification.
func Open(spec Spec) tea.Cmd {
	return func() tea.Msg { return OpenMsg{Spec: spec} }
}

func cancel(id string) tea.Cmd {
	return func() tea.Msg { return CancelledMsg{ID: id} }
}

// Model is a modal panel. Parent models decide what a close event means.
type Model struct {
	Title         string
	Content       string
	Width         int
	Height        int
	Visible       bool
	Style         lipgloss.Style
	TitleStyle    lipgloss.Style
	SelectedStyle lipgloss.Style

	spec    Spec
	form    bool
	focus   int
	choice  int
	option  int
	values  []string
	cursors []int
	checked []bool
	body    textarea.Model
	cycle   int
}

// New creates a hidden content modal.
func New(title, content string) Model {
	t := theme.Default()
	return Model{Title: title, Content: content, Visible: false, Style: t.Panel, TitleStyle: t.Title, SelectedStyle: t.Selected}
}

// NewForm creates a form model owned by a screen.
func NewForm(spec Spec, t theme.Theme, width int) Model {
	values := make([]string, len(spec.Fields))
	cursors := make([]int, len(spec.Fields))
	for i, field := range spec.Fields {
		values[i] = field.Value
		cursors[i] = len([]rune(field.Value))
	}
	checked := make([]bool, len(spec.Options))
	copy(checked, spec.Checked)
	choice := max(0, min(spec.Selected, max(0, len(spec.Choices)-1)))
	m := Model{Width: width, Style: t.Panel, TitleStyle: t.Title, SelectedStyle: t.Selected,
		spec: spec, form: true, choice: choice, values: values, cursors: cursors, checked: checked}
	if spec.Body {
		m.body = textarea.New()
		m.body.Prompt = "Body: "
		m.body.CharLimit = 5000
		m.body.ShowLineNumbers = false
		m.body.SetHeight(5)
		m.resizeBody()
		m.syncBodyFocus()
	}
	return m
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
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m.updateFormKey(msg)
	case tea.MouseClickMsg:
		return m.updateFormMouse(msg)
	case tea.PasteMsg:
		if !m.spec.Busy {
			m.insertFocused(msg.Content)
		}
		return m, nil
	default:
		return m, nil
	}
}

func (m Model) updateFormKey(keyMsg tea.KeyPressMsg) (Model, tea.Cmd) {
	if key.Matches(keyMsg, key.NewBinding(key.WithKeys("esc"))) {
		if !m.spec.Forced {
			return m, cancel(m.spec.ID)
		}
		return m, nil
	}
	if key.Matches(keyMsg, key.NewBinding(key.WithKeys("tab"))) {
		focusable := m.focusable()
		if focusable > 0 {
			m.focus = (m.focus + 1) % focusable
			m.syncBodyFocus()
		}
		return m, nil
	}
	if key.Matches(keyMsg, key.NewBinding(key.WithKeys("shift+tab"))) {
		focusable := m.focusable()
		if focusable > 0 {
			m.focus = (m.focus + focusable - 1) % focusable
			m.syncBodyFocus()
		}
		return m, nil
	}
	if key.Matches(keyMsg, key.NewBinding(key.WithKeys("enter"))) {
		return m, m.submit()
	}
	if len(m.spec.Choices) > 0 {
		if key.Matches(keyMsg, key.NewBinding(key.WithKeys("up", "k"))) {
			m.choice = max(0, m.choice-1)
			return m, nil
		}
		if key.Matches(keyMsg, key.NewBinding(key.WithKeys("down", "j"))) {
			m.choice = min(len(m.spec.Choices)-1, m.choice+1)
			return m, nil
		}
		return m, nil
	}
	if m.optionsFocused() {
		if key.Matches(keyMsg, key.NewBinding(key.WithKeys("up", "k"))) {
			m.option = max(0, m.option-1)
			return m, nil
		}
		if key.Matches(keyMsg, key.NewBinding(key.WithKeys("down", "j"))) {
			m.option = min(len(m.checked)-1, m.option+1)
			return m, nil
		}
		if key.Matches(keyMsg, key.NewBinding(key.WithKeys("space"))) {
			m.checked[m.option] = !m.checked[m.option]
		}
		return m, nil
	}
	if len(m.spec.Cycle) > 0 && key.Matches(keyMsg, key.NewBinding(key.WithKeys("t"))) {
		m.cycle = (m.cycle + 1) % len(m.spec.Cycle)
		return m, nil
	}
	if m.spec.Busy {
		return m, nil
	}
	if m.bodyFocused() {
		var cmd tea.Cmd
		m.body, cmd = m.body.Update(keyMsg)
		return m, cmd
	}
	if m.focus >= 0 && m.focus < len(m.values) {
		m.values[m.focus], m.cursors[m.focus] = editText(m.values[m.focus], m.cursors[m.focus], keyMsg)
	}
	return m, nil
}

func (m Model) updateFormMouse(msg tea.MouseClickMsg) (Model, tea.Cmd) {
	mouse := msg.Mouse()
	if mouse.Button != tea.MouseLeft || m.spec.Busy {
		return m, nil
	}
	if zones.In(mouse, zones.ModalSubmit) {
		return m, m.submit()
	}
	if zones.In(mouse, zones.ModalCancel) && !m.spec.Forced {
		return m, cancel(m.spec.ID)
	}
	for i := range m.spec.Choices {
		if zones.In(mouse, zones.ModalChoice(i)) {
			m.choice = i
			return m, nil
		}
	}
	for i := range m.spec.Options {
		if zones.In(mouse, zones.ModalOption(i)) {
			m.focus = m.focusable() - 1
			m.option = i
			m.syncBodyFocus()
			if i < len(m.checked) {
				m.checked[i] = !m.checked[i]
			}
			return m, nil
		}
	}
	if m.spec.Body && zones.In(mouse, zones.ModalBody) {
		m.focus = len(m.values)
		m.syncBodyFocus()
		return m, nil
	}
	for i := range m.values {
		if zones.In(mouse, zones.ModalField(i)) {
			m.focus = i
			m.syncBodyFocus()
			return m, nil
		}
	}
	return m, nil
}

func (m Model) submit() tea.Cmd {
	if m.spec.Busy {
		return nil
	}
	values := append([]string(nil), m.values...)
	options := make([]string, 0, len(m.checked))
	for i, checked := range m.checked {
		if checked && i < len(m.spec.Options) {
			options = append(options, m.spec.Options[i])
		}
	}
	choice := m.choice
	body := m.body.Value()
	cycle := m.cycle
	return func() tea.Msg {
		return SubmittedMsg{ID: m.spec.ID, Values: values, Body: body, Options: options, Choice: choice, Cycle: cycle}
	}
}

// Resize changes the form width.
func (m *Model) Resize(width int) {
	m.Width = width
	m.resizeBody()
}

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
	for i, choice := range m.spec.Choices {
		prefix, style := "  ", lipgloss.NewStyle()
		if i == m.choice {
			prefix, style = "› ", m.SelectedStyle
		}
		lines = append(lines, zones.Mark(zones.ModalChoice(i), style.Render(prefix+choice)))
	}
	for i, field := range m.spec.Fields {
		value := ""
		if i < len(m.values) {
			value = m.values[i]
		}
		lines = append(lines, zones.Mark(zones.ModalField(i), field.Prompt+": "+m.cursorText(value, m.cursors[i], m.focus == i)))
		if errorText := m.spec.FieldError(i); errorText != "" {
			lines = append(lines, errorText)
		}
	}
	if m.spec.Body {
		view := m.body.View()
		if m.bodyFocused() {
			view = m.SelectedStyle.Render(view)
		}
		lines = append(lines, zones.Mark(zones.ModalBody, view))
	}
	if len(m.spec.Cycle) > 0 {
		target := m.spec.Cycle[min(m.cycle, len(m.spec.Cycle)-1)]
		lines = append(lines, "Target: "+target+"  press t to change")
	}
	for i, option := range m.spec.Options {
		if i == 0 {
			prompt := m.spec.OptionsPrompt
			if prompt == "" {
				prompt = "Options"
			}
			lines = append(lines, prompt)
		}
		mark := " "
		if i < len(m.checked) && m.checked[i] {
			mark = "x"
		}
		prefix, style := "  ", lipgloss.NewStyle()
		if m.optionsFocused() && i == m.option {
			prefix, style = "› ", m.SelectedStyle
		}
		lines = append(lines, zones.Mark(zones.ModalOption(i), style.Render(prefix+"["+mark+"] "+option)))
	}
	if errorText := m.spec.FormError(); errorText != "" {
		lines = append(lines, errorText)
	}
	submit := m.submitLabel()
	if m.spec.Busy {
		submit += "…"
	}
	lines = append(lines, zones.Mark(zones.ModalSubmit, submit))
	if !m.spec.Forced {
		lines = append(lines, zones.Mark(zones.ModalCancel, "Cancel"))
	}
	width := max(20, min(72, m.Width-8))
	return m.Style.Width(width).Render(strings.Join(lines, "\n"))
}

func (m Model) submitLabel() string {
	if m.spec.Submit != "" {
		return m.spec.Submit
	}
	return "Submit"
}

func (m Model) optionsFocused() bool {
	return len(m.spec.Options) > 0 && m.focus == m.focusable()-1
}

func (m Model) bodyFocused() bool { return m.spec.Body && m.focus == len(m.values) }

func (m *Model) syncBodyFocus() {
	if !m.spec.Body {
		return
	}
	if m.bodyFocused() {
		m.body.Focus()
		return
	}
	m.body.Blur()
}

func (m *Model) resizeBody() {
	if !m.spec.Body {
		return
	}
	width := max(20, min(72, m.Width-8))
	m.body.SetWidth(max(10, width-4))
}

func (m Model) focusable() int {
	total := len(m.values)
	if m.spec.Body {
		total++
	}
	if len(m.spec.Options) > 0 {
		total++
	}
	return total
}

func (m *Model) insertFocused(value string) {
	if m.bodyFocused() {
		m.body, _ = m.body.Update(tea.PasteMsg{Content: value})
	} else if m.focus >= 0 && m.focus < len(m.values) {
		m.values[m.focus], m.cursors[m.focus] = insertText(m.values[m.focus], m.cursors[m.focus], value)
	}
}

func (m Model) cursorText(value string, position int, focused bool) string {
	if !focused || m.spec.Busy {
		return value
	}
	runes := []rune(value)
	position = max(0, min(position, len(runes)))
	return string(runes[:position]) + "▏" + string(runes[position:])
}

func insertText(value string, position int, inserted string) (string, int) {
	runes := []rune(value)
	position = max(0, min(position, len(runes)))
	insertedRunes := []rune(inserted)
	value = string(runes[:position]) + inserted + string(runes[position:])
	return value, position + len(insertedRunes)
}

func editText(value string, position int, keyMsg tea.KeyPressMsg) (string, int) {
	runes := []rune(value)
	position = max(0, min(position, len(runes)))
	if keyMsg.Text != "" {
		return insertText(value, position, keyMsg.Text)
	}
	switch keyMsg.Code {
	case tea.KeyLeft:
		return value, max(0, position-1)
	case tea.KeyRight:
		return value, min(len(runes), position+1)
	case tea.KeyHome:
		return value, 0
	case tea.KeyEnd:
		return value, len(runes)
	case tea.KeyBackspace:
		if position > 0 {
			return string(runes[:position-1]) + string(runes[position:]), position - 1
		}
	case tea.KeyDelete:
		if position < len(runes) {
			return string(runes[:position]) + string(runes[position+1:]), position
		}
	}
	return value, position
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
