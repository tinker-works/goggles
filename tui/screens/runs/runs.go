// Package runs implements the agent-run list and detail transcript reader.
package runs

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/tinker-works/donsy/netomatic"
	"github.com/tinker-works/goggles/tui/components/filter"
	"github.com/tinker-works/goggles/tui/components/panel"
	"github.com/tinker-works/goggles/tui/components/statusline"
	"github.com/tinker-works/goggles/tui/components/text"
	"github.com/tinker-works/goggles/tui/components/transcript"
	"github.com/tinker-works/goggles/tui/theme"
	"github.com/tinker-works/goggles/tui/viewmodel"
	"github.com/tinker-works/goggles/tui/zones"
)

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
	Runners []viewmodel.Runner
	Index   int
	Filter  filterState
	Project string

	output      []transcript.Entry
	outputRunID string
	offset      int64
	outputErr   error
	top         int
	detached    bool
	samples     []int
}

const activityCells = 24

func New() Model                           { return Model{} }
func (m Model) Output() []transcript.Entry { return m.output }
func (m Model) OutputOffset() int64        { return m.offset }
func (m Model) OutputRunID() string        { return m.outputRunID }
func (m Model) OutputTop() int             { return m.top }
func (m Model) FollowingOutput() bool      { return !m.detached }

func (m Model) TrackOutput(runID string) Model {
	if m.outputRunID != runID {
		m.output, m.offset, m.outputErr = nil, 0, nil
		m.top, m.detached, m.samples = 0, false, nil
		m.outputRunID = runID
	}
	return m
}

func (m Model) AppendOutput(runID string, entries []transcript.Entry, next int64, err error) Model {
	if runID != m.outputRunID {
		return m
	}
	m.outputErr = err
	if err != nil {
		return m
	}
	m.samples = append(m.samples, len(entries))
	if len(m.samples) > activityCells {
		m.samples = m.samples[len(m.samples)-activityCells:]
	}
	for _, entry := range entries {
		if at, ok := m.indexOfCall(entry); ok {
			m.output[at] = entry
		} else {
			m.output = append(m.output, entry)
		}
	}
	m.offset = next
	return m
}

func (m Model) indexOfCall(entry transcript.Entry) (int, bool) {
	if entry.CallID == "" {
		return 0, false
	}
	for i := range m.output {
		if m.output[i].CallID == entry.CallID && m.output[i].Kind == entry.Kind {
			return i, true
		}
	}
	return 0, false
}

func (m Model) ReloadOutput() Model {
	m.output, m.offset, m.outputErr = nil, 0, nil
	m.top, m.detached, m.samples = 0, false, nil
	return m
}

func (m Model) ScrollOutput(rows, total, window int) Model {
	end := max(0, total-window)
	next := max(0, min(m.logTop(total, window)+rows, end))
	m.top, m.detached = next, next < end
	return m
}

func (m Model) logTop(total, window int) int {
	end := max(0, total-window)
	if !m.detached {
		return end
	}
	return min(m.top, end)
}

func (m Model) SetRunners(runners []viewmodel.Runner) Model {
	selectedID := ""
	if selected, ok := m.Selected(); ok {
		selectedID = selected.Run.ID
	}
	m.Runners = runners
	visible := m.Visible()
	if selectedID != "" {
		for i := range visible {
			if visible[i].Run.ID == selectedID {
				m.Index = i
				return m
			}
		}
	}
	m.Index = max(0, min(m.Index, max(0, len(visible)-1)))
	return m
}

func (m Model) Visible() []viewmodel.Runner {
	kept := make([]viewmodel.Runner, 0, len(m.Runners))
	for _, runner := range m.Runners {
		if filter.Matches(m.Filter.Value, runner.Run.Agent, runner.Subject, theme.RunStatusLabel(runner.Run.Status)) {
			kept = append(kept, runner)
		}
	}
	return kept
}

func (m Model) Selected() (viewmodel.Runner, bool) {
	visible := m.Visible()
	if m.Index < 0 || m.Index >= len(visible) {
		return viewmodel.Runner{}, false
	}
	return visible[m.Index], true
}

func (m Model) MoveUp() Model   { m.Index = max(0, m.Index-1); return m }
func (m Model) MoveDown() Model { m.Index = min(max(0, len(m.Visible())-1), m.Index+1); return m }

func (m Model) SelectRow(index int) Model {
	if index >= 0 && index < len(m.Visible()) {
		m.Index = index
	}
	return m
}

func (m Model) StartFilter() Model { m.Filter = m.Filter.Start(); return m }

func (m Model) ApplyFilterKey(msg tea.KeyPressMsg) Model {
	m.Filter = applyFilterKey(m.Filter, msg)
	m.Index = max(0, min(m.Index, max(0, len(m.Visible())-1)))
	if !m.Filter.Active && m.Filter.Value == "" {
		m.Index = 0
	}
	return m
}

