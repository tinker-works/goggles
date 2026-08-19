// Package toast renders transient feedback without knowing its source.
package toast

import (
	"charm.land/lipgloss/v2"
	"github.com/tinker-works/goggles/tui/theme"
)

// Kind controls the visual treatment of a toast.
type Kind int

const (
	Info Kind = iota
	Success
	Warning
	Error
)

// Model stores one transient notification.
type Model struct {
	Message string
	Kind    Kind
	Visible bool
	Style   lipgloss.Style
}

// New creates a hidden toast.
func New() Model { return Model{Style: theme.Default.Panel} }

// Show displays message with kind.
func (m *Model) Show(message string, kind Kind) {
	m.Message, m.Kind, m.Visible = message, kind, true
	t := theme.Default
	switch kind {
	case Success:
		m.Style = t.Panel.BorderForeground(t.Palette.Success)
	case Warning:
		m.Style = t.Panel.BorderForeground(t.Palette.Warning)
	case Error:
		m.Style = t.Panel.BorderForeground(t.Palette.Error)
	default:
		m.Style = t.Panel
	}
}

// Hide hides the toast.
func (m *Model) Hide() { m.Visible = false }

// View renders the toast.
func (m Model) View() string {
	if !m.Visible {
		return ""
	}
	return m.Style.Render(m.Message)
}
