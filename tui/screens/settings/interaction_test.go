package settings

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/tinker-works/donsy/netomatic"
	"github.com/tinker-works/goggles/tui/components/statusline"
	"github.com/tinker-works/goggles/tui/theme"
	"github.com/tinker-works/goggles/tui/zones"
)

func settingsPool(extra int) Model {
	pool := []netomatic.Repository{
		{Name: "api", FullName: "acme/api"},
		{Name: "web", FullName: "acme/web"},
	}
	for i := range extra {
		pool = append(pool, netomatic.Repository{Name: fmt.Sprintf("svc-%02d", i), FullName: fmt.Sprintf("acme/svc-%02d", i)})
	}
	return New().SetPool(pool).SetLinked([]string{"acme/api", "acme/web"})
}

func settingsTyped(m Model, value string) Model {
	m = m.StartFilter()
	for _, r := range value {
		m = m.ApplyFilterKey(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return m
}

func TestModel_SetSettingsAndSelectedRole_ShouldHandleAssignedAndMissingProfiles(t *testing.T) {
	assigned := withPool().SetSettings([]netomatic.AgentSettings{
		{Agent: "opencode", Variant: "high"},
		{Agent: "opencode", Variant: "high"},
	}).SelectTab(int(AgentRoles))
	role, profile, ok := assigned.SelectedRole()
	if !ok || role != Roles[0] || profile.Agent != "opencode" || profile.Variant != "high" {
		t.Fatalf("expected the assigned first role, got %q %+v %t", role, profile, ok)
	}
	missing := withPool().SelectTab(int(AgentRoles))
	role, profile, ok = missing.SelectedRole()
	if !ok || role == "" || profile.Agent != "" {
		t.Fatalf("expected a named missing role, got %q %+v %t", role, profile, ok)
	}
	if _, _, ok := withPool().SelectedRole(); ok {
		t.Fatal("expected no selected role outside the roles tab")
	}
}

func TestModel_SetEpics_ShouldCountEachEpicOncePerRepository(t *testing.T) {
	m := withPool().SetEpics([]netomatic.Epic{
		{ID: "checkout", Repositories: []string{"acme/api", "acme/api"}, Issues: []netomatic.Issue{
			{ID: "root"}, {ID: "card", ParentID: "root", Repository: "acme/web"},
		}},
		{ID: "search", Issues: []netomatic.Issue{{ID: "root"}, {ID: "index", ParentID: "root", Repository: "acme/api"}}},
	})
	if m.EpicsPerRepo["acme/api"] != 2 || m.EpicsPerRepo["acme/web"] != 1 {
		t.Fatalf("expected repository epic counts, got %v", m.EpicsPerRepo)
	}
}

func TestModel_View_ShouldLabelRepositoryEpicCountsAndRoleWarnings(t *testing.T) {
	zones.Init()
	for count, want := range map[int]string{0: "no epics", 1: "1 epic", 4: "4 epics"} {
		epics := make([]netomatic.Epic, 0, count)
		for i := range count {
			epics = append(epics, netomatic.Epic{ID: fmt.Sprintf("epic-%d", i), Repositories: []string{"acme/api"}})
		}
		view := withPool().SetEpics(epics).View(theme.Default(), statusline.Model{}, "acme", 120, 30)
		if !strings.Contains(view, want) {
			t.Fatalf("expected %q in repository view, got:\n%s", want, view)
		}
	}
	assigned := withPool().SetSettings(netomatic.AgentSettings{Agent: "opencode", Variant: "high"}).SelectTab(int(AgentRoles))
	missing := withPool().SelectTab(int(AgentRoles))
	assignedView := assigned.View(theme.Default(), statusline.Model{}, "acme", 140, 30)
	missingView := missing.View(theme.Default(), statusline.Model{}, "acme", 140, 30)
	if !strings.Contains(assignedView, "refiner") || !strings.Contains(assignedView, "opencode") {
		t.Fatalf("expected assigned roles in view, got:\n%s", assignedView)
	}
	if !strings.Contains(missingView, "no agent set") {
		t.Fatalf("expected missing-profile warning, got:\n%s", missingView)
	}
}

func TestModel_ThemeSelection_ShouldSyncKnownAndIgnoreUnknownNames(t *testing.T) {
	m := withPool().SelectTab(int(Appearance))
	names := theme.Names()
	target := names[len(names)-1]
	if got := m.SyncTheme(target).SelectedTheme(); got != target {
		t.Fatalf("expected %q selected, got %q", target, got)
	}
	m = m.MoveDown()
	before := m.SelectedTheme()
	if got := m.SyncTheme("not-a-theme").SelectedTheme(); got != before {
		t.Fatalf("expected unknown theme to leave %q selected, got %q", before, got)
	}
	m.ThemeIndex = len(names) + 5
	if got := m.SelectedTheme(); got != theme.DefaultPaletteName {
		t.Fatalf("expected the default fallback theme, got %q", got)
	}
}

func TestModel_SelectTab_ShouldIgnoreInvalidTabsAndClearFilter(t *testing.T) {
	m := withPool().SelectTab(int(Appearance))
	if m.SelectTab(-1).Tab != Appearance || m.SelectTab(len(TabLabels)).Tab != Appearance {
		t.Fatal("expected invalid tab selections to be ignored")
	}
	filtered := settingsTyped(settingsPool(4), "svc").MoveDown().SelectTab(int(AgentRoles))
	if filtered.Filter.Value != "" || filtered.Index != 0 {
		t.Fatalf("expected tab change to clear the filter and rewind, got %+v", filtered)
	}
}

func TestModel_Cursors_ShouldStopAtTheEndsAndFollowClicks(t *testing.T) {
	repositories := settingsPool(4)
	for range 50 {
		repositories = repositories.MoveDown()
	}
	if repositories.Index != repositories.RowCount()-1 {
		t.Fatalf("expected the last repository row, got %d of %d", repositories.Index, repositories.RowCount())
	}
	for range 50 {
		repositories = repositories.MoveUp()
	}
	if repositories.Index != 0 {
		t.Fatalf("expected the first repository row, got %d", repositories.Index)
	}
	appearance := withPool().SelectTab(int(Appearance)).SelectRow(1)
	for range 50 {
		appearance = appearance.MoveDown()
	}
	if appearance.ThemeIndex != len(theme.Names())-1 {
		t.Fatalf("expected the last palette, got %d", appearance.ThemeIndex)
	}
	clickedAppearance := withPool().SelectTab(int(Appearance)).SelectRow(1)
	if withPool().SelectRow(1).Index != 1 || clickedAppearance.ThemeIndex != 1 {
		t.Fatal("expected clicks to focus the selected repository and palette")
	}
	if withPool().SelectRow(-1).Index != 0 || clickedAppearance.SelectRow(999).ThemeIndex != 1 {
		t.Fatal("expected clicks outside either list to be ignored")
	}
}

func TestModel_Filter_ShouldHandleMatchingRowsAndEscape(t *testing.T) {
	m := settingsPool(4)
	if got := m.SetFilter("SVC-0").RowCount(); got != 4 {
		t.Fatalf("expected case-insensitive service matches, got %d", got)
	}
	narrowed := m.SetFilter("api")
	name, linked, ok := narrowed.SelectedRepository()
	if !ok || name != "acme/api" || !linked || narrowed.RowCount() != 1 {
		t.Fatalf("expected the linked repository match, got %q %t %t rows=%d", name, linked, ok, narrowed.RowCount())
	}
	if settingsTyped(m.MoveDown(), "api").Index != 0 {
		t.Fatal("expected filtering to clamp the cursor")
	}
	cleared := settingsTyped(m, "api").ApplyFilterKey(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cleared.Filter.Active || cleared.Filter.Value != "" || cleared.Index != 0 {
		t.Fatalf("expected escape to clear and rewind the filter, got %+v", cleared)
	}
	if settingsPool(0).SelectTab(int(AgentRoles)).StartFilter().Filter.Active {
		t.Fatal("expected filtering to stay closed off the repositories tab")
	}
}

func TestModel_SelectedRepositoryAndToggleLinked_ShouldUseTheFilteredRow(t *testing.T) {
	m := settingsPool(4).SetFilter("svc-02")
	next, ok := m.ToggleLinked()
	if !ok || strings.Join(next, ",") != "acme/api,acme/svc-02,acme/web" {
		t.Fatalf("expected the filtered repository toggled, got %v", next)
	}
	if _, _, ok := withPool().SelectTab(int(Appearance)).SelectedRepository(); ok {
		t.Fatal("expected no selected repository outside the repositories tab")
	}
}

func TestModel_View_ShouldBoundLongListsAndKeepTheSelectionVisible(t *testing.T) {
	zones.Init()
	m := settingsPool(60)
	const height = 20
	view := m.View(theme.Default(), statusline.Model{}, "acme-platform", 100, height)
	if len(strings.Split(view, "\n")) != height || !strings.Contains(view, "↓") {
		t.Fatalf("expected a bounded long repository panel, got:\n%s", view)
	}
	for range 40 {
		m = m.MoveDown()
	}
	view = m.View(theme.Default(), statusline.Model{}, "acme-platform", 100, height)
	name, _, _ := m.SelectedRepository()
	if !strings.Contains(view, name) || strings.Contains(view, "acme/api") || !strings.Contains(view, "↕") {
		t.Fatalf("expected the selected row and scroll marker, got:\n%s", view)
	}
	appearance := settingsPool(0).SelectTab(int(Appearance))
	appearanceView := appearance.View(theme.Default(), statusline.Model{}, "acme-platform", 100, 8)
	if !strings.Contains(appearanceView, "use the selected theme") || len(strings.Split(appearanceView, "\n")) != 8 {
		t.Fatalf("expected the appearance hint pinned below the list, got:\n%s", appearanceView)
	}
}

func TestModel_View_ShouldRenderFilterEmptyStateAndFooter(t *testing.T) {
	view := settingsTyped(settingsPool(4), "nope").View(theme.Default(), statusline.Model{}, "acme", 100, 20)
	if !strings.Contains(view, "/nope") || !strings.Contains(view, "No repositories match nope") {
		t.Fatalf("expected the filter prompt and empty state, got:\n%s", view)
	}
	footer := withPool().Footer(theme.Default(), 200)
	for _, want := range []string{"switch tab", "toggle"} {
		if !strings.Contains(footer, want) {
			t.Fatalf("expected %q in the footer, got %q", want, footer)
		}
	}
	repositories := settingsPool(20).ScrollBy(3)
	appearance := settingsPool(20).SelectTab(int(Appearance)).ScrollBy(3).ScrollBy(-1)
	if repositories.Index != 3 || appearance.ThemeIndex != 2 || settingsPool(20).ScrollBy(-5).Index != 0 {
		t.Fatalf("expected scroll to move the active cursor, repositories=%d appearance=%d", repositories.Index, appearance.ThemeIndex)
	}
}
