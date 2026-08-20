package board

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/tinker-works/donsy/netomatic"
	"github.com/tinker-works/goggles/tui/components/panel"
	"github.com/tinker-works/goggles/tui/components/progress"
	"github.com/tinker-works/goggles/tui/components/scroll"
	"github.com/tinker-works/goggles/tui/components/sparkline"
	"github.com/tinker-works/goggles/tui/components/statusline"
	"github.com/tinker-works/goggles/tui/components/text"
	"github.com/tinker-works/goggles/tui/layout"
	"github.com/tinker-works/goggles/tui/theme"
	"github.com/tinker-works/goggles/tui/viewmodel"
	"github.com/tinker-works/goggles/tui/zones"
)

var draftingRail = []string{"Concept", "Refine", "Review", "Proposed", "Ready"}

func (m Model) View(th theme.Theme, project *netomatic.Project, status statusline.Model, width, height int, now time.Time) string {
	sections := []string{m.header(th, project, status, width)}
	if items := m.Attention(now); len(items) > 0 {
		sections = append(sections, m.attentionStrip(th, items, width))
	}
	budget := m.bodyBudget(sections, width, height)
	switch layoutTier(width) {
	case layout.TierNarrow:
		sections = append(sections, m.narrowBody(th, width, budget, now))
	case layout.TierMedium:
		sections = append(sections, m.listBody(th, width, budget, status), m.railTicker(th, width, now))
	default:
		railWidth := layout.RailWidth(width)
		sections = append(sections, lipgloss.JoinHorizontal(lipgloss.Top,
			m.listBody(th, width-railWidth-1, budget, status), " ", m.rail(th, railWidth, budget, now)))
	}
	joined := strings.Join(sections, "\n")
	if height <= 0 {
		return joined
	}
	return text.FitHeight(joined, height)
}

func (m Model) ZoomView(th theme.Theme, project *netomatic.Project, status statusline.Model, width, height int, now time.Time) string {
	lane, ok := m.ZoomedLane()
	if !ok {
		return m.View(th, project, status, width, height, now)
	}
	sections := []string{m.header(th, project, status, width)}
	budget := m.bodyBudget(sections, width, height)
	switch layoutTier(width) {
	case layout.TierNarrow:
		sections = append(sections, m.zoomBodyNarrow(th, lane, width, budget))
	case layout.TierMedium:
		sections = append(sections, m.zoomBody(th, lane, width, budget), m.railTicker(th, width, now))
	default:
		railWidth := layout.RailWidth(width)
		sections = append(sections, lipgloss.JoinHorizontal(lipgloss.Top,
			m.zoomBody(th, lane, width-railWidth-1, budget), " ", m.rail(th, railWidth, budget, now)))
	}
	joined := strings.Join(sections, "\n")
	if height <= 0 {
		return joined
	}
	return text.FitHeight(joined, height)
}

func layoutTier(width int) layout.Tier {
	return layout.Of(width)
}

func (m Model) bodyBudget(chrome []string, width, height int) int {
	rows := height
	for _, section := range chrome {
		rows -= lipgloss.Height(lipgloss.NewStyle().Width(width).Render(section))
	}
	if layoutTier(width) == layout.TierMedium {
		rows--
	}
	return max(3, rows)
}

func (m Model) header(th theme.Theme, project *netomatic.Project, status statusline.Model, width int) string {
	name := "Project"
	if project != nil && project.Name != "" {
		name = project.Name
	}
	left := zones.Mark(zones.HeaderProject, th.Title.Render(name)) + th.Muted.Render(
		fmt.Sprintf(" · %d epics · %d issues", len(m.Epics), m.IssueCount()))
	if m.GroupBy != viewmodel.GroupByEpic {
		left += th.Accent.Render("  by " + m.GroupBy.String())
	}
	filterValue := m.Filter.Value()
	if m.Filter.Active || filterValue != "" {
		left += th.Selected.Render("  /" + filterValue)
		if m.Filter.Active {
			left += th.Selected.Render("▏")
		}
	}
	return text.Justify(left, status.View(), width)
}

