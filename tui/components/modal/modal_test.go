package modal

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/tinker-works/goggles/tui/theme"
	"github.com/tinker-works/goggles/tui/zones"
)

func TestModalVisibility(t *testing.T) {
	m := New("title", "body")
	if m.View() != "" {
		t.Fatal("new modal is visible")
	}
	m.Show("body")
	if m.View() == "" {
		t.Fatal("shown modal is empty")
	}
	m.Hide()
	if m.View() != "" {
		t.Fatal("hidden modal rendered content")
	}
}

func TestFormBusy_ShouldRefuseASecondSubmit(t *testing.T) {
	m := NewForm(Spec{ID: "setup", Fields: []Field{{Prompt: "Model"}}, Busy: true}, theme.Default(), 80)
	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	if cmd != nil {
		t.Fatal("busy form accepted a second submission")
	}
}

func TestFormMouse_ShouldFocusFieldsToggleOptionsAndSubmit(t *testing.T) {
	zones.Init()
	m := NewForm(Spec{ID: "setup", Fields: []Field{{Prompt: "Model"}, {Prompt: "Variant"}},
		Options: []string{"acme/widgets"}, Submit: "Initialise"}, theme.Default(), 80)
	view := zones.Scan(m.View())
	if view == "" {
		t.Fatal("expected a rendered form")
	}

	fieldX, fieldY, ok := zones.Bounds(zones.ModalField(1))
	if !ok {
		t.Fatal("variant field zone was not rendered")
	}
	m, _ = m.Update(tea.MouseClickMsg{X: fieldX, Y: fieldY, Button: tea.MouseLeft})
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Text: "v", Code: 'v'}))

	optionX, optionY, ok := zones.Bounds(zones.ModalOption(0))
	if !ok {
		t.Fatal("repository option zone was not rendered")
	}
	m, _ = m.Update(tea.MouseClickMsg{X: optionX, Y: optionY, Button: tea.MouseLeft})

	submitX, submitY, ok := zones.Bounds(zones.ModalSubmit)
	if !ok {
		t.Fatal("submit zone was not rendered")
	}
	_, cmd := m.Update(tea.MouseClickMsg{X: submitX, Y: submitY, Button: tea.MouseLeft})
	msg, ok := cmd().(SubmittedMsg)
	if !ok || msg.Values[1] != "v" || len(msg.Options) != 1 || msg.Options[0] != "acme/widgets" {
		t.Fatalf("unexpected mouse submission: %#v", msg)
	}
}

func TestFormChoice_ShouldRenderSelectAndSubmitTheChoice(t *testing.T) {
	zones.Init()
	m := NewForm(Spec{ID: "switch", Title: "Switch", Choices: []string{"one", "two"}, Submit: "Open"}, theme.Default(), 80)
	view := zones.Scan(m.View())
	if !strings.Contains(view, "one") || !strings.Contains(view, "two") {
		t.Fatalf("expected choices in the modal: %q", view)
	}

	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	msg, ok := cmd().(SubmittedMsg)
	if !ok || msg.Choice != 1 {
		t.Fatalf("unexpected choice submission: %#v", msg)
	}
}

func TestFormPaste_ShouldAppendToTheFocusedField(t *testing.T) {
	m := NewForm(Spec{ID: "setup", Fields: []Field{{Prompt: "Model"}}}, theme.Default(), 80)
	m, _ = m.Update(tea.PasteMsg{Content: "provider/model"})
	if values := m.Values(); len(values) != 1 || values[0] != "provider/model" {
		t.Fatalf("unexpected pasted value: %v", values)
	}
}
