package epicdetail

import (
	"strings"
	"testing"
	"time"

	"charm.land/bubbletea/v2"
	"github.com/tinker-works/donsy/netomatic"
	"github.com/tinker-works/goggles/tui/components/statusline"
	"github.com/tinker-works/goggles/tui/components/transcript"
	"github.com/tinker-works/goggles/tui/theme"
	"github.com/tinker-works/goggles/tui/viewmodel"
	"github.com/tinker-works/goggles/tui/zones"
)

var now = time.Date(2026, 8, 20, 9, 0, 0, 0, time.UTC)

func checkout() netomatic.Epic {
	return netomatic.Epic{
		ID: "checkout", Title: "Checkout rewrite", State: "Ready", BranchPrefix: "checkout",
		Issues: []netomatic.Issue{
			{ID: "root", Title: "Checkout rewrite", Body: "the whole thing", State: "open", CreatedAt: now.Add(-time.Hour).Format(time.RFC3339)},
			{ID: "cart", ParentID: "root", Title: "Split cart", Repository: "api", State: "open",
				Comments: []netomatic.Comment{{ID: "c1", Author: "luuk", Body: "looks right"}}},
			{ID: "card", ParentID: "root", Title: "Card form", Repository: "web", State: "pr"},
		},
		PullRequests: []netomatic.PullRequest{{ID: "pr-7", IssueID: "card", Title: "Card form", Repository: "web", Number: 7, Status: "open"}},
	}
}

func loaded() Model { zones.Init(); return New().SetEpic(checkout()) }

func render(m Model, runner *viewmodel.Runner, entries []transcript.Entry) string {
	zones.Init()
	return zones.Scan(m.View(theme.Default(), statusline.Model{}, runner, entries, 180, 40, now))
}

func TestModel_SetEpic_ShouldFlattenAndRetainSelectionOnReload(t *testing.T) {
	m := loaded().MoveDown().ScrollDetail(3, 40, 5)
	m = m.SetEpic(checkout())
	if m.SelectedIssueID() != "cart" || m.DetailScroll() == 0 {
		t.Fatalf("expected selected issue and detail offset retained: id=%q offset=%d", m.SelectedIssueID(), m.DetailScroll())
	}
	if len(m.Rows()) != 3 || m.Rows()[2].Number != 7 {
		t.Fatalf("unexpected issue rows: %+v", m.Rows())
	}
}

func TestModel_Comment_ShouldFocusAndPreserveOrDiscardDraft(t *testing.T) {
	m, cmd := loaded().StartComment()
	if cmd == nil || !m.Commenting() {
		t.Fatal("expected comment editor focus")
	}
	m, _ = m.UpdateComment(tea.KeyPressMsg{Code: 'h', Text: "h"})
	if m.Comment.Value() != "h" {
		t.Fatalf("expected typed draft, got %q", m.Comment.Value())
	}
	if m.EndComment(true).Comment.Value() != "h" || m.EndComment(false).Comment.Value() != "" {
		t.Fatal("expected keep and discard to control the draft")
	}
}

func TestModel_View_ShouldRenderPublicEpicDataAndOutput(t *testing.T) {
	epic := checkout()
	epic.State = "Refine"
	epic.DraftingPasses = 2
	m := New().SetEpic(epic).ToggleOutput()
	runner := &viewmodel.Runner{Run: netomatic.AgentRun{ID: "run-1", Agent: "refiner", Status: "running"}, Elapsed: 74 * time.Second}
	view := render(m, runner, []transcript.Entry{{Kind: transcript.Text, Text: "writing the plan"}})
	for _, want := range []string{"Checkout rewrite", "ISSUES", "Split cart", "AGENT OUTPUT", "writing the plan"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected %q in epic detail view:\n%s", want, view)
		}
	}
}

func TestModel_View_ShouldRenderAtEveryReferenceWidth(t *testing.T) {
	for _, width := range []int{60, 90, 110, 200} {
		view := zones.Scan(loaded().View(theme.Default(), statusline.Model{}, nil, nil, width, 30, now))
		if !strings.Contains(view, "Checkout rewrite") {
			t.Fatalf("width %d did not render the epic:\n%s", width, view)
		}
	}
}
