// Package disco provides a small, deterministic animated indicator.
package disco

import tea "charm.land/bubbletea/v2"

// TickMsg advances an indicator by one frame.
type TickMsg struct{}

// Model is a frame-based animation model.
type Model struct {
	Frames []string
	Frame  int
	Active bool
}

// New creates a disco indicator.
func New(frames ...string) Model {
	if len(frames) == 0 {
		frames = []string{"·", "✦", "*", "✧"}
	}
	return Model{Frames: append([]string(nil), frames...), Active: true}
}

// Tick returns a command that advances the indicator when sent by a parent.
func Tick() tea.Cmd { return func() tea.Msg { return TickMsg{} } }

// Update handles animation ticks.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if _, ok := msg.(TickMsg); ok && m.Active && len(m.Frames) > 0 {
		m.Frame = (m.Frame + 1) % len(m.Frames)
	}
	return m, nil
}

// View returns the current frame.
func (m Model) View() string {
	if len(m.Frames) == 0 {
		return ""
	}
	return m.Frames[m.Frame%len(m.Frames)]
}
