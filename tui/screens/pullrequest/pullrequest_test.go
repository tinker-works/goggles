package pullrequest

import (
	"errors"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/tinker-works/donsy/netomatic"
	"github.com/tinker-works/goggles/tui/components/statusline"
	"github.com/tinker-works/goggles/tui/theme"
	"github.com/tinker-works/goggles/tui/zones"
)

func open() netomatic.PullRequest {
	return netomatic.PullRequest{ID: "pr-7", IssueID: "card", Title: "Card form", Repository: "web", Number: 7,
		Status: "open", Head: "feature/card", Base: "main", Comments: []netomatic.Comment{{ID: "c1", Author: "reviewer", Body: "rename the field"}}}
}

func issue() *netomatic.Issue {
	return &netomatic.Issue{ID: "card", ParentID: "root", Title: "Card form", Repository: "web", State: "pr", Body: "add the card form"}
}

func files() []DiffFile {
	return []DiffFile{{Path: "web/src/card.tsx", Hunks: []DiffHunk{{Kind: '@', Text: "@@ -1 +1 @@"}, {Kind: '+', Text: "const card = 1"}, {Kind: '-', Text: "const old = 1"}, {Kind: ' ', Text: "context"}}}}
}

func loaded() Model { return New().Load("checkout", open(), issue()) }

func TestModel_Load_ShouldResetDifferentPRAndKeepSamePRReaderState(t *testing.T) {
	m := loaded().SetDiff("pr-7", files(), nil).SwitchTab().MoveDown().SetFilter("card")
	m = m.Load("checkout", open(), issue())
	if m.Tab != Diff || m.FileIndex != 0 || m.Filter.Value != "card" {
		t.Fatalf("expected same PR state retained: tab=%d file=%d filter=%q", m.Tab, m.FileIndex, m.Filter.Value)
	}
	other := open()
	other.ID = "pr-8"
	m = m.Load("checkout", other, issue())
	if m.Tab != Conversation || m.Filter.Value != "" || m.DiffLayout != Inline {
		t.Fatalf("expected different PR state reset: %+v", m)
	}
}

func TestModel_SetDiff_ShouldDropStaleResultsAndFilterFiles(t *testing.T) {
	m := loaded()
	stale := m.SetDiff("other", files(), nil)
	ok, _ := stale.DiffState()
	if ok {
		t.Fatal("stale diff result was accepted")
	}
	m = m.SetDiff("pr-7", append(files(), DiffFile{Path: "api/cart.go"}), nil).SetFilter("cart")
	if len(m.Files()) != 1 || m.Files()[0].Path != "api/cart.go" {
		t.Fatalf("unexpected filtered files: %+v", m.Files())
	}
	got := m.SetDiff("pr-7", nil, errors.New("daemon unavailable"))
	if got.diffErr == nil {
		t.Fatal("expected diff error to be retained")
	}
}

func TestParseDiff_ShouldUsePublicDiffPresentationTypes(t *testing.T) {
	diff := "diff --git a/old.txt b/new.txt\n--- a/old.txt\n+++ b/new.txt\n@@ -1 +1 @@\n-old\n+new\n"
	parsed := ParseDiff(diff)
	if len(parsed) != 1 || parsed[0].Path != "new.txt" || len(parsed[0].Hunks) != 3 {
		t.Fatalf("unexpected parsed diff: %+v", parsed)
	}
}

func TestModel_View_ShouldRenderConversationDiffAndFailureStates(t *testing.T) {
	zones.Init()
	conversation := zones.Scan(loaded().View(theme.Default(), statusline.Model{}, 140, 30))
	if !strings.Contains(conversation, "Card form · web#7") || !strings.Contains(ansi.Strip(conversation), "rename the") {
		t.Fatalf("unexpected conversation view:\n%s", conversation)
	}
	diff := zones.Scan(loaded().SwitchTab().SetDiff("pr-7", files(), nil).View(theme.Default(), statusline.Model{}, 140, 30))
	if !strings.Contains(ansi.Strip(diff), "const card = 1") {
		t.Fatalf("unexpected diff view:\n%s", diff)
	}
	failed := zones.Scan(loaded().SwitchTab().SetDiff("pr-7", nil, errors.New("daemon unavailable")).View(theme.Default(), statusline.Model{}, 140, 30))
	if !strings.Contains(failed, "daemon unavailable") {
		t.Fatalf("expected diff failure in view:\n%s", failed)
	}
}

func TestModel_View_ShouldKeepDiffLinesWithinReferenceWidths(t *testing.T) {
	pr := open()
	pr.Base = strings.Repeat("base/", 8)
	pr.Head = strings.Repeat("head/", 8)
	long := DiffFile{Path: strings.Repeat("very-long-directory/", 5) + "file.go", Hunks: []DiffHunk{
		{Kind: '@', Text: "@@ -1 +1 @@"},
		{Kind: '-', Text: strings.Repeat("old", 30)},
		{Kind: '+', Text: strings.Repeat("new", 30)},
	}}
	m := New().Load("checkout", pr, issue()).SetDiff("pr-7", []DiffFile{long}, nil).SwitchTab().ToggleDiffLayout()
	for _, width := range []int{62, 100, 140} {
		view := zones.Scan(m.View(theme.Default(), statusline.Model{}, width, 30))
		for _, line := range strings.Split(view, "\n") {
			if got := lipgloss.Width(line); got > width {
				t.Fatalf("width %d rendered a %d-cell line", width, got)
			}
		}
	}
}
