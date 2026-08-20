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

func renderPull(m Model, status statusline.Model, width, height int) string {
	zones.Init()
	return zones.Scan(m.View(theme.Default(), status, width, height))
}

func withViewDiff() Model {
	return loaded().SetDiff("pr-7", []DiffFile{
		{Path: "web/src/card.tsx", Hunks: []DiffHunk{
			{Kind: '@', Text: "@@ -1,3 +1,5 @@"},
			{Kind: '+', Text: "const card = 1"},
			{Kind: '-', Text: "const old = 1"},
			{Kind: ' ', Text: "context"},
		}},
		{Path: "web/src/checkout.tsx"},
		{Path: "api/cart.go"},
	}, nil)
}

func TestModel_View_ShouldReportTheEmptyAndOfflineStates(t *testing.T) {
	empty := renderPull(New(), statusline.Model{}, 120, 30)
	offline := renderPull(loaded(), statusline.Model{}.Fail(), 140, 30)
	if !strings.Contains(empty, "No pull request selected") {
		t.Fatalf("expected the empty state, got:\n%s", empty)
	}
	if !strings.Contains(offline, "offline") {
		t.Fatalf("expected the offline suffix, got:\n%s", offline)
	}
}

func TestModel_View_ShouldTitleTheIssueAndFallbackToThePullRequest(t *testing.T) {
	view := ansi.Strip(renderPull(loaded(), statusline.Model{}, 140, 30))
	withoutIssue := renderPull(New().Load("checkout", open(), nil), statusline.Model{}, 140, 30)
	if !strings.Contains(view, "Card form · web#7") {
		t.Fatalf("expected the issue title and repository number, got:\n%s", view)
	}
	if !strings.Contains(withoutIssue, "Card form") {
		t.Fatalf("expected the pull request title fallback, got:\n%s", withoutIssue)
	}
	pr := open()
	pr.Number = 0
	withoutNumber := renderPull(New().Load("checkout", pr, issue()), statusline.Model{}, 140, 30)
	if strings.Contains(withoutNumber, "web#") {
		t.Fatalf("expected no number in the title, got:\n%s", withoutNumber)
	}
}

func TestModel_View_ShouldShowBranchesAndEveryStatus(t *testing.T) {
	view := renderPull(loaded(), statusline.Model{}, 140, 30)
	for _, want := range []string{"feature/card", "main", "open"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected the header to contain %q, got:\n%s", want, view)
		}
	}
	missing := open()
	missing.Head, missing.Base = "", ""
	missingView := renderPull(New().Load("checkout", missing, issue()), statusline.Model{}, 140, 30)
	if !strings.Contains(missingView, "(head)") || !strings.Contains(missingView, "(base)") {
		t.Fatalf("expected branch placeholders, got:\n%s", missingView)
	}
	for _, status := range []string{"open", "merged", "closed"} {
		pr := open()
		pr.Status = status
		view := renderPull(New().Load("checkout", pr, issue()), statusline.Model{}, 140, 30)
		if !strings.Contains(view, status) {
			t.Fatalf("expected the %q status badge, got:\n%s", status, view)
		}
	}
}

func TestModel_View_ShouldRenderConversationContentAndEmptyReview(t *testing.T) {
	view := ansi.Strip(renderPull(loaded(), statusline.Model{}, 140, 30))
	for _, want := range []string{"Conversation", "add the card form", "Comments (1)", "rename the field"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected the conversation to contain %q, got:\n%s", want, view)
		}
	}
	pr := open()
	pr.Comments = nil
	blank := *issue()
	blank.Body = ""
	empty := ansi.Strip(renderPull(New().Load("checkout", pr, &blank), statusline.Model{}, 140, 30))
	if !strings.Contains(empty, "No description") || !strings.Contains(empty, "No review comments yet") {
		t.Fatalf("expected empty conversation markers, got:\n%s", empty)
	}
}

func TestModel_View_ShouldRenderActionsForOpenStaleAndTerminalRequests(t *testing.T) {
	openView := renderPull(loaded(), statusline.Model{}, 140, 30)
	if !strings.Contains(openView, "[ Merge ]") || !strings.Contains(openView, "requires click") {
		t.Fatalf("expected the merge action, got:\n%s", openView)
	}
	stale := open()
	stale.Reviews, stale.ReviewedHead, stale.ReviewedBase = 1, "old-head", "main"
	staleView := renderPull(New().Load("checkout", stale, issue()), statusline.Model{}, 140, 30)
	if !strings.Contains(staleView, "Review is stale") {
		t.Fatalf("expected the stale banner, got:\n%s", staleView)
	}
	terminal := open()
	terminal.Status = "merged"
	terminalView := renderPull(New().Load("checkout", terminal, issue()), statusline.Model{}, 140, 30)
	if strings.Contains(terminalView, "[ Merge ]") || !strings.Contains(terminalView, "open on GitHub") {
		t.Fatalf("expected terminal actions to collapse, got:\n%s", terminalView)
	}
	current := issue()
	current.State = "pr"
	closed := open()
	closed.Status = "closed"
	mergeView := renderPull(New().Load("checkout", closed, current), statusline.Model{}, 140, 30)
	if !strings.Contains(mergeView, "current pull request") || !strings.Contains(mergeView, "Merge") {
		t.Fatalf("expected the current pull request merge action, got:\n%s", mergeView)
	}
}

