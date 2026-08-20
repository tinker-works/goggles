package runs

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/tinker-works/donsy/netomatic"
	"github.com/tinker-works/goggles/tui/components/statusline"
	"github.com/tinker-works/goggles/tui/components/transcript"
	"github.com/tinker-works/goggles/tui/theme"
	"github.com/tinker-works/goggles/tui/viewmodel"
)

func runner(status string) viewmodel.Runner {
	return viewmodel.Runner{Run: netomatic.AgentRun{ID: "run-1", Agent: "coder", Project: "checkout", Variant: "fast", Status: status}, Subject: "Checkout rewrite"}
}

func TestModel_AppendOutput_ShouldAccumulateAndReplaceToolCalls(t *testing.T) {
	m := New().TrackOutput("run-1").AppendOutput("run-1", []transcript.Entry{{Kind: transcript.Text, Text: "first"}}, 5, nil)
	m = m.AppendOutput("run-1", []transcript.Entry{{Kind: transcript.ToolUse, Tool: "grep", CallID: "call-1", Text: "running"}}, 10, nil)
	m = m.AppendOutput("run-1", []transcript.Entry{{Kind: transcript.ToolUse, Tool: "grep", CallID: "call-1", Text: "done"}}, 15, nil)
	if len(m.Output()) != 2 || m.Output()[1].Text != "done" || m.OutputOffset() != 15 {
		t.Fatalf("unexpected transcript accumulation: %+v offset=%d", m.Output(), m.OutputOffset())
	}
}

func TestModel_ScrollOutput_ShouldParkAndResumeFollowing(t *testing.T) {
	entries := make([]transcript.Entry, 40)
	for i := range entries {
		entries[i] = transcript.Entry{Kind: transcript.Text, Text: "line"}
	}
	m := New().TrackOutput("run-1").AppendOutput("run-1", entries, 1, nil)
	parked := m.ScrollOutput(-5, m.LogRows(100), m.LogWindow(30))
	if parked.FollowingOutput() || parked.OutputTop() <= 0 {
		t.Fatal("expected output to park above the tail")
	}
	if !parked.ScrollOutput(1000, m.LogRows(100), m.LogWindow(30)).FollowingOutput() {
		t.Fatal("expected returning to the tail to reattach")
	}
}

func TestModel_View_ShouldRenderPublicRunFieldsAndTerminalTranscript(t *testing.T) {
	m := New().SetRunners([]viewmodel.Runner{runner("succeeded")}).TrackOutput("run-1").AppendOutput("run-1", []transcript.Entry{{Kind: transcript.Text, Text: "finished"}}, 8, nil)
	view := m.View(theme.Default(), statusline.Model{}, 140, 20)
	if !strings.Contains(view, "coder") || !strings.Contains(view, "Checkout rewrite") {
		t.Fatalf("unexpected run list:\n%s", view)
	}
	detail := m.DetailView(theme.Default(), statusline.Model{}, runner("succeeded"), 100, 30)
	if !strings.Contains(detail, "finished") || strings.Contains(detail, "Waiting for the round") {
		t.Fatalf("unexpected run detail:\n%s", detail)
	}
	if got := lipgloss.Height(detail); got != 30 {
		t.Fatalf("expected fixed run detail height, got %d", got)
	}
}
