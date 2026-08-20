package tree

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/tinker-works/donsy/netomatic"
	"github.com/tinker-works/goggles/tui/theme"
)

func TestRender_ShouldShowPublicIssueMetadata(t *testing.T) {
	out := ansi.Strip(Render(theme.Default(), []Row{
		{Issue: netomatic.Issue{ID: "root", Title: "Root", State: "open"}, HasChild: true},
		{Issue: netomatic.Issue{ID: "child", Title: "Child", Repository: "api", State: "pr"}, Depth: 1, Number: 7},
	}, 0, 60))
	for _, want := range []string{"▾ Root", "    Child", "api", "#7", "pr"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in %q", want, out)
		}
	}
}
