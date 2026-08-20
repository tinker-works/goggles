// Package comments renders public comment DTOs as a readable discussion thread.
package comments

import (
	"fmt"
	"strings"

	"github.com/tinker-works/donsy/netomatic"
	"github.com/tinker-works/goggles/tui/components/markdown"
	"github.com/tinker-works/goggles/tui/theme"
)

const indent = 2

// Render returns the thread's display lines, headed by its count. renderer may
// be a markdown.Model or *markdown.Model; nil uses a fresh renderer per body.
func Render(th theme.Theme, renderer any, thread []netomatic.Comment, width int) []string {
	if len(thread) == 0 {
		return []string{th.Muted.Render("Comments (0)")}
	}

	rows := []string{th.Accent.Render(fmt.Sprintf("Comments (%d)", len(thread)))}
	for _, comment := range thread {
		author := comment.Author
		if Automated(author) {
			author += " ⚙"
		}
		rows = append(rows, "", th.Selected.Render(author))

		body := render(renderer, comment.Body, max(1, width-indent))
		if body == "" {
			rows = append(rows, th.Muted.Render(strings.Repeat(" ", indent)+"(empty)"))
			continue
		}
		for _, line := range strings.Split(body, "\n") {
			rows = append(rows, strings.Repeat(" ", indent)+line)
		}
	}
	return rows
}

func render(renderer any, source string, width int) string {
	var model markdown.Model
	switch value := renderer.(type) {
	case markdown.Model:
		model = value
	case *markdown.Model:
		if value != nil {
			model = *value
		}
	}
	model.SetSource(source)
	model.SetWidth(width)
	return model.Render()
}

// Automated reports whether an author is one of the daemon's agent roles.
func Automated(author string) bool {
	switch strings.ToLower(strings.TrimSpace(author)) {
	case "refiner", "issue-reviewer", "coding", "pr-reviewer", "merge":
		return true
	default:
		return false
	}
}