const maxAttentionRows = 3

func (m Model) attentionStrip(th theme.Theme, items []viewmodel.AttentionItem, width int) string {
	if m.AttentionCollapsed {
		return zones.Mark(zones.AttentionRow(0), th.Warning.Render(fmt.Sprintf("⚠ %d needing attention", len(items)))+
			th.Muted.Render("  a expand"))
	}
	inner := panel.ContentWidth(width)
	rows := make([]string, 0, len(items))
	for i, item := range items {
		if i == maxAttentionRows && len(items) > maxAttentionRows {
			rows = append(rows, th.Muted.Render(text.Truncate(fmt.Sprintf("+%d more · a to fold", len(items)-maxAttentionRows), inner)))
			break
		}
		subject := text.Pad(item.Subject, min(20, max(8, inner/3)))
		row := th.Warning.Render(subject) + " " + th.Muted.Render(text.Truncate(item.Detail, max(1, inner-lipgloss.Width(subject)-1)))
		rows = append(rows, zones.Mark(zones.AttentionRow(i), row))
	}
	return panel.Render(th, "⚠ Needs attention", strings.Join(rows, "\n"), width, 0, false)
}

func (m Model) listBody(th theme.Theme, width, budget int, status statusline.Model) string {
	if len(m.lanes) == 0 {
		if status.IsLoading() {
			return panel.Render(th, "BOARD", m.skeleton(th, width), width, budget, false)
		}
		return panel.Render(th, "BOARD", emptyBoard(th), width, budget, false)
	}
	inner := panel.ContentWidth(width)
	rows := make([]string, 0, len(m.lanes))
	for i, lane := range m.lanes {
		rows = append(rows, m.laneRow(th, i, lane, inner))
	}
	window := max(1, (panel.ContentHeight(budget)+1)/2)
	offset := scroll.Follow(len(rows), window, m.Lane)
	body := strings.Join(scroll.Window(rows, window, offset), "\n"+strings.Repeat("─", inner)+"\n")
	return panel.Render(th, "BOARD"+scroll.Mark(len(rows), window, offset), body, width, budget, false)
}

func emptyBoard(th theme.Theme) string {
	return th.Muted.Render("No epics yet.") + "\n\n" + th.Accent.Render("N") +
		th.Muted.Render("   start one — write a rough note, agents draft the tree")
}

func (m Model) laneRow(th theme.Theme, index int, lane viewmodel.Lane, width int) string {
	gear := " "
	if m.hasRunner(lane) {
		gear = th.Accent.Render("⚙")
	}
	tail := fmt.Sprintf("%s %s  %s %s", text.Pad(th.Badge(th.EpicStateStyle(lane.State), lane.State), 18), gear,
		text.Pad(laneCounts(lane), 36), progress.Bar(lane.Merged, lane.Total, progress.DefaultCells))
	head := "▸ " + text.Truncate(lane.Title, max(10, width-lipgloss.Width(tail)-3))
	if !lane.Matched {
		head = th.Subtle.Render(head)
	} else if m.Lane == index {
		head = th.Selected.Render(head)
	}
	return zones.Mark(zones.BoardEpicCard(index), text.Truncate(text.Justify(head, tail, width), width))
}

func laneCounts(lane viewmodel.Lane) string {
	active := lane.Count(viewmodel.CodingColumn) + lane.Count(viewmodel.ReviewColumn)
	return fmt.Sprintf("Open %d · Active %d · PR %d · Merged %d", lane.Count(viewmodel.OpenColumn), active,
		lane.Count(viewmodel.PRColumn), lane.Count(viewmodel.MergedColumn))
}

