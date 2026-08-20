// Package settings implements project repositories, agent profiles, and themes.
package settings

import (
	"fmt"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/tinker-works/donsy/netomatic"
	"github.com/tinker-works/goggles/tui/components/filter"
	"github.com/tinker-works/goggles/tui/components/panel"
	"github.com/tinker-works/goggles/tui/components/scroll"
	"github.com/tinker-works/goggles/tui/components/statusline"
	"github.com/tinker-works/goggles/tui/components/tabs"
	"github.com/tinker-works/goggles/tui/components/text"
	"github.com/tinker-works/goggles/tui/theme"
	"github.com/tinker-works/goggles/tui/zones"
)

type Tab uint8

const (
	Repositories Tab = iota
	AgentRoles
	Appearance
)

var TabLabels = []string{"Repositories", "Agent roles", "Appearance"}

// Roles is the stable pipeline order used by the daemon's settings response.
// The public response intentionally carries profile DTOs without domain types;
// the response order is therefore the role order shown by this screen.
var Roles = []string{"refiner", "issue-reviewer", "coding", "pr-reviewer", "merge"}

type filterState struct {
	Active bool
	Value  string
}

func (f filterState) Start() filterState           { f.Active = true; return f }
func (f filterState) Set(value string) filterState { f.Value = value; return f }

func applyFilterKey(state filterState, msg tea.KeyPressMsg) filterState {
	if msg.Code == tea.KeyEscape {
		return filterState{}
	}
	if msg.Code == tea.KeyEnter {
		state.Active = false
		return state
	}
	if msg.Code == tea.KeyBackspace {
		runes := []rune(state.Value)
		if len(runes) > 0 {
			state.Value = string(runes[:len(runes)-1])
		}
		return state
	}
	if msg.Code == tea.KeySpace {
		state.Value += " "
		return state
	}
	state.Value += msg.Text
	return state
}

type Model struct {
	Tab   Tab
	Index int

	Linked       []string
	Pool         []netomatic.Repository
	Settings     []netomatic.AgentSettings
	EpicsPerRepo map[string]int

	Filter     filterState
	ThemeIndex int
}

func New() Model { return Model{EpicsPerRepo: map[string]int{}} }

func (m Model) SetLinked(linked []string) Model {
	m.Linked = append([]string(nil), linked...)
	return m.clamp()
}
func (m Model) SetPool(pool []netomatic.Repository) Model {
	m.Pool = append([]netomatic.Repository(nil), pool...)
	return m.clamp()
}

// SetSettings accepts either the action response slice or one profile DTO. The
// variadic form keeps callers that load one project profile convenient while the
// screen remains entirely on public netomatic data.
func (m Model) SetSettings(value ...any) Model {
	for _, item := range value {
		switch settings := item.(type) {
		case []netomatic.AgentSettings:
			m.Settings = append([]netomatic.AgentSettings(nil), settings...)
		case netomatic.AgentSettings:
			m.Settings = []netomatic.AgentSettings{settings}
		}
	}
	return m
}

func (m Model) SetEpics(epics []netomatic.Epic) Model {
	counts := map[string]int{}
	for _, epic := range epics {
		seen := map[string]struct{}{}
		for _, repository := range epic.Repositories {
			seen[repository] = struct{}{}
		}
		for _, issue := range epic.Issues {
			if issue.Repository != "" {
				seen[issue.Repository] = struct{}{}
			}
		}
		for repository := range seen {
			counts[repository]++
		}
	}
	m.EpicsPerRepo = counts
	return m
}

func (m Model) SyncTheme(name string) Model {
	for i, candidate := range theme.Names() {
		if candidate == name {
			m.ThemeIndex = i
			break
		}
	}
	return m
}

func (m Model) SwitchTab() Model {
	m.Tab = (m.Tab + 1) % Tab(len(TabLabels))
	m.Index = 0
	m.Filter = filterState{}
	return m.clamp()
}

func (m Model) SelectTab(index int) Model {
	if index < 0 || index >= len(TabLabels) {
		return m
	}
	m.Tab, m.Index, m.Filter = Tab(index), 0, filterState{}
	return m.clamp()
}

func (m Model) StartFilter() Model {
	if m.Tab == Repositories {
		m.Filter = m.Filter.Start()
	}
	return m
}

func (m Model) ApplyFilterKey(msg tea.KeyPressMsg) Model {
	m.Filter = applyFilterKey(m.Filter, msg)
	if !m.Filter.Active && m.Filter.Value == "" {
		m.Index = 0
	}
	return m.clamp()
}

