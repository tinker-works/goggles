package settings

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/tinker-works/donsy/netomatic"
	"github.com/tinker-works/goggles/tui/components/statusline"
	"github.com/tinker-works/goggles/tui/theme"
	"github.com/tinker-works/goggles/tui/zones"
)

func withPool() Model {
	return New().SetPool([]netomatic.Repository{{Name: "api", FullName: "acme/api"}, {Name: "web", FullName: "acme/web"}}).
		SetLinked([]string{"acme/api", "acme/web"})
}

func TestModel_Filter_ShouldNarrowRepositories(t *testing.T) {
	m := withPool().SetFilter("api")
	name, linked, ok := m.SelectedRepository()
	if !ok || name != "acme/api" || !linked || m.RowCount() != 1 {
		t.Fatalf("unexpected filtered repository: %q %t %t", name, linked, ok)
	}
}

func TestModel_ToggleLinked_ShouldActOnSelectedRepository(t *testing.T) {
	m := withPool().SetFilter("api")
	next, ok := m.ToggleLinked()
	if !ok || len(next) != 1 || next[0] != "acme/web" {
		t.Fatalf("unexpected toggle result: %v", next)
	}
}

func TestModel_SwitchTab_ShouldClearFilter(t *testing.T) {
	m := withPool().SetFilter("api").SwitchTab().SwitchTab().SwitchTab()
	if m.Filter.Value != "" || m.Filter.Active {
		t.Fatalf("filter survived tab cycle: %+v", m.Filter)
	}
}

func TestModel_SelectedRole_ShouldExposePublicAgentSettings(t *testing.T) {
	m := withPool().SetSettings([]netomatic.AgentSettings{{Agent: "opencode", Variant: "high"}}).SelectTab(int(AgentRoles))
	role, profile, ok := m.SelectedRole()
	if !ok || role != Roles[0] || profile.Agent != "opencode" {
		t.Fatalf("unexpected role: %q %+v %t", role, profile, ok)
	}
}

func TestModel_View_ShouldRenderTabsAndRows(t *testing.T) {
	zones.Init()
	view := withPool().View(theme.Default(), statusline.Model{}, "acme", 100, 20)
	if !strings.Contains(view, "Settings") || !strings.Contains(view, "acme/api") {
		t.Fatalf("unexpected settings view:\n%s", view)
	}
}

func TestModel_ApplyFilterKey_ShouldClearOnEscape(t *testing.T) {
	m := withPool().StartFilter()
	m = m.ApplyFilterKey(tea.KeyPressMsg{Code: 'a', Text: "a"}).ApplyFilterKey(tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.Filter.Active || m.Filter.Value != "" {
		t.Fatalf("expected cleared filter: %+v", m.Filter)
	}
}
