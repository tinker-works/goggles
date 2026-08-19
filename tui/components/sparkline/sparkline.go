// Package sparkline renders a compact sequence of values using block glyphs.
package sparkline

import (
	"math"
	"strings"

	"charm.land/lipgloss/v2"
)

// Model stores sparkline values and rendering options.
type Model struct {
	Values []float64
	Width  int
	Style  lipgloss.Style
}

// New creates an empty sparkline.
func New(width int) Model { return Model{Width: width} }

// SetValues replaces the data points.
func (m *Model) SetValues(values []float64) { m.Values = append(m.Values[:0], values...) }

// View renders values normalized to the range of the current series.
func (m Model) View() string {
	values := m.Values
	if m.Width > 0 && len(values) > m.Width {
		values = values[len(values)-m.Width:]
	}
	if len(values) == 0 {
		return ""
	}
	minValue, maxValue := values[0], values[0]
	for _, value := range values[1:] {
		minValue, maxValue = math.Min(minValue, value), math.Max(maxValue, value)
	}
	levels := []rune("▁▂▃▄▅▆▇█")
	result := make([]rune, len(values))
	for i, value := range values {
		ratio := 0.0
		if maxValue > minValue {
			ratio = (value - minValue) / (maxValue - minValue)
		}
		result[i] = levels[int(math.Round(ratio*float64(len(levels)-1)))]
	}
	return m.Style.Render(strings.TrimSpace(string(result)))
}