func (m Model) SetFilter(value string) Model {
	m.Filter = m.Filter.Set(value)
	m.Index = max(0, min(m.Index, max(0, len(m.Visible())-1)))
	return m
}

func (m Model) View(th theme.Theme, status statusline.Model, width, height int) string {
	inner := panel.ContentWidth(width)
	visible := m.Visible()
	columns, tokens := m.columns(inner)
	hint := "f filter   r refresh"
	if m.Filter.Active {
		hint = "/" + m.Filter.Value + "▏"
	}
	rows := []string{text.Justify(th.Muted.Render(m.summary(visible, tokens)), th.Accent.Render(hint), inner), strings.Repeat("─", inner)}
	rows = append(rows, m.renderTable(th, columns, visible, inner))
	if len(visible) == 0 {
		empty := "No runs yet."
		if m.Filter.Value != "" {
			empty = "No runs match " + m.Filter.Value
		}
		rows = append(rows, "", th.Muted.Render(empty))
	}
	return panel.Render(th, m.panelTitle()+status.TitleSuffix(), strings.Join(rows, "\n"), width, height, false)
}

type column struct {
	title string
	width int
}

func (m Model) columns(width int) ([]column, bool) {
	wide := []column{{"ROLE", 10}, {"SUBJECT", 16}, {"STATUS", 11}, {"VERDICT", 9}, {"IN", 7}, {"OUT", 7}, {"COST", 7}, {"AGE", 7}}
	if fits(wide, width) {
		return widen(wide, width), true
	}
	narrow := []column{{"ROLE", 10}, {"SUBJECT", 16}, {"STATUS", 11}, {"VERDICT", 8}, {"COST", 7}, {"AGE", 7}}
	return widen(narrow, width), false
}

func fits(columns []column, width int) bool {
	used := 0
	for _, column := range columns {
		used += column.width + 1
	}
	return width-used-4 >= 0
}

func widen(columns []column, width int) []column {
	used := 0
	for _, column := range columns {
		used += column.width + 1
	}
	columns[1].width += max(0, width-used-4)
	return columns
}

func (m Model) renderTable(th theme.Theme, columns []column, runners []viewmodel.Runner, width int) string {
	line := make([]string, len(columns))
	for i, column := range columns {
		line[i] = text.Pad(column.title, column.width)
	}
	rows := []string{strings.TrimRight(strings.Join(line, " "), " ")}
	for i, runner := range runners {
		in, out, cost := usageCells(runner.Run)
		values := []string{runner.Run.Agent, runner.Subject, th.RunBadge(runner.Run.Status), runner.Verdict()}
		if len(columns) == 8 {
			values = append(values, in, out)
		}
		values = append(values, cost, statusline.Elapsed(runner.Elapsed))
		cells := make([]string, len(values))
		for j, value := range values {
			cells[j] = text.Pad(value, columns[j].width)
		}
		row := strings.TrimRight(strings.Join(cells, " "), " ")
		if i == m.Index {
			row = th.Selected.Render(row)
		}
		rows = append(rows, zones.Mark(zones.RunRow(i), text.Truncate(row, width)))
	}
	return strings.Join(rows, "\n")
}

func (m Model) Footer(th theme.Theme, width int) string {
	return text.Truncate(th.Muted.Render("j/k move   enter open   f filter   r reload   esc back"), width)
}

func (m Model) DetailFooter(th theme.Theme, width int) string {
	footer := "j/k scroll   K kill run   r reload   esc runs"
	if runner, ok := m.Selected(); ok && isTerminal(runner.Run.Status) {
		footer = "j/k scroll   r reload   esc runs"
	}
	return text.Truncate(th.Muted.Render(footer), width)
}

func (m Model) summary(visible []viewmodel.Runner, tokens bool) string {
	count := fmt.Sprintf("%d total", len(visible))
	if len(visible) != len(m.Runners) {
		count = fmt.Sprintf("%d of %d", len(visible), len(m.Runners))
	}
	in, out, cost := usageCellsFrom(viewmodel.TotalUsage(visible))
	parts := []string{count, fmt.Sprintf("%d running", viewmodel.LiveCount(visible))}
	if tokens {
		parts = append(parts, "in "+in, "out "+out)
	}
	return strings.Join(append(parts, cost), " · ")
}

func (m Model) panelTitle() string {
	if m.Project == "" {
		return "Runs"
	}
	return "Runs — " + m.Project
}