func (m Model) SetFilter(value string) Model { m.Filter = m.Filter.Set(value); return m.clamp() }

func (m Model) ScrollBy(rows int) Model {
	if m.Tab == Appearance {
		m.ThemeIndex += rows
	} else {
		m.Index += rows
	}
	return m.clamp()
}

func (m Model) MoveUp() Model {
	if m.Tab == Appearance {
		m.ThemeIndex = max(0, m.ThemeIndex-1)
	} else {
		m.Index = max(0, m.Index-1)
	}
	return m
}

func (m Model) MoveDown() Model {
	if m.Tab == Appearance {
		m.ThemeIndex = min(len(theme.Names())-1, m.ThemeIndex+1)
	} else {
		m.Index = min(max(0, m.RowCount()-1), m.Index+1)
	}
	return m
}

func (m Model) SelectRow(index int) Model {
	if m.Tab == Appearance {
		if index >= 0 && index < len(theme.Names()) {
			m.ThemeIndex = index
		}
		return m
	}
	if index >= 0 && index < m.RowCount() {
		m.Index = index
	}
	return m
}

func (m Model) clamp() Model {
	m.Index = max(0, min(m.Index, max(0, m.RowCount()-1)))
	m.ThemeIndex = max(0, min(m.ThemeIndex, len(theme.Names())-1))
	return m
}

func (m Model) RowCount() int {
	switch m.Tab {
	case AgentRoles:
		return len(Roles)
	case Appearance:
		return len(theme.Names())
	default:
		return len(m.repositoryRows())
	}
}

func (m Model) repositoryRows() []string {
	linked := map[string]struct{}{}
	rows, unlinked := []string{}, []string{}
	for _, repository := range m.Linked {
		linked[repository] = struct{}{}
		if filter.Matches(m.Filter.Value, repository) {
			rows = append(rows, repository)
		}
	}
	sort.Strings(rows)
	for _, repository := range m.Pool {
		name := repository.FullName
		if name == "" {
			name = repository.Name
		}
		if _, ok := linked[name]; ok {
			continue
		}
		if filter.Matches(m.Filter.Value, name) {
			unlinked = append(unlinked, name)
		}
	}
	sort.Strings(unlinked)
	return append(rows, unlinked...)
}

func (m Model) SelectedRepository() (string, bool, bool) {
	rows := m.repositoryRows()
	if m.Tab != Repositories || m.Index < 0 || m.Index >= len(rows) {
		return "", false, false
	}
	name := rows[m.Index]
	return name, m.isLinked(name), true
}

func (m Model) isLinked(name string) bool {
	for _, repository := range m.Linked {
		if repository == name {
			return true
		}
	}
	return false
}

func (m Model) ToggleLinked() ([]string, bool) {
	name, linked, ok := m.SelectedRepository()
	if !ok {
		return nil, false
	}
	next := make([]string, 0, len(m.Linked)+1)
	for _, repository := range m.Linked {
		if repository != name {
			next = append(next, repository)
		}
	}
	if !linked {
		next = append(next, name)
	}
	sort.Strings(next)
	return next, true
}

func (m Model) SelectedRole() (string, netomatic.AgentSettings, bool) {
	if m.Tab != AgentRoles || m.Index < 0 || m.Index >= len(Roles) {
		return "", netomatic.AgentSettings{}, false
	}
	if m.Index >= len(m.Settings) {
		return Roles[m.Index], netomatic.AgentSettings{}, true
	}
	return Roles[m.Index], m.Settings[m.Index], true
}

func (m Model) SelectedTheme() string {
	names := theme.Names()
	if m.ThemeIndex < 0 || m.ThemeIndex >= len(names) {
		return theme.DefaultPaletteName
	}
	return names[m.ThemeIndex]
}

func (m Model) View(th theme.Theme, status statusline.Model, projectName string, width, height int) string {
	inner := panel.ContentWidth(width)
	head := []string{tabs.Render(th, TabLabels, int(m.Tab))}
	if prompt := m.filterPrompt(th); prompt != "" {
		head = append(head, prompt)
	}
	head = append(head, strings.Repeat("─", inner))
	list, tail := m.tabRows(th, inner)
	window := max(1, panel.ContentHeight(height)-len(head)-len(tail))
	offset := scroll.Follow(len(list), window, m.cursor())
	rows := append([]string(nil), head...)
	rows = append(rows, scroll.Window(list, window, offset)...)
	rows = append(rows, tail...)
	title := "Settings"
	if projectName != "" {
		title = "Settings — " + projectName
	}
	return panel.Render(th, title+status.TitleSuffix()+scroll.Mark(len(list), window, offset), strings.Join(rows, "\n"), width, height, false)
}

