// Package keys contains the keyboard bindings shared by the goggles views.
package keys

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// KeyMap is the common navigation key map. Applications can embed it and
// replace individual bindings without making components know about screens.
type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Home     key.Binding
	End      key.Binding
	Enter    key.Binding
	Back     key.Binding
	Escape   key.Binding
	Help     key.Binding
	Quit     key.Binding
}

// Default returns the bindings used by the common presentation components.
func Default() KeyMap {
	return KeyMap{
		Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Left:     key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "left")),
		Right:    key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "right")),
		PageUp:   key.NewBinding(key.WithKeys("pgup", "b", "ctrl+u"), key.WithHelp("pgup/b", "page up")),
		PageDown: key.NewBinding(key.WithKeys("pgdown", "f", "ctrl+d", "space"), key.WithHelp("pgdn/f", "page down")),
		Home:     key.NewBinding(key.WithKeys("home", "g"), key.WithHelp("g/home", "top")),
		End:      key.NewBinding(key.WithKeys("end", "G"), key.WithHelp("G/end", "bottom")),
		Enter:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		Back:     key.NewBinding(key.WithKeys("backspace"), key.WithHelp("backspace", "delete")),
		Escape:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		Help:     key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}

// DefaultKeyMap is retained as a convenient value for components that accept
// a key map as a field.
var DefaultKeyMap = Default()

// ShortHelp implements bubbles' help.KeyMap interface.
func (m KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{m.Up, m.Down, m.Enter, m.Escape}
}

// FullHelp implements bubbles' help.KeyMap interface.
func (m KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{m.Up, m.Down, m.Left, m.Right},
		{m.PageUp, m.PageDown, m.Home, m.End},
		{m.Enter, m.Back, m.Escape, m.Help, m.Quit},
	}
}

// Matches reports whether msg matches binding.
func Matches(msg tea.KeyPressMsg, binding key.Binding) bool {
	return key.Matches(msg, binding)
}