func TestModel_View_ShouldRegisterTheMergeButtonZone(t *testing.T) {
	view := renderPull(loaded(), statusline.Model{}, 140, 30)
	if _, _, ok := zones.Bounds(zones.MergeButton); !ok {
		t.Fatalf("expected the merge button to register a click zone, got:\n%s", view)
	}
}

func TestModel_View_ShouldReportDiffLoadingFailureEmptyAndFilterStates(t *testing.T) {
	computing := renderPull(loaded().SwitchTab(), statusline.Model{}, 140, 30)
	failed := renderPull(loaded().SwitchTab().SetDiff("pr-7", nil, errors.New("clone missing")), statusline.Model{}, 140, 30)
	empty := renderPull(loaded().SwitchTab().SetDiff("pr-7", nil, nil), statusline.Model{}, 140, 30)
	filtered := renderPull(withViewDiff().SwitchTab().SetFilter("nonsense"), statusline.Model{}, 140, 30)

	if !strings.Contains(computing, "Computing the diff") {
		t.Fatalf("expected the computing message, got:\n%s", computing)
	}
	if !strings.Contains(failed, "No diff available") || !strings.Contains(failed, "clone missing") {
		t.Fatalf("expected the diff failure, got:\n%s", failed)
	}
	if !strings.Contains(empty, "No changes between base and head") {
		t.Fatalf("expected the empty diff message, got:\n%s", empty)
	}
	if !strings.Contains(filtered, "No files match nonsense") {
		t.Fatalf("expected the no-match message, got:\n%s", filtered)
	}
}

func TestModel_View_ShouldDrawInlineAndTwoColumnDiffs(t *testing.T) {
	inline := ansi.Strip(renderPull(withViewDiff().SwitchTab(), statusline.Model{}, 140, 30))
	twoColumn := ansi.Strip(renderPull(withViewDiff().SwitchTab().ToggleDiffLayout(), statusline.Model{}, 140, 30))
	for _, want := range []string{"Inline", "v toggle", "@@ web/src/card.tsx", "+ const card = 1", "- const old = 1"} {
		if !strings.Contains(inline, want) {
			t.Fatalf("expected inline diff to contain %q, got:\n%s", want, inline)
		}
	}
	for _, want := range []string{"Two-column", "BASE main", "HEAD feature/card", "context"} {
		if !strings.Contains(twoColumn, want) {
			t.Fatalf("expected two-column diff to contain %q, got:\n%s", want, twoColumn)
		}
	}
	if strings.Count(twoColumn, "@@ -1,3 +1,5 @@") != 1 {
		t.Fatalf("expected one full-width hunk header, got:\n%s", twoColumn)
	}
}

func TestModel_View_ShouldKeepDiffContentWithinReferenceWidths(t *testing.T) {
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
		view := renderPull(m, statusline.Model{}, width, 30)
		for _, line := range strings.Split(view, "\n") {
			if got := lipgloss.Width(line); got > width {
				t.Fatalf("width %d rendered a %d-cell line", width, got)
			}
		}
	}
}

func TestModel_View_ShouldShortenPathsAndStackDiffPanes(t *testing.T) {
	long := []DiffFile{{Path: "web/src/features/checkout/components/CardForm.tsx"}}
	view := renderPull(loaded().SwitchTab().SetDiff("pr-7", long, nil), statusline.Model{}, 140, 30)
	if !strings.Contains(view, ".../CardForm.tsx") && !strings.Contains(view, "CardForm.tsx") {
		t.Fatalf("expected the file name to remain visible, got:\n%s", view)
	}
	narrow := renderPull(withViewDiff().SwitchTab(), statusline.Model{}, 70, 40)
	if strings.Index(narrow, "const card = 1") < strings.LastIndex(narrow, "card.tsx") {
		t.Fatalf("expected the hunks below the file tree, got:\n%s", narrow)
	}
	offTree := withViewDiff().SwitchTab()
	offTree.FileIndex = 99
	if !strings.Contains(renderPull(offTree, statusline.Model{}, 140, 30), "Select a file") {
		t.Fatal("expected a prompt when the file cursor is off the tree")
	}
}

