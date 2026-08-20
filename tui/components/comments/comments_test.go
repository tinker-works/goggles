package comments

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/tinker-works/donsy/netomatic"
	"github.com/tinker-works/goggles/tui/theme"
)

func TestRender_ShouldMarkAgentCommentsAndIndentBodies(t *testing.T) {
	rows := Render(theme.Default(), nil, []netomatic.Comment{{Author: "coding", Body: "pushed"}, {Author: "person", Body: "thanks"}}, 40)
	out := ansi.Strip(strings.Join(rows, "\n"))
	if !strings.Contains(out, "coding ⚙") || strings.Contains(out, "person ⚙") || !strings.Contains(out, "  pushed") {
		t.Fatalf("unexpected comment rendering: %q", out)
	}
}
