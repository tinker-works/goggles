// Package keyhelp renders keyboard hints through bubbles' help component.
package keyhelp

import (
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/tinker-works/goggles/tui/components/text"
	"github.com/tinker-works/goggles/tui/theme"
)

// KeyMap is the minimal interface accepted by the help view.
type KeyMap interface {
	ShortHelp() []key.Binding
	FullHelp() [][]key.Binding
}

// Model wraps the standard help renderer with a compact public API.
type Model struct {
	Help help.Model
	Full bool
}

// New creates a help model.
func New() Model { return Model{Help: help.New()} }

// SetWidth limits the rendered help line.
func (m *Model) SetWidth(width int) { m.Help.SetWidth(width) }

// Update toggles the full help view on question mark.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok && key.Matches(keyMsg, key.NewBinding(key.WithKeys("?"))) {
		m.Full = !m.Full
		m.Help.ShowAll = m.Full
	}
	return m, nil
}

// View renders bindings.
func (m Model) View(bindings KeyMap) string { return m.Help.View(bindings) }

// Footer renders a compact short-help line for a screen key map.
func Footer(th theme.Theme, bindings KeyMap, width int) string {
	return FooterBindings(th, bindings.ShortHelp(), width)
}

// FooterBindings renders a compact short-help line from bindings.
func FooterBindings(th theme.Theme, bindings []key.Binding, width int) string {
	parts := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		if binding.Help().Desc == "" {
			continue
		}
		parts = append(parts, th.Accent.Render(binding.Help().Key)+" "+th.Muted.Render(binding.Help().Desc))
	}
	return text.Truncate(strings.Join(parts, "   "), width)
}
