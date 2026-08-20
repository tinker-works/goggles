package board

import (
	"strings"
	"testing"
	"time"

	"github.com/tinker-works/donsy/netomatic"
	"github.com/tinker-works/goggles/tui/components/statusline"
	"github.com/tinker-works/goggles/tui/theme"
	"github.com/tinker-works/goggles/tui/viewmodel"
	"github.com/tinker-works/goggles/tui/zones"
)

var boardNow = time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC)

func boardEpics() []netomatic.Epic {
	return []netomatic.Epic{
		{ID: "checkout", Title: "Checkout rewrite", State: "Ready", Repositories: []string{"api"},
			Issues: []netomatic.Issue{{ID: "root", Title: "Checkout rewrite"},
				{ID: "cart", ParentID: "root", Title: "Split cart", Repository: "api", State: "open"}}},
		{ID: "closed", Title: "Old idea", State: "Closed", Issues: []netomatic.Issue{{ID: "root"}}},
	}
}

func loadedBoard() Model { return New().SetEpics(boardEpics()) }

func TestModel_ShouldHideClosedEpicsAndExposeTheOpenCardAfterZoom(t *testing.T) {
	m := loadedBoard()
	if len(m.Lanes()) != 1 || m.Lanes()[0].Key != "checkout" {
		t.Fatalf("unexpected lanes: %+v", m.Lanes())
	}
	m = m.Zoom()
	issue, ok := m.FocusedIssue()
	if !ok || issue.Issue.ID != "cart" || m.Column != viewmodel.OpenColumn {
		t.Fatalf("expected first open issue after zoom, got %+v column=%d", issue, m.Column)
	}
}

func TestModel_SetEpics_ShouldKeepTheFocusedLaneWhenRowsArrive(t *testing.T) {
	m := loadedBoard()
	m = m.SetEpics(append([]netomatic.Epic{{ID: "billing", Title: "Billing", State: "Ready"}}, boardEpics()...)).MoveDown()
	focused := m.Lanes()[m.Lane].Key
	m = m.SetEpics(append([]netomatic.Epic{{ID: "new", Title: "New", State: "Ready"}}, m.Epics...))
	if m.Lanes()[m.Lane].Key != focused {
		t.Fatalf("selection moved from %q to %q", focused, m.Lanes()[m.Lane].Key)
	}
}

func TestModel_Filter_ShouldDimWithoutRemovingLanes(t *testing.T) {
	m := loadedBoard()
	count := len(m.Lanes())
	m = m.SetFilter("does-not-match")
	if len(m.Lanes()) != count || m.Lanes()[0].Matched {
		t.Fatalf("filter changed lane geometry: %+v", m.Lanes())
	}
}

func TestModel_RecordActivity_ShouldUseGrowthBetweenSamples(t *testing.T) {
	m := loadedBoard().SetRuns([]netomatic.AgentRun{{ID: "run", Agent: "coding", Project: "tracker", Status: "running"}}, boardNow)
	m = m.RecordActivity(map[string]int64{"run": 100}).RecordActivity(map[string]int64{"run": 175})
	if got := m.ActivityFor("run"); len(got) != 1 || got[0] != 75 {
		t.Fatalf("unexpected activity trace: %v", got)
	}
}

func TestModel_View_ShouldMatchLiveRunnerByItsPresentationSubject(t *testing.T) {
	zones.Init()
	m := loadedBoard()
	m.Runners = []viewmodel.Runner{{
		Run:     netomatic.AgentRun{ID: "run", Agent: "coding", Project: "tracker", Status: "running"},
		Subject: "Checkout rewrite",
	}}
	view := m.View(theme.Default(), &netomatic.Project{Name: "tracker"}, statusline.Model{}, 140, 30, boardNow)
	if !strings.Contains(view, "⚙") {
		t.Fatalf("expected the live runner indicator in the matching epic: %s", view)
	}

	zoom := m.Zoom().ZoomView(theme.Default(), &netomatic.Project{Name: "tracker"}, statusline.Model{}, 140, 30, boardNow)
	if !strings.Contains(zoom, "⚙") {
		t.Fatalf("expected runner details in the matching epic: %s", zoom)
	}
}

func TestModel_View_ShouldRenderListAndZoomColumns(t *testing.T) {
	zones.Init()
	m := loadedBoard()
	status := statusline.Model{}.Sync("main", false, boardNow)
	list := m.View(theme.Default(), &netomatic.Project{Name: "acme"}, status, 140, 30, boardNow)
	if !strings.Contains(list, "Checkout rewrite") || strings.Contains(list, "CODING") {
		t.Fatalf("unexpected list view:\n%s", list)
	}
	zoom := m.Zoom().ZoomView(theme.Default(), &netomatic.Project{Name: "acme"}, status, 140, 30, boardNow)
	for _, want := range []string{"OPEN", "CODING", "IN PR", "Split cart"} {
		if !strings.Contains(zoom, want) {
			t.Fatalf("expected %q in zoom view:\n%s", want, zoom)
		}
	}
}
