package storesetup

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/tinker-works/donsy/netomatic"
	"github.com/tinker-works/goggles/tui/theme"
	"github.com/tinker-works/goggles/tui/zones"
)

const setupWidth = 100

func setupTheme() theme.Theme { return theme.Default() }

func typeSetupText(t *testing.T, m Model, value string) Model {
	t.Helper()
	for _, r := range value {
		m, _ = m.Update(tea.KeyPressMsg(tea.Key{Text: string(r), Code: r}))
	}
	return m
}

func setupScreen(t *testing.T) Model {
	t.Helper()
	m, _ := New(setupTheme(), netomatic.Project{Name: "tracker"}, setupWidth)
	return m
}

func TestSpec_ShouldForceSetupAndOmitAnEmptyRepositoryList(t *testing.T) {
	spec := Spec(nil, nil, "", "")
	if !spec.Forced || len(spec.Options) != 0 || len(spec.Fields) != 2 {
		t.Fatalf("unexpected setup spec: %+v", spec)
	}
}

func TestModel_SetPool_ShouldPreserveTypedValuesAndNamedChecks(t *testing.T) {
	m := typeSetupText(t, setupScreen(t), "provider/model")
	m, _ = m.SetPool(setupTheme(), []string{"acme/widgets", "acme/gadgets"}, setupWidth)
	m.spec.Checked = []bool{false, true}
	m, _ = m.rebuild(setupTheme(), setupWidth)
	m, _ = m.SetPool(setupTheme(), []string{"acme/gadgets", "acme/widgets"}, setupWidth)
	if got := m.spec.Fields[FieldModel].Value; got != "provider/model" {
		t.Fatalf("typed model was lost: %q", got)
	}
	if !m.spec.Checked[0] || m.spec.Checked[1] {
		t.Fatalf("checks did not follow repository names: %v", m.spec.Checked)
	}
}

func TestModel_SetPool_ShouldKeepTheFormBusyWhileSubmitting(t *testing.T) {
	m := setupScreen(t)
	m, _ = m.Submitting(setupTheme(), setupWidth)
	m, _ = m.SetPool(setupTheme(), []string{"acme/widgets"}, setupWidth)
	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	if cmd != nil {
		t.Fatal("repository discovery re-enabled a submitting form")
	}
}

func TestModel_Failed_ShouldRouteRepositoryErrorsToTheForm(t *testing.T) {
	m := setupScreen(t)
	m, _ = m.SetPool(setupTheme(), []string{"other/repo"}, setupWidth)
	m, _ = m.Failed(setupTheme(), fmt.Errorf("repository owner is not allowed"), setupWidth)
	if m.spec.FormError() == "" || m.spec.FieldError(FieldModel) != "" {
		t.Fatalf("expected a form error, got %+v", m.spec.Errors)
	}
	if !strings.Contains(m.View(setupTheme(), setupWidth, 25), "repository owner") {
		t.Fatal("expected the setup error to be visible")
	}
}

func TestModel_View_ShouldExplainAnEmptyRepositoryPool(t *testing.T) {
	zones.Init()
	if view := setupScreen(t).View(setupTheme(), setupWidth, 25); !strings.Contains(view, "No repositories discovered") {
		t.Fatalf("expected empty-pool guidance, got:\n%s", view)
	}
}
