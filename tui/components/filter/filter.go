// Package filter provides a text filter input and case-insensitive matcher.
package filter

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

// Model is a focused filter input.
type Model struct {
	Active bool
	Input  textinput.Model
}

// New creates a filter with the supplied placeholder.
func New(placeholder string) Model {
	input := textinput.New()
	input.Prompt = "/ "
	input.Placeholder = placeholder
	input.CharLimit = 256
	input.Focus()
	return Model{Active: true, Input: input}
}

// NewModel is an alias useful to callers that use Model constructors by name.
func NewModel() Model { return New("filter") }

// Focus gives the input keyboard focus.
func (m *Model) Focus() tea.Cmd {
	m.Active = true
	return m.Input.Focus()
}

// Blur removes keyboard focus.
func (m *Model) Blur() {
	m.Active = false
	m.Input.Blur()
}

// SetValue sets the current filter.
func (m *Model) SetValue(value string) { m.Input.SetValue(value) }

// Value returns the current filter.
func (m Model) Value() string { return m.Input.Value() }

// Matches reports whether value passes the current filter.
func (m Model) Matches(value string) bool {
	return Matches(m.Value(), value)
}

// Matches reports whether any candidate contains the case-insensitive filter.
func Matches(filter string, candidates ...string) bool {
	filter = strings.ToLower(strings.TrimSpace(filter))
	if filter == "" {
		return true
	}
	for _, candidate := range candidates {
		if strings.Contains(strings.ToLower(candidate), filter) {
			return true
		}
	}
	return false
}

// Update handles text input messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.Input, cmd = m.Input.Update(msg)
	return m, cmd
}

// View renders the input.
func (m Model) View() string { return m.Input.View() }