func (m Model) zoomBody(th theme.Theme, lane viewmodel.Lane, width, budget int) string {
	inner := panel.ContentWidth(width)
	rows := append([]string(nil), m.zoomHeaderRows(th, lane, inner)...)
	rows = append(rows, strings.Repeat("─", inner))
	if lane.Drafting {
		rows = append(rows, m.draftingBody(th, lane, inner))
		return panel.Render(th, "", strings.Join(rows, "\n"), width, budget, false)
	}
	avail := panel.ContentHeight(budget) - len(rows) - 1
	widths := zoomColumnWidths(inner, int(viewmodel.LastColumn-viewmodel.OpenColumn)+1)
	rows = append(rows, m.zoomColumnHeaders(th, widths), m.zoomColumns(th, lane, widths, avail))
	return panel.Render(th, "", strings.Join(rows, "\n"), width, budget, false)
}

func (m Model) zoomHeaderRows(th theme.Theme, lane viewmodel.Lane, inner int) []string {
	title := th.Title.Render(text.Truncate(lane.Title, max(1, inner-20)))
	head := title + " " + th.Badge(th.EpicStateStyle(lane.State), lane.State)
	if m.hasRunner(lane) {
		head += " " + th.Accent.Render("⚙")
	}
	tally := th.Muted.Render(laneCounts(lane)) + "  " + progress.Bar(lane.Merged, lane.Total, progress.DefaultCells)
	if runner, ok := m.runnerFor(lane); ok && runner.Live() {
		tally += th.Muted.Render(fmt.Sprintf("  ·  ⚙ %s %s", runner.Run.Agent, statusline.Elapsed(runner.Elapsed)))
	}
	return []string{text.Truncate(head, inner), text.Truncate(tally, inner)}
}

func zoomColumnWidths(inner, columns int) []int {
	avail := max(columns, inner-(columns-1))
	base, rest := avail/columns, avail%columns
	widths := make([]int, columns)
	for i := range widths {
		widths[i] = base
		if i < rest {
			widths[i]++
		}
	}
	return widths
}

func (m Model) zoomColumnHeaders(th theme.Theme, widths []int) string {
	cells := make([]string, 0, len(widths))
	for i, title := range viewmodel.ColumnTitles[1:] {
		cells = append(cells, text.Pad(th.Accent.Render(title), widths[i]))
	}
	return strings.Join(cells, " ")
}

func (m Model) zoomColumns(th theme.Theme, lane viewmodel.Lane, widths []int, rows int) string {
	cells := make([]string, 0, len(widths)*2)
	for column := viewmodel.OpenColumn; column <= viewmodel.MergedColumn; column++ {
		if column > viewmodel.OpenColumn {
			cells = append(cells, " ")
		}
		cells = append(cells, m.issueColumn(th, m.Lane, lane, column, widths[int(column-viewmodel.OpenColumn)], rows))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cells...)
}

func (m Model) zoomBodyNarrow(th theme.Theme, lane viewmodel.Lane, width, budget int) string {
	inner := panel.ContentWidth(width)
	head := append([]string(nil), m.zoomHeaderRows(th, lane, inner)...)
	head = append(head, strings.Repeat("─", inner))
	if lane.Drafting {
		head = append(head, m.draftingBody(th, lane, inner))
		return panel.Render(th, "", strings.Join(head, "\n"), width, budget, false)
	}
	body, focused := []string{}, 0
	for column := viewmodel.OpenColumn; column <= viewmodel.MergedColumn; column++ {
		issues := lane.Issues(column)
		if len(issues) == 0 {
			continue
		}
		body = append(body, th.Accent.Render(viewmodel.ColumnTitles[int(column)]))
		for i, issue := range issues {
			marker, style := " ", th.Muted
			if m.Column == column && m.Index == i {
				marker, style, focused = "›", th.Selected, len(body)
			}
			row := style.Render(marker+" "+text.Truncate(issue.Issue.Title, max(1, inner-8))) + th.Muted.Render(issueMeta(issue, 0))
			body = append(body, zones.Mark(zones.BoardIssueCard(m.Lane, int(column), i), text.Truncate(row, inner)))
		}
	}
	window := max(1, panel.ContentHeight(budget)-len(head))
	offset := scroll.Follow(len(body), window, focused)
	return panel.Render(th, scroll.Mark(len(body), window, offset), strings.Join(append(head, scroll.Window(body, window, offset)...), "\n"), width, budget, false)
}

