// Package storesetup implements the forced form shown until a project store is
// ready to run.
package storesetup

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tinker-works/donsy/netomatic"
	"github.com/tinker-works/goggles/tui/components/modal"
	"github.com/tinker-works/goggles/tui/theme"
)

const ModalID = "store-setup"

const explain = "This tracker has no agents assigned yet. Pick the model every " +
	"role starts on — you can give each role its own in Settings later."
const constraints = "A model is \"provider/model\", as OpenCode names it."
const noRepositories = "No repositories discovered yet — link them in Settings."

const (
	FieldModel   = 0
	FieldVariant = 1
)

type Model struct {
	form modal.Model
	spec modal.Spec

	Project netomatic.Project
	pool    []string
}

func Spec(pool []string, checked []bool, model, variant string) modal.Spec {
	spec := modal.Spec{ID: ModalID, Title: "Set up tracker", Explain: explain,
		Fields: []modal.Field{{Prompt: "Model", Value: model}, {Prompt: "Variant", Value: variant}},
		Submit: "Initialise", Forced: true}
	if len(pool) > 0 {
		spec.OptionsPrompt = "Repositories"
		spec.Options = append([]string(nil), pool...)
		spec.Checked = make([]bool, len(pool))
		copy(spec.Checked, checked)
	}
	return spec
}

func New(th theme.Theme, project netomatic.Project, width int) (Model, tea.Cmd) {
	spec := Spec(nil, nil, "", "")
	form, cmd := modal.NewForm(spec, th, width), tea.Cmd(nil)
	return Model{form: form, spec: spec, Project: project}, cmd
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	form, cmd := m.form.Update(msg)
	m.form = form
	return m, cmd
}

func (m *Model) Resize(width int) { m.form.Resize(width) }

func (m Model) SetPool(th theme.Theme, pool []string, width int) (Model, tea.Cmd) {
	busy := m.spec.Busy
	ticked := map[string]bool{}
	checked := m.form.Checked()
	for i, repository := range m.pool {
		if i < len(checked) && checked[i] {
			ticked[repository] = true
		}
	}
	m.pool = append([]string(nil), pool...)
	next := make([]bool, len(m.pool))
	for i, repository := range m.pool {
		next[i] = ticked[repository]
	}
	values := m.form.Values()
	m.spec = Spec(m.pool, next, field(values, FieldModel), field(values, FieldVariant))
	m.spec.Busy = busy
	return m.rebuild(th, width)
}

func field(values []string, index int) string {
	if index < 0 || index >= len(values) {
		return ""
	}
	return values[index]
}

func (m Model) Submitting(th theme.Theme, width int) (Model, tea.Cmd) {
	m.spec.Busy = true
	return m.capture().rebuild(th, width)
}

func (m Model) Failed(th theme.Theme, err error, width int) (Model, tea.Cmd) {
	index := FieldModel
	if isRepositoryProblem(err) {
		index = len(m.spec.Fields)
	}
	m = m.capture()
	m.spec = m.spec.WithError(index, err.Error())
	return m.rebuild(th, width)
}

func (m Model) capture() Model {
	values := m.form.Values()
	fields := append([]modal.Field(nil), m.spec.Fields...)
	for i := range fields {
		fields[i].Value = field(values, i)
	}
	m.spec.Fields = fields
	if len(m.spec.Options) > 0 {
		m.spec.Checked = m.form.Checked()
	}
	return m
}

func (m Model) rebuild(th theme.Theme, width int) (Model, tea.Cmd) {
	form, cmd := modal.NewForm(m.spec, th, width), tea.Cmd(nil)
	m.form = form
	return m, cmd
}

func isRepositoryProblem(err error) bool {
	if err == nil {
		return false
	}
	value := strings.ToLower(err.Error())
	return strings.Contains(value, "owner") || strings.Contains(value, "repositor")
}

func (m Model) View(th theme.Theme, width, height int) string {
	lines := []string{m.form.View(), ""}
	if len(m.pool) == 0 {
		lines = append(lines, th.Muted.Render("  "+noRepositories))
	}
	if m.spec.FieldError(FieldModel) == "" && m.spec.FormError() == "" {
		lines = append(lines, th.Muted.Render("  "+constraints))
	}
	block := strings.Join(lines, "\n")
	if width <= 0 || height <= 0 {
		return block
	}
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, block)
}
