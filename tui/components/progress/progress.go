// Package progress renders a compact determinate progress bar.
package progress

import (
	"fmt"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
)

const DefaultCells = 8

// Bar renders a compact count-based progress indicator.
func Bar(done, total, width int) string {
	if width <= 0 {
		return ""
	}
	if total <= 0 {
		return strings.Repeat("░", width)
	}
	done = max(0, min(done, total))
	filled := int(math.Round(float64(width) * float64(done) / float64(total)))
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

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
func (m *Model) SetPercent(value float64) { m.Value = clampPercent(value) }

// Percent returns the clamped progress value.
func (m Model) Percent() float64 { return clampPercent(m.Value) }

func clampPercent(value float64) float64 {
	if math.IsNaN(value) {
		return 0
	}
	return math.Max(0, math.Min(1, value))
}

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
