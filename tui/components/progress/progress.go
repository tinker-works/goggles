// Package progress renders a compact determinate progress bar.
package progress

import (
	"fmt"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
)

// Model stores progress state.
type Model struct {
	Width          int
	Value          float64
	Full           rune
	Empty          rune
	Style          lipgloss.Style
	ShowPercentage bool
}

// New creates a progress bar.
func New(width int) Model {
	return Model{Width: width, Full: '█', Empty: '░', ShowPercentage: true}
}

// SetPercent sets and clamps the progress value.
func (m *Model) SetPercent(value float64) { m.Value = math.Max(0, math.Min(1, value)) }

// Percent returns the clamped progress value.
func (m Model) Percent() float64 { return math.Max(0, math.Min(1, m.Value)) }

// View renders the progress bar.
func (m Model) View() string {
	width := max(m.Width, 0)
	filled := int(math.Round(float64(width) * m.Percent()))
	bar := strings.Repeat(string(m.Full), filled) + strings.Repeat(string(m.Empty), width-filled)
	if !m.ShowPercentage {
		return m.Style.Render(bar)
	}
	return m.Style.Render(fmt.Sprintf("%s %3.0f%%", bar, m.Percent()*100))
}