func (m Model) draftingBody(th theme.Theme, lane viewmodel.Lane, inner int) string {
	labels := make([]string, 0, len(draftingRail))
	marker := strings.Builder{}
	for _, state := range draftingRail {
		if state == lane.State {
			labels = append(labels, th.EpicStateStyle(state).Render(state))
			marker.WriteString(th.Accent.Render(strings.Repeat("━", max(1, len(state)-1)) + "●"))
		} else {
			labels = append(labels, th.Muted.Render(state))
			marker.WriteString(strings.Repeat(" ", len(state)))
		}
		marker.WriteString("   ")
	}
	rows := []string{text.Truncate(strings.Join(labels, th.Muted.Render(" ──▶ ")), inner), text.Truncate(marker.String(), inner)}
	if runner, ok := m.runnerFor(lane); ok {
		rows = append(rows, th.Muted.Render(text.Truncate(fmt.Sprintf("⚙ %s · %s", runner.Run.Agent, statusline.Elapsed(runner.Elapsed)), inner)))
	} else {
		rows = append(rows, th.Muted.Render("no runner on this epic yet"))
	}
	return strings.Join(append(rows, th.Accent.Render("⏎")+th.Muted.Render(" read the draft   ")+th.Accent.Render("A")+th.Muted.Render(" approve   ")+th.Accent.Render("X")+th.Muted.Render(" discard")), "\n")
}

func (m Model) hasRunner(lane viewmodel.Lane) bool {
	for _, runner := range m.Runners {
		if !runner.Live() {
			continue
		}
		if lane.Epic != nil && runner.Run.Project == lane.Epic.ID {
			return true
		}
		if lane.Epic == nil && runner.Subject == lane.Title {
			return true
		}
	}
	return false
}

func (m Model) runnerFor(lane viewmodel.Lane) (viewmodel.Runner, bool) {
	for _, runner := range m.Runners {
		if lane.Epic != nil && runner.Run.Project == lane.Epic.ID {
			return runner, true
		}
	}
	return viewmodel.Runner{}, false
}

const cardRows = 4

func (m Model) issueColumn(th theme.Theme, laneIndex int, lane viewmodel.Lane, column viewmodel.Column, width, rows int) string {
	issues := lane.Issues(column)
	if len(issues) == 0 {
		return text.Fit(th.Muted.Render("  —"), width)
	}
	if rows < cardRows {
		return text.Fit(th.Muted.Render(fmt.Sprintf(" %d", len(issues))), width)
	}
	cards := make([]string, 0, len(issues))
	for i, issue := range issues {
		focused := m.Lane == laneIndex && m.Column == column && m.Index == i
		title := text.Truncate(issue.Issue.Title, panel.ContentWidth(width))
		meta := issueMeta(issue, panel.ContentWidth(width))
		if !issue.Matched {
			title, meta = th.Subtle.Render(title), th.Subtle.Render(meta)
		} else {
			meta = th.Muted.Render(meta)
		}
		card := panel.Render(th, "", title+"\n"+meta, width, 0, focused)
		cards = append(cards, zones.Mark(zones.BoardIssueCard(laneIndex, int(column), i), card))
	}
	window := columnWindow(len(cards), rows)
	offset := 0
	if m.Lane == laneIndex && m.Column == column {
		offset = scroll.Follow(len(cards), window, m.Index)
	}
	shown := scroll.Window(cards, window, offset)
	if hidden := len(cards) - window; hidden > 0 {
		shown = append(shown, text.Pad(th.Muted.Render(fmt.Sprintf("%s %d more", strings.TrimSpace(scroll.Mark(len(cards), window, offset)), hidden)), width))
	}
	return strings.Join(shown, "\n")
}

