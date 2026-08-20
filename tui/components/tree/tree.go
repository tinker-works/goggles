// Package tree renders an epic's issue tree with depth and pull-request metadata.
package tree

import (
	"fmt"
	"strings"

	"github.com/tinker-works/donsy/netomatic"
	"github.com/tinker-works/goggles/tui/components/text"
	"github.com/tinker-works/goggles/tui/theme"
	"github.com/tinker-works/goggles/tui/zones"
)

// Row is one issue placed in the tree.
type Row struct {
	Issue    netomatic.Issue
	Depth    int
	HasChild bool
	Number   int
}

// Render draws all issue entries with selected highlighted.
func Render(th theme.Theme, rows []Row, selected, width int) string {
	return RenderWindow(th, rows, selected, 0, len(rows), width)
}

// RenderWindow renders a window of complete issue entries while retaining their
// original indexes in click zones.
func RenderWindow(th theme.Theme, rows []Row, selected, start, count, width int) string {
	if len(rows) == 0 {
		return th.Muted.Render("No issues yet — this epic hasn't been drafted.")
	}
	start = max(0, min(start, len(rows)))
	end := min(len(rows), start+max(0, count))
	lines := make([]string, 0, (end-start)*2)
	for i := start; i < end; i++ {
		row := rows[i]
		indent := strings.Repeat("  ", row.Depth)
		marker := " "
		if row.HasChild {
			marker = "▾"
		}
		title := text.Truncate(row.Issue.Title, max(1, width-len(indent)-2))
		titleLine := indent + marker + " " + title
		if i == selected {
			titleLine = th.Selected.Render(titleLine)
		}
		lines = append(lines,
			zones.Mark(zones.EpicTreeRow(i), titleLine),
			th.Muted.Render(indent+"   "+meta(th, row)),
		)
	}
	return strings.Join(lines, "\n")
}

func meta(th theme.Theme, row Row) string {
	parts := []string{}
	if row.Issue.Repository != "" {
		parts = append(parts, row.Issue.Repository)
	}
	if row.Number > 0 {
		parts = append(parts, fmt.Sprintf("#%d", row.Number))
	}
	parts = append(parts, th.Badge(th.IssueStateStyle(row.Issue.State), row.Issue.State))
	return strings.Join(parts, "  ")
}
