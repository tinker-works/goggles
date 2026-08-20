// Package projects implements the project picker and organisation discovery
// screen.
package projects

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/tinker-works/donsy/netomatic"
	"github.com/tinker-works/goggles/tui/components/modal"
	"github.com/tinker-works/goggles/tui/components/panel"
	"github.com/tinker-works/goggles/tui/components/scroll"
	"github.com/tinker-works/goggles/tui/components/statusline"
	"github.com/tinker-works/goggles/tui/components/tabs"
	"github.com/tinker-works/goggles/tui/components/text"
	"github.com/tinker-works/goggles/tui/theme"
	"github.com/tinker-works/goggles/tui/viewmodel"
	"github.com/tinker-works/goggles/tui/zones"
)

type Tab uint8

const (
	Organisations Tab = iota
	Trackers
)

var TabLabels = []string{"Organisations", "Trackers"}

const explain = "A local project stores epics, issues and reviews on this machine."
const constraints = "Letters, numbers and dashes."

const ModalAddProject = "add-project"

const (
	FieldName = 0
	FieldRepo = 1
)

func AddProjectSpec() modal.Spec {
	return modal.Spec{ID: ModalAddProject, Title: "Add project", Explain: explain,
		Fields: []modal.Field{{Prompt: "Name"}}, Submit: "Create"}
}

func AddProjectFailed(spec modal.Spec, err error) modal.Spec {
	field := FieldRepo
	if isNameProblem(err) {
		field = FieldName
	}
	return spec.WithError(field, err.Error())
}

func isNameProblem(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "name")
}

const ModalAddOrganisation = "add-organisation"

func AddOrganisationSpec() modal.Spec {
	return modal.Spec{ID: ModalAddOrganisation, Title: "Add organisation",
		Explain: "The organisation or user account your repositories live under.",
		Fields:  []modal.Field{{Prompt: "Organisation"}}, Submit: "Add"}
}

const ModalSwitchProject = "switch-project"
const AddTrackerChoice = "+ Add a tracker…"

func SwitchProjectSpec(projects []netomatic.Project, current *netomatic.Project) modal.Spec {
	choices := make([]string, 0, len(projects)+1)
	selected := 0
	for i, project := range projects {
		choices = append(choices, project.Name)
		if current != nil && project.ID == current.ID {
			selected = i
		}
	}
	choices = append(choices, AddTrackerChoice)
	return modal.Spec{ID: ModalSwitchProject, Title: "Switch tracker", Choices: choices,
		Selected: selected, Submit: "Open"}
}

type Model struct {
	Tab   Tab
	Index int

	summaries map[uint]viewmodel.ProjectSummary
	loaded    bool

	Organisations   []netomatic.Organisation
	organisationsAt bool
	reposPerOrg     map[string]int

	Err     error
	ErrorAt uint
}

func New() Model {
	return Model{summaries: map[uint]viewmodel.ProjectSummary{}, reposPerOrg: map[string]int{}}
}

func (m Model) SetOrganisations(organisations []netomatic.Organisation) Model {
	m.Organisations = append([]netomatic.Organisation(nil), organisations...)
	sort.Slice(m.Organisations, func(i, j int) bool { return m.Organisations[i].Name < m.Organisations[j].Name })
	if !m.organisationsAt {
		m.organisationsAt = true
		if len(m.Organisations) > 0 {
			m.Tab = Trackers
		}
	}
	return m
}

func (m Model) SetRepositoryPool(pool []netomatic.Repository) Model {
	counts := map[string]int{}
	for _, repository := range pool {
		owner := repository.Organisation
		if owner == "" {
			owner = strings.SplitN(repository.FullName, "/", 2)[0]
		}
		if owner != "" {
			counts[owner]++
		}
	}
	m.reposPerOrg = counts
	return m
}

// CanAddTracker reports whether a project can be created. The daemon validates
// repository linkage during setup, so an empty discovery result is allowed.
func (m Model) CanAddTracker() bool { return true }

func (m Model) SwitchTab(projects int) Model {
	m.Tab = (m.Tab + 1) % Tab(len(TabLabels))
	m.Index = 0
	return m.Clamp(projects)
}

func (m Model) SelectTab(index, projects int) Model {
	if index < 0 || index >= len(TabLabels) {
		return m
	}
	m.Tab, m.Index = Tab(index), 0
	return m.Clamp(projects)
}

func (m Model) ShowOrganisations() Model {
	m.Tab, m.Index = Organisations, 0
	return m
}

func (m Model) RowCount(projects int) int {
	if m.Tab == Organisations {
		return len(m.Organisations)
	}
	return projects
}

func (m Model) MoveUp() Model {
	m.Index = max(0, m.Index-1)
	return m
}

func (m Model) MoveDown(projects int) Model {
	m.Index = min(max(0, m.RowCount(projects)-1), m.Index+1)
	return m
}

func (m Model) Clamp(projects int) Model {
	m.Index = max(0, min(m.Index, max(0, m.RowCount(projects)-1)))
	return m
}

func (m Model) SelectRow(index, projects int) Model {
	if index >= 0 && index < m.RowCount(projects) {
		m.Index = index
	}
	return m
}

func (m Model) SelectedOrganisation() (string, bool) {
	if m.Tab != Organisations || m.Index < 0 || m.Index >= len(m.Organisations) {
		return "", false
	}
	return m.Organisations[m.Index].Name, true
}

func (m Model) SetSummaries(summaries []viewmodel.ProjectSummary) Model {
	m.summaries = make(map[uint]viewmodel.ProjectSummary, len(summaries))
	for _, summary := range summaries {
		m.summaries[summary.ProjectID] = summary
	}
	m.loaded = true
	return m
}