func columnWindow(cards, rows int) int {
	if fit := rows / cardRows; fit >= cards {
		return max(1, fit)
	}
	return max(1, (rows-1)/cardRows)
}

func issueMeta(issue viewmodel.BoardIssue, inner int) string {
	repository := issue.Issue.Repository
	if repository == "" {
		repository = "root"
	}
	if issue.Number <= 0 {
		if inner == 0 {
			return "  " + repository
		}
		return repository
	}
	number := fmt.Sprintf("#%d", issue.Number)
	if inner == 0 {
		return "  " + repository + " " + number
	}
	return text.Pad(repository, max(1, inner-len(number))) + number
}

func (m Model) narrowBody(th theme.Theme, width, budget int, now time.Time) string {
	if len(m.Epics) == 0 {
		return panel.Render(th, "BOARD", emptyBoard(th), width, budget, false)
	}
	inner := panel.ContentWidth(width)
	blocks := []string{}
	for i, lane := range m.lanes {
		marker, style := "▸", th.Muted
		if m.Lane == i {
			marker, style = "›", th.Selected
		}
		gear := ""
		if m.hasRunner(lane) {
			gear = th.Accent.Render("  ⚙")
		}
		head := style.Render(marker+" "+text.Pad(lane.Title, min(22, max(1, inner-14)))) + th.Badge(th.EpicStateStyle(lane.State), lane.State) + gear
		detail := th.Muted.Render("    "+laneCounts(lane)) + "        " + progress.Bar(lane.Merged, lane.Total, progress.DefaultCells)
		blocks = append(blocks, zones.Mark(zones.BoardEpicCard(i), text.Truncate(head, inner))+"\n"+text.Truncate(detail, inner))
	}
	avail := panel.ContentHeight(budget)
	ticker := m.tickerLine(th, inner, now)
	if ticker != "" {
		avail -= 2
	}
	window := max(1, avail/2)
	offset := scroll.Follow(len(blocks), window, m.Lane)
	rows := scroll.Window(blocks, window, offset)
	if ticker != "" {
		rows = append(rows, strings.Repeat("─", inner), ticker)
	}
	return panel.Render(th, scroll.Mark(len(blocks), window, offset), strings.Join(rows, "\n"), width, budget, false)
}

func (m Model) rail(th theme.Theme, width, budget int, now time.Time) string {
	inner := panel.ContentWidth(width)
	avail := panel.ContentHeight(budget)
	sandbox := m.sandboxRows(th, inner, avail-4)
	if len(m.Runners) == 0 {
		return panel.Render(th, "RUNNERS", strings.Join(append(sandbox, th.Muted.Render("idle — no runners")), "\n"), width, budget, false)
	}
	blocks := make([]string, 0, len(m.Runners))
	for i, runner := range m.Runners {
		blocks = append(blocks, zones.Mark(zones.BoardRunnerRow(i), m.runnerBlock(th, runner, inner)))
	}
	rows := m.railRows(th, blocks, avail-len(sandbox), inner)
	return panel.Render(th, "RUNNERS", strings.Join(append(sandbox, rows...), "\n"), width, budget, false)
}

func (m Model) railRows(th theme.Theme, blocks []string, avail, inner int) []string {
	if avail < 4 {
		return []string{th.Muted.Render(text.Truncate(fmt.Sprintf("%d runners · s for all", len(blocks)), inner))}
	}
	window := max(1, avail/4)
	rows := strings.Split(strings.Join(scroll.Window(blocks, window, 0), "\n\n"), "\n")
	if hidden := len(blocks) - window; hidden > 0 {
		rows = append(rows, th.Muted.Render(text.Truncate(fmt.Sprintf("+%d more · s for all runs", hidden), inner)))
	}
	return rows
}