func TestModel_View_ShouldShowFilterPromptAndPinTheCommentEditor(t *testing.T) {
	filtering := renderPull(loaded().StartFilter().SetFilter("web"), statusline.Model{}, 140, 30)
	if !strings.Contains(filtering, "filter: web") || !strings.Contains(filtering, "esc clear") {
		t.Fatalf("expected the filter prompt, got:\n%s", filtering)
	}
	if strings.Contains(filtering, "[ Merge ]") {
		t.Fatal("expected the merge button hidden while filtering")
	}
	pr := open()
	long := *issue()
	long.Body = strings.Repeat("a scrollable paragraph.\n\n", 40)
	m := New().Load("checkout", pr, &long)
	m, _ = m.StartComment()
	m.Resize(120)
	top := renderPull(m, statusline.Model{}, 140, 30)
	scrolled := renderPull(m.ScrollConversation(10, m.ConversationRows(theme.Default(), 140),
		m.ConversationWindow(theme.Default(), 140, 30)), statusline.Model{}, 140, 30)
	if top == scrolled {
		t.Fatal("expected the conversation scroll to change the frame")
	}
	for _, frame := range []string{top, scrolled} {
		if !strings.Contains(frame, "ctrl+s post") {
			t.Fatalf("expected the editor pinned in every frame, got:\n%s", frame)
		}
	}
}

func TestModel_View_ShouldMarkConversationScrollAndHideTheEditorOnDiff(t *testing.T) {
	long := *issue()
	long.Body = strings.Repeat("a scrollable paragraph.\n\n", 40)
	view := renderPull(New().Load("checkout", open(), &long), statusline.Model{}, 140, 20)
	if !strings.Contains(view, "↓") {
		t.Fatalf("expected a conversation scroll marker, got:\n%s", view)
	}
	m := withViewDiff().SwitchTab()
	m, _ = m.StartComment()
	if strings.Contains(renderPull(m, statusline.Model{}, 140, 30), "write a comment") {
		t.Fatal("expected the comment editor to be hidden on the diff tab")
	}
}

func TestModel_ConversationWindow_ShouldShrinkForBannersAndEditor(t *testing.T) {
	stale := open()
	stale.Reviews, stale.ReviewedHead = 1, "old-head"
	plain := loaded()
	banner := New().Load("checkout", stale, issue())
	editing, _ := plain.StartComment()
	plainWindow := plain.ConversationWindow(theme.Default(), 140, 30)
	if banner.ConversationWindow(theme.Default(), 140, 30) >= plainWindow {
		t.Fatal("expected the stale banner to shrink the conversation window")
	}
	if editing.ConversationWindow(theme.Default(), 140, 30) >= plainWindow {
		t.Fatal("expected the editor to shrink the conversation window")
	}
	if loaded().ConversationWindow(theme.Default(), 40, 4) < 1 {
		t.Fatal("expected a positive conversation window on a tiny terminal")
	}
}

func TestModel_ConversationRowsAndDiffRows_ShouldGrowWithContent(t *testing.T) {
	quiet := open()
	quiet.Comments = nil
	loud := open()
	for i := range 10 {
		loud.Comments = append(loud.Comments, netomatic.Comment{ID: string(rune('a' + i)), Author: "reviewer", Body: "another remark"})
	}
	if New().Load("checkout", loud, issue()).ConversationRows(theme.Default(), 140) <=
		New().Load("checkout", quiet, issue()).ConversationRows(theme.Default(), 140) {
		t.Fatal("expected a longer review thread to render more rows")
	}
	m := withViewDiff().SwitchTab()
	if m.DiffRows(theme.Default(), 140) != 6 || m.ToggleDiffLayout().DiffRows(theme.Default(), 140) != 5 {
		t.Fatalf("expected inline and two-column row totals, got %d and %d", m.DiffRows(theme.Default(), 140), m.ToggleDiffLayout().DiffRows(theme.Default(), 140))
	}
}

func TestModel_Footer_ShouldAdvertiseTheActiveScreenKeys(t *testing.T) {
	conversation := loaded().Footer(theme.Default(), 200)
	diff := loaded().SwitchTab().Footer(theme.Default(), 200)
	for _, want := range []string{"comment", "retry", "back"} {
		if !strings.Contains(conversation, want) {
			t.Fatalf("expected conversation footer to contain %q, got %q", want, conversation)
		}
	}
	if strings.Contains(conversation, "inline/two-column") || !strings.Contains(diff, "inline/two-column") {
		t.Fatalf("expected layout help only on the diff tab, conversation=%q diff=%q", conversation, diff)
	}
}
