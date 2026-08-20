package epicdetail

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/tinker-works/donsy/netomatic"
	"github.com/tinker-works/goggles/tui/components/statusline"
	"github.com/tinker-works/goggles/tui/components/transcript"
	"github.com/tinker-works/goggles/tui/theme"
	"github.com/tinker-works/goggles/tui/viewmodel"
	"github.com/tinker-works/goggles/tui/zones"
)

func renderSized(m Model, status statusline.Model, runner *viewmodel.Runner, entries []transcript.Entry, width, height int) string {
	zones.Init()
	return ansi.Strip(zones.Scan(m.View(theme.Default(), status, runner, entries, width, height, now)))
}

func TestModel_View_ShouldRenderTheEpicHeaderAndBothPanels(t *testing.T) {
	view := renderSized(loaded(), statusline.Model{}.Sync("main", false, now), nil, nil, 180, 40)

	for _, want := range []string{
		"Checkout rewrite", "Ready", "gm/checkout", "ISSUES", "Split cart", "Card form",
		"main", "transition",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected the epic screen to contain %q, got:\n%s", want, view)
		}
	}
}

func TestModel_View_ShouldReportLoadingAndOfflineStates(t *testing.T) {
	loading := renderSized(New(), statusline.Model{}, nil, nil, 120, 30)
	offline := renderSized(New(), statusline.Model{}.Fail(), nil, nil, 120, 30)

	if !strings.Contains(loading, "Loading epic") {
		t.Fatalf("expected a loading panel, got:\n%s", loading)
	}
	if !strings.Contains(offline, "no connection") || !strings.Contains(offline, "retry") {
		t.Fatalf("expected the offline message and retry, got:\n%s", offline)
	}
}

func TestModel_View_ShouldStackThePanelsOnANarrowTerminal(t *testing.T) {
	view := renderSized(loaded(), statusline.Model{}, nil, nil, 70, 40)
	tree := strings.Index(view, "ISSUES")
	detail := strings.Index(view, "the whole thing")
	if tree < 0 || detail < 0 || detail < tree {
		t.Fatalf("expected the detail panel below the tree, got:\n%s", view)
	}
	if lines := strings.Split(view, "\n"); len(lines) < 20 {
		t.Fatalf("expected the stacked layout to be tall, got %d rows", len(lines))
	}
}

func TestModel_Layout_ShouldCapANarrowTreeAtItsNaturalHeight(t *testing.T) {
	got := loaded().layout(theme.Default(), statusline.Model{}, nil, 70, 40, now)
	if got.treeHeight >= got.mainHeight/3 {
		t.Fatalf("expected a short tree to receive less than its target, got %+v", got)
	}
	if got.treeHeight+got.detailHeight != got.mainHeight {
		t.Fatalf("expected narrow panes to consume the main budget, got %+v", got)
	}
}

func TestModel_View_ShouldShowTheSelectedIssueBodyAndThread(t *testing.T) {
	view := renderSized(loaded().MoveDown(), statusline.Model{}, nil, nil, 180, 40)
	for _, want := range []string{"Split cart", "api", "Comments (1)", "looks right"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected the detail panel to contain %q, got:\n%s", want, view)
		}
	}
}

func TestModel_View_ShouldNumberTheDetailPanelAfterItsPullRequest(t *testing.T) {
	view := renderSized(loaded().SelectRow(2), statusline.Model{}, nil, nil, 180, 40)
	if !strings.Contains(view, "#7") || !strings.Contains(view, "PR #7") {
		t.Fatalf("expected the pull request number in the detail panel, got:\n%s", view)
	}
}