func (m Model) cursor() int {
	if m.Tab == Appearance {
		return m.ThemeIndex
	}
	return m.Index
}

func (m Model) filterPrompt(th theme.Theme) string {
	switch {
	case m.Filter.Active:
		return th.Selected.Render("/" + m.Filter.Value + "▏")
	case m.Filter.Value != "":
		return th.Muted.Render("/" + m.Filter.Value)
	default:
		return ""
	}
}

func (m Model) tabRows(th theme.Theme, width int) (list, tail []string) {
	switch m.Tab {
	case AgentRoles:
		return m.roleRows(th, width), []string{"", th.Accent.Render("⏎") + th.Muted.Render("  edit the selected role")}
	case Appearance:
		return m.themeRows(th, width), []string{"", th.Accent.Render("⏎") + th.Muted.Render("  use the selected theme")}
	default:
		return m.repoRows(th, width)
	}
}

func (m Model) Footer(th theme.Theme, width int) string {
	return text.Truncate(th.Muted.Render("j/k move   d switch tab   space toggle   / filter   enter edit   esc back"), width)
}

func (m Model) repoRows(th theme.Theme, width int) (list, tail []string) {
	entries := m.repositoryRows()
	if len(entries) == 0 {
		message := "No repositories discovered — sync your GitHub organisations first."
		if m.Filter.Value != "" {
			message = "No repositories match " + m.Filter.Value
		}
		return []string{th.Muted.Render(text.Truncate(message, width))}, []string{"", th.Accent.Render("N") + th.Muted.Render("  sync and add a repository")}
	}
	rows := make([]string, 0, len(entries))
	for i, name := range entries {
		mark, style := "  ", th.Muted
		if m.isLinked(name) {
			mark = "✓ "
		}
		marker := "  "
		if i == m.Index {
			marker, style = th.Selected.Render("▌ "), th.Selected
		}
		detail := "(not linked)"
		if m.isLinked(name) {
			detail = epicsLabel(m.EpicsPerRepo[name])
		}
		rows = append(rows, zones.Mark(zones.SettingsRow(i), text.Truncate(marker+mark+style.Render(text.Pad(name, min(28, max(12, width/2))))+th.Muted.Render(detail), width)))
	}
	return rows, nil
}

func epicsLabel(count int) string {
	switch count {
	case 0:
		return "no epics"
	case 1:
		return "1 epic"
	default:
		return fmt.Sprintf("%d epics", count)
	}
}

func (m Model) roleRows(th theme.Theme, width int) []string {
	rows := make([]string, 0, len(Roles))
	for i, role := range Roles {
		marker, style := "  ", th.Muted
		if i == m.Index {
			marker, style = th.Selected.Render("▌ "), th.Selected
		}
		profile := netomatic.AgentSettings{}
		if i < len(m.Settings) {
			profile = m.Settings[i]
		}
		assignment := text.Pad(profile.Agent, 12) + text.Pad(profile.Variant, 8)
		state := th.Success.Render("✓")
		if profile.Agent == "" {
			assignment = th.Muted.Render(text.Pad("—", 12) + text.Pad("—", 8))
			state = th.Error.Render("✗ no agent set")
		}
		rows = append(rows, zones.Mark(zones.SettingsRow(i), text.Truncate(marker+style.Render(text.Pad(role, 16))+assignment+state, width)))
	}
	return rows
}

func (m Model) themeRows(th theme.Theme, width int) []string {
	rows := []string{}
	for i, palette := range theme.Palettes() {
		marker, style := "  ", th.Muted
		if i == m.ThemeIndex {
			marker, style = th.Selected.Render("▌ "), th.Selected
		}
		active := "  "
		if palette.Name == th.Name {
			active = th.Success.Render("✓ ")
		}
		rows = append(rows, zones.Mark(zones.SettingsRow(i), text.Truncate(marker+active+style.Render(text.Pad(palette.Name, 14))+theme.Swatch(palette.Name, th.IsDark)+"  "+th.Muted.Render(palette.Description), width)))
	}
	return rows
}
