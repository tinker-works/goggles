package projects

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

func projectTheme() theme.Theme { return theme.Default() }

func projectList() []netomatic.Project {
	return []netomatic.Project{{ID: 1, Name: "acme-platform"}, {ID: 2, Name: "acme-web"}}
}

func TestModel_SetOrganisations_ShouldChooseTheUsefulTabAndSort(t *testing.T) {
	m := New().SetOrganisations([]netomatic.Organisation{{Name: "globex"}, {Name: "acme"}})
	if m.Tab != Trackers || m.Organisations[0].Name != "acme" {
		t.Fatalf("unexpected initial project picker state: tab=%d orgs=%v", m.Tab, m.Organisations)
	}
	got, ok := m.SelectTab(int(Organisations), 2).SelectedOrganisation()
	if !ok || got != "acme" {
		t.Fatalf("expected the first organisation, got %q", got)
	}
}

func TestSwitchProjectSpec_ShouldPreselectTheCurrentProjectAndOfferAdding(t *testing.T) {
	projects := projectList()
	spec := SwitchProjectSpec(projects, &projects[1])
	if spec.Choices[spec.Selected] != "acme-web" || spec.Choices[len(spec.Choices)-1] != AddTrackerChoice {
		t.Fatalf("unexpected switcher: %+v", spec)
	}
}

func TestModel_View_ShouldRenderSummariesAndDaemonTimestamp(t *testing.T) {
	now := time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC)
	projects := projectList()
	projects[0].LastOpenedAt = now.Add(-2 * time.Minute).Format(time.RFC3339)
	m := New().SetOrganisations([]netomatic.Organisation{{Name: "acme"}}).
		SetSummaries([]viewmodel.ProjectSummary{{ProjectID: 1, Epics: 3, Running: 1}, {ProjectID: 2, Epics: 1}})
	zones.Init()
	view := m.View(projectTheme(), statusline.Model{}, projects, 120, 30, now)
	for _, want := range []string{"3 epics", "1 running", "opened 2m00s ago", "never opened"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected %q in picker:\n%s", want, view)
		}
	}
}

func TestModel_Error_ShouldStayAttachedToTheSelectedProject(t *testing.T) {
	m := New().SetOrganisations([]netomatic.Organisation{{Name: "acme"}}).Fail(2, errProjectOpen{})
	zones.Init()
	view := m.View(projectTheme(), statusline.Model{}, projectList(), 100, 20, time.Now())
	if !strings.Contains(view, "could not open project") {
		t.Fatalf("expected an inline error, got:\n%s", view)
	}
}

type errProjectOpen struct{}

func (errProjectOpen) Error() string { return "could not open project" }
