// Package tabs provides keyboard-selectable tabs.
package tabs

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tinker-works/goggles/tui/theme"
)

// Model stores tab labels and the active tab.
type Model struct {
	Items         []string
	Active        int
	Width         int
	ActiveStyle   lipgloss.Style
	InactiveStyle lipgloss.Style
}

// New creates tabs.
func New(items ...string) Model {
	return Model{
		Items:       append([]string(nil), items...),
		ActiveStyle: lipgloss.NewStyle().Bold(true).Underline(true),
	}
}

// SetItems replaces tab labels.
func (m *Model) SetItems(items []string) {
	m.Items = append(m.Items[:0], items...)
	m.clamp()
}

// SetActive selects a tab by index.
func (m *Model) SetActive(index int) { m.Active = index; m.clamp() }

// ActiveIndex returns the selected tab index, or -1 when empty.
func (m Model) ActiveIndex() int {
	if len(m.Items) == 0 {
		return -1
	}
	return m.Active
}

// Update handles left/right and home/end navigation.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(keyMsg, key.NewBinding(key.WithKeys("left", "h"))):
			m.Active--
		case key.Matches(keyMsg, key.NewBinding(key.WithKeys("right", "l"))):
			m.Active++
		case key.Matches(keyMsg, key.NewBinding(key.WithKeys("home", "g"))):
			m.Active = 0
		case key.Matches(keyMsg, key.NewBinding(key.WithKeys("end", "G"))):
			m.Active = len(m.Items) - 1
		}
		m.clamp()
	}
	return m, nil
}

// View renders tabs separated by spaces.
func (m Model) View() string {
	parts := make([]string, len(m.Items))
	for i, item := range m.Items {
		if i == m.Active {
			parts[i] = m.ActiveStyle.Render(item)
		} else {
			parts[i] = m.InactiveStyle.Render(item)
		}
	}
	view := strings.Join(parts, "  ")
	if m.Width > 0 {
		view = lipgloss.NewStyle().Width(m.Width).Render(view)
	}
	return view
}

func (m *Model) clamp() {
	if len(m.Items) == 0 {
		m.Active = 0
		return
	}
	if m.Active < 0 {
		m.Active = 0
	}
	if m.Active >= len(m.Items) {
		m.Active = len(m.Items) - 1
	}
}

// Render is the compact helper used by screen headers.
func Render(th theme.Theme, items []string, active int) string {
	model := New(items...)
	model.ActiveStyle = th.Accent
	model.InactiveStyle = th.Muted
	model.SetActive(active)
	return model.View()
}