func (m Model) runnerBlock(th theme.Theme, runner viewmodel.Runner, inner int) string {
	status := th.AgentRunStatusStyle(runner.Run.Status).Render(theme.RunDot(runner.Run.Status) + " " + runner.Run.Agent)
	tail := statusline.Elapsed(runner.Elapsed)
	if !runner.Live() {
		tail += "  " + theme.RunStatusLabel(runner.Run.Status)
	} else if samples := m.ActivityFor(runner.Run.ID); len(samples) > 0 {
		values := make([]float64, len(samples))
		for i, sample := range samples {
			values[i] = float64(sample)
		}
		line := sparkline.New(railActivityCells)
		line.SetValues(values)
		tail += "  " + line.View()
	}
	return strings.Join([]string{status, th.Muted.Render(text.Truncate("  "+runner.Subject, inner)), th.Muted.Render("  " + tail)}, "\n")
}

func (m Model) sandboxRows(th theme.Theme, inner, avail int) []string {
	if len(m.Sandboxes) == 0 || avail < 3 {
		return nil
	}
	dots := strings.Split(m.sandboxDots(th, inner), "\n")
	if len(dots) > avail-2 {
		dots = dots[:max(1, avail-2)]
	}
	return append(append([]string{th.Accent.Render("Sandboxes")}, dots...), strings.Repeat("─", inner))
}

func (m Model) sandboxDots(th theme.Theme, width int) string {
	lines, current := []string{}, ""
	for _, sandbox := range m.Sandboxes {
		dot := th.SandboxDot(sandbox.Status) + " "
		if lipgloss.Width(current)+2 > width {
			lines = append(lines, current)
			current = ""
		}
		current += dot
	}
	return strings.Join(append(lines, current), "\n")
}

func (m Model) railTicker(th theme.Theme, width int, now time.Time) string {
	if line := m.tickerLine(th, width, now); line != "" {
		return line
	}
	return th.Muted.Render("runners  idle")
}

func (m Model) tickerLine(th theme.Theme, width int, _ time.Time) string {
	if len(m.Runners) == 0 {
		return ""
	}
	parts := make([]string, 0, len(m.Runners))
	for _, runner := range m.Runners {
		parts = append(parts, th.AgentRunStatusStyle(runner.Run.Status).Render(theme.RunDot(runner.Run.Status))+" "+runner.Run.Agent+" "+th.Muted.Render(runner.Subject+" "+statusline.Elapsed(runner.Elapsed)))
	}
	return text.Truncate(th.Accent.Render(fmt.Sprintf("runners %d live  ", viewmodel.LiveCount(m.Runners)))+strings.Join(parts, th.Muted.Render(" · ")), width)
}

func (m Model) skeleton(th theme.Theme, width int) string {
	inner := panel.ContentWidth(width)
	rows := []string{}
	for i := 0; i < 3; i++ {
		rows = append(rows, th.Subtle.Render(strings.Repeat("▁", min(inner, 18+i*4))), th.Subtle.Render(strings.Repeat("▁", min(inner, 10))), "")
	}
	return strings.Join(rows, "\n")
}

func (m Model) Footer(th theme.Theme, width int) string {
	if m.Filter.Active || m.Filter.Value() != "" {
		return th.Muted.Render("filter: " + m.Filter.Value() + "   enter apply   esc clear")
	}
	return th.Muted.Render(text.Truncate("j/k move   enter open   g group   / filter   r reload   ? keys", width))
}

func (m Model) ZoomFooter(th theme.Theme, width int) string {
	if m.Filter.Active || m.Filter.Value() != "" {
		return th.Muted.Render("filter: " + m.Filter.Value() + "   enter apply   esc clear")
	}
	return th.Muted.Render(text.Truncate("j/k move   h/l column   J/K epic   esc back   ? keys", width))
}