func TestModel_View_ShouldHandleEmptyAndUnselectedDetails(t *testing.T) {
	empty := checkout()
	empty.Issues[0].Body = ""
	noDescription := renderSized(New().SetEpic(empty), statusline.Model{}, nil, nil, 180, 40)
	noIssue := renderSized(New().SetEpic(netomatic.Epic{ID: "empty", Title: "Empty", State: "Concept"}),
		statusline.Model{}, nil, nil, 180, 40)

	if !strings.Contains(noDescription, "No description") {
		t.Fatalf("expected the empty-body marker, got:\n%s", noDescription)
	}
	if !strings.Contains(noIssue, "Select an issue") {
		t.Fatalf("expected the empty detail panel, got:\n%s", noIssue)
	}
}

func TestModel_View_ShouldRenderDraftingAndProposalStates(t *testing.T) {
	drafting := netomatic.Epic{ID: "checkout", Title: "Checkout rewrite", State: "Refine"}
	runner := &viewmodel.Runner{
		Run:     netomatic.AgentRun{ID: "run-1", Agent: "refiner", Status: "running"},
		Elapsed: 74 * time.Second,
	}
	draftView := renderSized(New().SetEpic(drafting), statusline.Model{}, runner, nil, 180, 40)
	proposalView := renderSized(loaded().ToggleProposal(), statusline.Model{}, nil, nil, 180, 40)

	for _, want := range []string{"refiner", "round 1", "01:14"} {
		if !strings.Contains(draftView, want) {
			t.Fatalf("expected drafting state to contain %q, got:\n%s", want, draftView)
		}
	}
	if !strings.Contains(proposalView, "proposed") || !strings.Contains(proposalView, "Approving would apply") {
		t.Fatalf("expected the proposal view, got:\n%s", proposalView)
	}
	if strings.Contains(proposalView, "Checkout rewrite") && strings.Contains(proposalView, "ISSUES") {
		if strings.Contains(proposalView, "▾ Checkout rewrite") {
			t.Fatalf("expected the proposal to replace the issue tree, got:\n%s", proposalView)
		}
	}
}

func TestModel_View_ShouldRenderOutputEmptyWaitingAndTranscriptStates(t *testing.T) {
	noRun := renderSized(loaded().ToggleOutput(), statusline.Model{}, nil, nil, 180, 40)
	runner := &viewmodel.Runner{Run: netomatic.AgentRun{ID: "run-1", Agent: "refiner", Status: "running"}}
	waiting := renderSized(loaded().ToggleOutput(), statusline.Model{}, runner, nil, 180, 40)
	output := renderSized(loaded().ToggleOutput(), statusline.Model{}, runner,
		[]transcript.Entry{{Kind: transcript.Text, Text: "writing the plan"}}, 180, 40)

	if !strings.Contains(noRun, "No run is targeting this epic") {
		t.Fatalf("expected the empty output panel, got:\n%s", noRun)
	}
	if !strings.Contains(waiting, "Waiting for the round") {
		t.Fatalf("expected the waiting message, got:\n%s", waiting)
	}
	if !strings.Contains(output, "AGENT OUTPUT") || !strings.Contains(output, "writing the plan") {
		t.Fatalf("expected the transcript output, got:\n%s", output)
	}
}

func TestModel_View_ShouldRenderAtEveryMovedScreenWidth(t *testing.T) {
	for _, width := range []int{60, 90, 110, 200} {
		view := renderSized(loaded(), statusline.Model{}, nil, nil, width, 30)
		if !strings.Contains(view, "Checkout rewrite") {
			t.Fatalf("width %d did not render the epic:\n%s", width, view)
		}
	}
}

func TestModel_DetailWindow_ShouldShrinkAndStayPositive(t *testing.T) {
	m := loaded()
	plain := m.DetailWindow(theme.Default(), nil, 180, 40)
	withOutput := m.ToggleOutput().DetailWindow(theme.Default(), nil, 180, 40)
	narrow := m.DetailWindow(theme.Default(), nil, 70, 40)
	if withOutput >= plain || narrow >= plain {
		t.Fatalf("expected output and stacked layouts to shrink the window, got %d, %d, %d", plain, withOutput, narrow)
	}
	if got := m.ToggleOutput().DetailWindow(theme.Default(), nil, 40, 4); got < 1 {
		t.Fatalf("expected a positive detail window on a tiny terminal, got %d", got)
	}
}