func (m Model) DetailView(th theme.Theme, status statusline.Model, runner viewmodel.Runner, width, height int) string {
	inner := panel.ContentWidth(width)
	run := runner.Run
	statusRow := th.RunBadge(run.Status) + th.Muted.Render("   started "+statusline.Elapsed(runner.Elapsed)+" ago   "+run.Agent+":"+run.Variant)
	usageLine := th.Muted.Render(usageText(run, runner.Branch))
	head := []string{statusRow, strings.Repeat("─", inner), usageLine,
		th.Muted.Render("activity  ") + m.activityLine(th, run), strings.Repeat("─", inner)}
	var tail []string
	if run.Error != "" {
		tail = append(tail, strings.Repeat("─", inner), th.Error.Render(text.Truncate("! "+run.Error, inner)))
	}
	log := m.outputWindow(th, run, inner, m.LogWindow(height))
	rows := append(append(append([]string{}, head...), log...), tail...)
	title := detailTitle(runner)
	logRows := transcript.Render(th, m.output, inner)
	if !isTerminal(run.Status) && !m.detached && len(logRows) > 0 {
		logRows = append(logRows, th.Accent.Render("_"))
	}
	if total := len(logRows); total > 0 {
		window := max(1, m.LogWindow(height))
		title += scrollMark(total, window, m.logTop(total, window))
	}
	return panel.Render(th, title+status.TitleSuffix(), strings.Join(rows, "\n"), width, height, false)
}

func (m Model) LogWindow(height int) int {
	return max(1, panel.ContentHeight(height)-chromeRows(m.selectedRun()))
}
func (m Model) LogRows(width int) int {
	return len(transcript.Lines(m.output, panel.ContentWidth(width)))
}

func (m Model) selectedRun() netomatic.AgentRun {
	if runner, ok := m.Selected(); ok {
		return runner.Run
	}
	return netomatic.AgentRun{}
}

func chromeRows(run netomatic.AgentRun) int {
	rows := 5
	if run.Error != "" {
		rows += 2
	}
	return rows
}

func (m Model) outputWindow(th theme.Theme, run netomatic.AgentRun, width, window int) []string {
	if m.outputErr != nil {
		return []string{th.Error.Render(text.Truncate("! "+m.outputErr.Error(), width))}
	}
	if len(m.output) == 0 {
		if isTerminal(run.Status) {
			return []string{th.Muted.Render("No transcript was recorded for this run.")}
		}
		return []string{th.Muted.Render("Waiting for the round to write its first output…")}
	}
	rows := transcript.Render(th, m.output, width)
	if !isTerminal(run.Status) && !m.detached {
		rows = append(rows, th.Accent.Render("_"))
	}
	return windowOf(rows, window, m.logTop(len(rows), window))
}

func windowOf(rows []string, window, top int) []string {
	if window <= 0 {
		return nil
	}
	if len(rows) < window {
		padded := make([]string, window)
		copy(padded, rows)
		return padded
	}
	top = max(0, min(top, len(rows)-window))
	return rows[top : top+window]
}

func detailTitle(runner viewmodel.Runner) string {
	return fmt.Sprintf("%s · %s", runner.Run.Agent, runner.Subject)
}

func isTerminal(status string) bool {
	switch normalize(status) {
	case "queued", "admitted", "running":
		return false
	default:
		return true
	}
}

func compact(value int64) string {
	switch {
	case value < 1000:
		return fmt.Sprintf("%d", value)
	case value < 1_000_000:
		return fmt.Sprintf("%.1fk", float64(value)/1000)
	default:
		return fmt.Sprintf("%.1fM", float64(value)/1_000_000)
	}
}

func usageCells(run netomatic.AgentRun) (string, string, string) {
	if run.InputTokens == 0 && run.OutputTokens == 0 {
		return "—", "—", "—"
	}
	return compact(run.InputTokens), compact(run.OutputTokens), "—"
}

func usageCellsFrom(usage viewmodel.Usage) (string, string, string) {
	if !usage.Reported() {
		return "—", "—", "—"
	}
	return compact(usage.TokensIn), compact(usage.TokensOut), "—"
}

func usageText(run netomatic.AgentRun, branch string) string {
	in, out, cost := usageCells(run)
	value := "tokens in " + in + "   out " + out + "   cost " + cost
	if branch != "" {
		value += "   branch " + branch
	}
	return value
}

func (m Model) activityLine(th theme.Theme, run netomatic.AgentRun) string {
	if run.ID != m.outputRunID || len(m.samples) == 0 {
		return th.Muted.Render("(no live trace)")
	}
	levels := []rune("▁▂▃▄▅▆▇█")
	values := m.samples
	if len(values) > activityCells {
		values = values[len(values)-activityCells:]
	}
	maxValue := 1
	for _, value := range values {
		maxValue = max(maxValue, value)
	}
	result := make([]rune, len(values))
	for i, value := range values {
		result[i] = levels[min(len(levels)-1, value*len(levels)/maxValue)]
	}
	return th.Accent.Render(string(result))
}

func scrollMark(total, window, top int) string {
	end := max(0, total-window)
	if end == 0 {
		return ""
	}
	if top == 0 {
		return " ↓"
	}
	if top >= end {
		return " ↑"
	}
	return " ↕"
}

func normalize(value string) string {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(value)), "_", "-")
}

// Elapsed re-exports the shared run-duration formatter for screen callers.
func Elapsed(duration time.Duration) string { return statusline.Elapsed(duration) }
