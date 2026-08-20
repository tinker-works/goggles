package projects

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/tinker-works/donsy/netomatic"
	"github.com/tinker-works/goggles/tui/components/statusline"
	"github.com/tinker-works/goggles/tui/theme"
	"github.com/tinker-works/goggles/tui/zones"
)

func TestModel_ShowOrganisations_ShouldResetTheCursor(t *testing.T) {
	m := New().SetOrganisations([]netomatic.Organisation{{Name: "acme"}}).SelectTab(int(Trackers), 2)
	m.Index = 1
	m = m.ShowOrganisations()
	if m.Tab != Organisations || m.Index != 0 {
		t.Fatalf("expected organisations at the top, got tab=%d index=%d", m.Tab, m.Index)
	}
}

func TestModel_SelectRow_ShouldIgnoreAStaleClick(t *testing.T) {
	m := New().SetOrganisations([]netomatic.Organisation{{Name: "acme"}}).SelectTab(int(Trackers), 2)
	m.Index = 1
	if got := m.SelectRow(99, 2).Index; got != 1 {
		t.Fatalf("stale click moved the cursor to %d", got)
	}
}

func TestModel_View_ShouldShowAnInlineOpenFailure(t *testing.T) {
	m := New().SetOrganisations([]netomatic.Organisation{{Name: "acme"}}).Fail(1, errors.New("store unavailable"))
	zones.Init()
	view := m.View(theme.Default(), statusline.Model{}, []netomatic.Project{{ID: 1, Name: "tracker"}}, 100, 20, time.Now())
	if !strings.Contains(view, "store unavailable") {
		t.Fatalf("expected inline error, got:\n%s", view)
	}
}
