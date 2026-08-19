// Package layout contains terminal-size and string layout helpers.
package layout

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// Size is a terminal rectangle.
type Size struct {
	Width  int
	Height int
}

// Clamp keeps n between min and max.
func Clamp(n, min, max int) int {
	if max < min {
		return min
	}
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

// Columns divides width between count columns. The first columns receive the
// remainder, so the result always adds up to width.
func Columns(width, count int) []int {
	if count <= 0 {
		return nil
	}
	width = max(width, 0)
	result := make([]int, count)
	for i := range result {
		result[i] = width / count
		if i < width%count {
			result[i]++
		}
	}
	return result
}

// JoinColumns renders values into equal-width columns.
func JoinColumns(values []string, width, gap int) string {
	if len(values) == 0 || width <= 0 {
		return ""
	}
	if gap < 0 {
		gap = 0
	}
	columnWidth := (width - gap*(len(values)-1)) / len(values)
	if columnWidth <= 0 {
		return strings.Join(values, strings.Repeat(" ", gap))
	}
	parts := make([]string, len(values))
	for i, value := range values {
		parts[i] = lipgloss.NewStyle().Width(columnWidth).Render(value)
	}
	return strings.TrimRight(strings.Join(parts, strings.Repeat(" ", gap)), " ")
}

// PadRight pads value to width using spaces. Values wider than width are
// returned unchanged.
func PadRight(value string, width int) string {
	if width <= lipgloss.Width(value) {
		return value
	}
	return value + strings.Repeat(" ", width-lipgloss.Width(value))
}

// PadLeft pads value to width using spaces.
func PadLeft(value string, width int) string {
	if width <= lipgloss.Width(value) {
		return value
	}
	return strings.Repeat(" ", width-lipgloss.Width(value)) + value
}

// Center centers value in width columns.
func Center(value string, width int) string {
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(value)
}
