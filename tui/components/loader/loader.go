// Package loader renders a loading state without coupling it to an operation.
package loader

import (
	"charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tinker-works/goggles/tui/components/disco"
)

// Model is a loading indicator with an optional message.
type Model struct {
	Indicator disco.Model
	Message   string
	Style     lipgloss.Style
	Visible   bool
}

// New creates a visible loader.
func New(message string) Model {
	return Model{Indicator: disco.New(), Message: message, Visible: true}
}

// SetMessage changes the loading message.
func (m *Model) SetMessage(message string) { m.Message = message }

// Update advances the indicator when it receives a tick.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.Indicator, cmd = m.Indicator.Update(msg)
	return m, cmd
}

// View renders the indicator and message.
func (m Model) View() string {
	if !m.Visible {
		return ""
	}
	return m.Style.Render(m.Indicator.View() + " " + m.Message)
}