func (m Model) Fail(projectID uint, err error) Model {
	m.Err, m.ErrorAt = err, projectID
	return m
}

func (m Model) ClearError() Model {
	m.Err, m.ErrorAt = nil, 0
	return m
}

func (m Model) View(th theme.Theme, status statusline.Model, list []netomatic.Project, width, height int, now time.Time) string {
	inner := panel.ContentWidth(width)
	head := []string{tabs.Render(th, TabLabels, int(m.Tab)), strings.Repeat("─", inner)}
	body, tail := m.tabRows(th, list, inner, now)
	window := max(1, panel.ContentHeight(height)-len(head)-len(tail))
	offset := scroll.Follow(len(body), window, m.Index)
	rows := append(append(append([]string{}, head...), scroll.Window(body, window, offset)...), tail...)
	return panel.Render(th, "goggles"+status.TitleSuffix()+scroll.Mark(len(body), window, offset),
		strings.Join(rows, "\n"), width, height, false)
}

func (m Model) tabRows(th theme.Theme, list []netomatic.Project, width int, now time.Time) (body, tail []string) {
	if m.Tab == Organisations {
		return m.organisationRows(th, width), m.organisationHints(th)
	}
	return m.trackerRows(th, list, width, now), m.trackerHints(th)
}

func (m Model) organisationRows(th theme.Theme, width int) []string {
	if len(m.Organisations) == 0 {
		return []string{th.Subtitle.Render("Git-backed epics, issues, and pull requests"), "",
			text.Truncate(th.Muted.Render(explain), width), "", th.Warning.Render("No organisation registered yet.")}
	}
	rows := make([]string, 0, len(m.Organisations))
	for i, organisation := range m.Organisations {
		marker, style := "  ", th.Muted
		if i == m.Index {
			marker, style = th.Selected.Render("▌ "), th.Selected
		}
		name := style.Render(text.Pad(organisation.Name, min(28, max(12, width/2))))
		rows = append(rows, zones.Mark(zones.OrganisationRow(i), text.Truncate(marker+name+
			th.Muted.Render(repositoriesLabel(m.reposPerOrg[organisation.Name])), width)))
	}
	return rows
}

func (m Model) organisationHints(th theme.Theme) []string {
	if len(m.Organisations) == 0 {
		return []string{"", th.Accent.Render("S") + th.Muted.Render("  discover your organisations") +
			th.Muted.Render("   ") + th.Accent.Render("N") + th.Muted.Render("  name one yourself")}
	}
	return []string{"", th.Accent.Render("S") + th.Muted.Render("  sync repositories") +
		th.Muted.Render("   ") + th.Accent.Render("N") + th.Muted.Render("  add an organisation")}
}

func repositoriesLabel(count int) string {
	switch count {
	case 0:
		return "no repositories discovered — S to sync"
	case 1:
		return "1 repository"
	default:
		return fmt.Sprintf("%d repositories", count)
	}
}

func (m Model) trackerRows(th theme.Theme, list []netomatic.Project, width int, now time.Time) []string {
	if len(list) == 0 {
		return []string{text.Truncate(th.Muted.Render(explain), width), "", th.Muted.Render(constraints)}
	}
	rows := []string{}
	for i, project := range list {
		rows = append(rows, m.row(th, project, i, width, now)...)
		if m.Err != nil && m.ErrorAt == project.ID {
			rows = append(rows, th.Error.Render("      "+text.Truncate(m.Err.Error(), max(1, width-6))))
		}
		rows = append(rows, "")
	}
	return rows
}

func (m Model) trackerHints(th theme.Theme) []string {
	return []string{"", th.Accent.Render("⏎") + th.Muted.Render("  open") + th.Muted.Render("   ") +
		th.Accent.Render("N") + th.Muted.Render("  add a tracker")}
}

// Footer is the compact key bar kept separate from the list so the root can
// reserve its row before rendering the screen body.
func (m Model) Footer(th theme.Theme, width int) string {
	return th.Muted.Render(text.Truncate("j/k move   d switch tab   enter open   N add   x remove   ? keys", width))
}

func (m Model) row(th theme.Theme, project netomatic.Project, index, width int, now time.Time) []string {
	marker, style := "  ", th.Muted
	if index == m.Index {
		marker, style = th.Selected.Render("▌ "), th.Selected
	}
	nameWidth := min(24, max(12, width/3))
	name := style.Render(text.Pad(project.Name, nameWidth))
	summary := text.Pad(m.summaryOf(th, project), min(26, max(14, width/3)))
	recency := th.Muted.Render(projectAge(project, now))
	return []string{zones.Mark(zones.ProjectRow(index), text.Truncate(marker+name+summary+recency, width))}
}

func (m Model) summaryOf(th theme.Theme, project netomatic.Project) string {
	if !m.loaded {
		return th.Muted.Render("…")
	}
	summary, ok := m.summaries[project.ID]
	if !ok {
		return th.Muted.Render("…")
	}
	if summary.Err != nil {
		return th.Warning.Render("unreadable")
	}
	epics := fmt.Sprintf("%d epics", summary.Epics)
	if summary.Epics == 1 {
		epics = "1 epic"
	}
	running := th.Muted.Render("idle")
	if summary.Running > 0 {
		running = th.Success.Render(fmt.Sprintf("%d running", summary.Running))
	}
	return th.Muted.Render(epics+" · ") + running
}

func projectAge(project netomatic.Project, now time.Time) string {
	if project.LastOpenedAt == "" {
		return "never opened"
	}
	parsed, err := time.Parse(time.RFC3339Nano, project.LastOpenedAt)
	if err != nil {
		return "never opened"
	}
	return "opened " + statusline.Age(parsed, now) + " ago"
}
