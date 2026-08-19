// Package selection provides a small, domain-free cursor model.
package selection

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// Item is a selectable presentation value. Value is intentionally an
// interface so the primitive does not prescribe a domain model.
type Item struct {
	Value any
	Label string
}

// Model tracks a cursor over a list of items.
type Model struct {
	Items  []Item
	Cursor int
	Active bool
	KeyMap KeyMap
}

// KeyMap contains the bindings used by Model.
type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Home     key.Binding
	End      key.Binding
	PageUp   key.Binding
	PageDown key.Binding
}

// DefaultKeyMap returns standard list navigation bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up:       key.NewBinding(key.WithKeys("up", "k")),
		Down:     key.NewBinding(key.WithKeys("down", "j")),
		Home:     key.NewBinding(key.WithKeys("home", "g")),
		End:      key.NewBinding(key.WithKeys("end", "G")),
		PageUp:   key.NewBinding(key.WithKeys("pgup", "b")),
		PageDown: key.NewBinding(key.WithKeys("pgdown", "f", "space")),
	}
}

// New creates a selection model.
func New(items ...Item) Model {
	return Model{Items: append([]Item(nil), items...), Active: true, KeyMap: DefaultKeyMap()}
}

// NewString creates a selection model from labels.
func NewString(items ...string) Model {
	values := make([]Item, len(items))
	for i, item := range items {
		values[i] = Item{Value: item, Label: item}
	}
	return New(values...)
}

// SetItems replaces the items and keeps the cursor within the new bounds.
func (m *Model) SetItems(items []Item) {
	m.Items = append(m.Items[:0], items...)
	m.clamp()
}

// SetStrings replaces the items with string values.
func (m *Model) SetStrings(items []string) {
	values := make([]Item, len(items))
	for i, item := range items {
		values[i] = Item{Value: item, Label: item}
	}
	m.SetItems(values)
}

// Selected returns the selected item, or false when the selection is empty.
func (m Model) Selected() (Item, bool) {
	if m.Cursor < 0 || m.Cursor >= len(m.Items) {
		return Item{}, false
	}
	return m.Items[m.Cursor], true
}

// Index returns the selected index, or -1 for an empty selection.
func (m Model) Index() int {
	if len(m.Items) == 0 {
		return -1
	}
	return m.Cursor
}

// Move moves the cursor by delta positions.
func (m *Model) Move(delta int) {
	m.Cursor += delta
	m.clamp()
}

// MoveUp moves the cursor up by one position.
func (m *Model) MoveUp() { m.Move(-1) }

// MoveDown moves the cursor down by one position.
func (m *Model) MoveDown() { m.Move(1) }

// First selects the first item.
func (m *Model) First() { m.Cursor = 0; m.clamp() }

// Last selects the final item.
func (m *Model) Last() { m.Cursor = len(m.Items) - 1; m.clamp() }

// Update handles navigation messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.Active {
		return m, nil
	}
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(keyMsg, m.KeyMap.Up):
			m.MoveUp()
		case key.Matches(keyMsg, m.KeyMap.Down):
			m.MoveDown()
		case key.Matches(keyMsg, m.KeyMap.Home):
			m.First()
		case key.Matches(keyMsg, m.KeyMap.End):
			m.Last()
		case key.Matches(keyMsg, m.KeyMap.PageUp):
			m.Move(-5)
		case key.Matches(keyMsg, m.KeyMap.PageDown):
			m.Move(5)
		}
	}
	return m, nil
}

func (m *Model) clamp() {
	if len(m.Items) == 0 {
		m.Cursor = 0
		return
	}
	if m.Cursor < 0 {
		m.Cursor = 0
	}
	if m.Cursor >= len(m.Items) {
		m.Cursor = len(m.Items) - 1
	}
}