func TestModel_DetailRows_ShouldGrowWithTheBodyAndBeZeroWithoutSelection(t *testing.T) {
	short := New().SetEpic(netomatic.Epic{ID: "short", Title: "Short", Issues: []netomatic.Issue{{ID: "root", Title: "Short"}}})
	longEpic := checkout()
	longEpic.Issues[0].Body = strings.Repeat("a long paragraph of body text. ", 40)
	long := New().SetEpic(longEpic)
	if long.DetailRows(theme.Default(), 120, now) <= short.DetailRows(theme.Default(), 120, now) {
		t.Fatal("expected a longer body to render more rows")
	}
	if got := New().SetEpic(netomatic.Epic{ID: "empty", Title: "Empty"}).DetailRows(theme.Default(), 120, now); got != 0 {
		t.Fatalf("expected no detail rows without a selected issue, got %d", got)
	}
}

func TestModel_View_ShouldScrollTheDetailPanelAndPinTheEditor(t *testing.T) {
	epic := checkout()
	epic.Issues[0].Body = strings.Repeat("scrollable body line\n\n", 40)
	m := New().SetEpic(epic)
	m, _ = m.StartComment()
	m.Resize(100)
	top := renderSized(m, statusline.Model{}, nil, nil, 180, 30)
	scrolled := renderSized(m.ScrollDetail(10, m.DetailRows(theme.Default(), 180, now),
		m.DetailWindow(theme.Default(), nil, 180, 30)), statusline.Model{}, nil, nil, 180, 30)

	if top == scrolled {
		t.Fatal("expected scrolling to change the visible detail rows")
	}
	for _, view := range []string{top, scrolled} {
		if !strings.Contains(view, "ctrl+s post") {
			t.Fatalf("expected the editor pinned in every frame, got:\n%s", view)
		}
	}
}

func TestModel_View_ShouldMarkClippedDetailAndFooterModes(t *testing.T) {
	epic := checkout()
	epic.Issues[0].Body = strings.Repeat("scrollable body line\n\n", 40)
	m := New().SetEpic(epic)
	view := renderSized(m, statusline.Model{}, nil, nil, 180, 30)
	if !strings.Contains(view, "↓") {
		t.Fatalf("expected a detail scroll marker, got:\n%s", view)
	}
	if !strings.Contains(m.Footer(theme.Default(), 200), "transition") {
		t.Fatal("expected the browsing key bar")
	}
	commenting, _ := m.StartComment()
	footer := commenting.Footer(theme.Default(), 200)
	if !strings.Contains(footer, "ctrl+s post") || !strings.Contains(footer, "esc cancel") {
		t.Fatalf("expected the editor key bar, got %q", footer)
	}
}

func TestModel_OutputNavigation_ShouldFollowAndDetachTheTranscript(t *testing.T) {
	m := loaded().ToggleOutput()
	detached := m.ScrollOutput(-3, 20, 5)
	refocused := detached.FocusDetail().FocusOutput()
	rejoined := refocused.ScrollOutput(99, 20, 5)
	hidden := rejoined.ToggleOutput().ToggleOutputFocus()

	if !m.OutputFocused() || m.OutputTop() != 0 || !m.FollowingOutput() {
		t.Fatalf("expected output opened focused at the tail, got %+v", m)
	}
	if detached.OutputTop() != 12 || detached.FollowingOutput() {
		t.Fatalf("expected output detached three rows from the tail, got %+v", detached)
	}
	if rejoined.OutputTop() != 15 || !rejoined.FollowingOutput() || hidden.OutputFocused() {
		t.Fatalf("expected output reattached and hidden, got rejoined=%+v hidden=%+v", rejoined, hidden)
	}
}
