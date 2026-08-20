package modal

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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

func TestFormEscape_ShouldEmitCancellationForCancellableForms(t *testing.T) {
	m := NewForm(Spec{ID: "add-project", Fields: []Field{{Prompt: "Name"}}}, theme.Default(), 80)
	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))
	msg, ok := cmd().(CancelledMsg)
	if !ok || msg.ID != "add-project" {
		t.Fatalf("unexpected cancellation: %#v", msg)
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

func TestFormOptions_ShouldNavigateWithinTheChecklist(t *testing.T) {
	m := NewForm(Spec{ID: "setup", Fields: []Field{{Prompt: "Model"}, {Prompt: "Variant"}},
		Options: []string{"acme/widgets", "acme/gadgets"}, OptionsPrompt: "Repositories"}, theme.Default(), 80)
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyTab}))
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyTab}))
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeySpace}))
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeySpace}))

	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	msg, ok := cmd().(SubmittedMsg)
	if !ok || len(msg.Options) != 2 || msg.Options[0] != "acme/widgets" || msg.Options[1] != "acme/gadgets" {
		t.Fatalf("unexpected checklist submission: %#v", msg)
	}
}

func TestFormOptions_ShouldRenderThePrompt(t *testing.T) {
	m := NewForm(Spec{ID: "setup", Options: []string{"acme/widgets"}, OptionsPrompt: "Repositories"}, theme.Default(), 80)
	if view := zones.Scan(m.View()); !strings.Contains(view, "Repositories") {
		t.Fatalf("expected checklist prompt in modal: %q", view)
	}
}

func TestFormPaste_ShouldAppendToTheFocusedField(t *testing.T) {
	m := NewForm(Spec{ID: "setup", Fields: []Field{{Prompt: "Model"}}}, theme.Default(), 80)
	m, _ = m.Update(tea.PasteMsg{Content: "provider/model"})
	if values := m.Values(); len(values) != 1 || values[0] != "provider/model" {
		t.Fatalf("unexpected pasted value: %v", values)
	}
}

func TestFormBodyAndCycle_ShouldSubmitBothValues(t *testing.T) {
	m := NewForm(Spec{ID: "comment", Fields: []Field{{Prompt: "Title"}}, Body: true,
		Cycle: []string{"Issue", "Pull request"}, Submit: "Post"}, theme.Default(), 80)
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Text: "title", Code: 'e'}))
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyTab}))
	m, _ = m.Update(tea.PasteMsg{Content: "comment body"})
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Text: "t", Code: 't'}))

	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	msg, ok := cmd().(SubmittedMsg)
	if !ok || msg.Values[0] != "title" || msg.Body != "comment body" || msg.Cycle != 1 {
		t.Fatalf("unexpected body submission: %#v", msg)
	}
}

func TestFormBody_ShouldRenderAndAcceptMouseFocus(t *testing.T) {
	zones.Init()
	m := NewForm(Spec{ID: "epic", Body: true, Submit: "Create"}, theme.Default(), 80)
	view := zones.Scan(m.View())
	if !strings.Contains(view, "Body:") {
		t.Fatalf("expected body control in modal: %q", view)
	}
	bodyX, bodyY, ok := zones.Bounds(zones.ModalBody)
	if !ok {
		t.Fatal("body zone was not rendered")
	}
	m, _ = m.Update(tea.MouseClickMsg{X: bodyX, Y: bodyY, Button: tea.MouseLeft})
	m, _ = m.Update(tea.PasteMsg{Content: "details"})
	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	msg, ok := cmd().(SubmittedMsg)
	if !ok || msg.Body != "details" {
		t.Fatalf("unexpected body mouse submission: %#v", msg)
	}
}

func TestFormBody_ShouldRenderAFiveRowTextareaAndNavigateLines(t *testing.T) {
	m := NewForm(Spec{ID: "epic", Body: true, Submit: "Create"}, theme.Default(), 80)
	m, _ = m.Update(tea.PasteMsg{Content: "first\nsecond"})
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyUp}))
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Text: "!", Code: '!'}))

	if view := m.View(); lipgloss.Height(view) < 7 {
		t.Fatalf("expected a five-row body textarea, got:\n%s", view)
	}
	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	msg, ok := cmd().(SubmittedMsg)
	if !ok || msg.Body != "first!\nsecond" {
		t.Fatalf("unexpected multiline submission: %#v", msg)
	}
}

func TestFormText_ShouldEditAtTheCursor(t *testing.T) {
	m := NewForm(Spec{ID: "project", Fields: []Field{{Prompt: "Name", Value: "ac"}}}, theme.Default(), 80)
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyLeft}))
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Text: "b", Code: 'b'}))
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyHome}))
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Text: "z", Code: 'z'}))
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyDelete}))
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnd}))
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyBackspace}))
	if values := m.Values(); len(values) != 1 || values[0] != "zb" {
		t.Fatalf("unexpected cursor edit: %v", values)
	}
}

func TestFormText_ShouldLimitAndScrollLongValues(t *testing.T) {
	m := NewForm(Spec{ID: "project", Fields: []Field{{Prompt: "Name"}}}, theme.Default(), 40)
	m, _ = m.Update(tea.PasteMsg{Content: strings.Repeat("a", 201)})
	if values := m.Values(); len(values) != 1 || len([]rune(values[0])) != fieldCharLimit {
		t.Fatalf("expected %d-rune value, got %q", fieldCharLimit, values)
	}
	for _, line := range strings.Split(m.View(), "\n") {
		if width := lipgloss.Width(line); width > 32 {
			t.Fatalf("modal line exceeds its bounded width (%d): %q", width, line)
		}
	}
}

func TestFormWithoutSubmitLabel_ShouldRenderMouseControls(t *testing.T) {
	zones.Init()
	m := NewForm(Spec{ID: "state", Choices: []string{"Ready"}}, theme.Default(), 80)
	zones.Scan(m.View())
	for _, zone := range []string{zones.ModalSubmit, zones.ModalCancel} {
		if _, _, ok := zones.Bounds(zone); !ok {
			t.Fatalf("expected %s zone", zone)
		}
	}
}

func TestFormView_ShouldRenderInsideABoundedPanel(t *testing.T) {
	m := NewForm(Spec{ID: "project", Title: "Add project", Explain: strings.Repeat("details ", 30),
		Fields: []Field{{Prompt: "Name"}}}, theme.Default(), 80)
	view := m.View()
	if !strings.Contains(view, "╭") || !strings.Contains(view, "╰") {
		t.Fatalf("expected a bordered panel: %q", view)
	}
	for _, line := range strings.Split(view, "\n") {
		if width := lipgloss.Width(line); width > 72 {
			t.Fatalf("modal line exceeds its bounded width (%d): %q", width, line)
		}
	}
}
