package epicdetail

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/tinker-works/donsy/netomatic"
	"github.com/tinker-works/goggles/tui/components/comments"
	"github.com/tinker-works/goggles/tui/components/panel"
	"github.com/tinker-works/goggles/tui/components/scroll"
	"github.com/tinker-works/goggles/tui/components/statusline"
	"github.com/tinker-works/goggles/tui/components/text"
	"github.com/tinker-works/goggles/tui/components/transcript"
	"github.com/tinker-works/goggles/tui/components/tree"
	"github.com/tinker-works/goggles/tui/mock"
	"github.com/tinker-works/goggles/tui/theme"
	"github.com/tinker-works/goggles/tui/viewmodel"
	"github.com/tinker-works/goggles/tui/zones"
)

func (m Model) View(th theme.Theme, status statusline.Model, runner *viewmodel.Runner,
	entries []transcript.Entry, width, height int, now time.Time) string {
	if !m.Loaded() {
		return m.loading(th, status, width, now)
	}
	geometry := m.layout(th, status, runner, width, height, now)
	var body string
	if width < 90 {
		body = lipgloss.JoinVertical(lipgloss.Left,
			m.treePanel(th, runner, width, geometry.treeHeight),
			m.sidePanel(th, width, geometry.detailHeight, now))
	} else {
		treeWidth, sideWidth := split(width)
		body = lipgloss.JoinHorizontal(lipgloss.Top,
			m.treePanel(th, runner, treeWidth, geometry.mainHeight), " ",
			m.sidePanel(th, sideWidth, geometry.mainHeight, now))
	}
	sections := []string{m.header(th, status, width, now), body}
	if m.ShowOutput {
		sections = append(sections, m.outputPanel(th, runner, entries, width, geometry.outputHeight))
	}
	return strings.Join(sections, "\n")
}

type screenLayout struct{ mainHeight, treeHeight, detailHeight, outputHeight int }

func (m Model) layout(th theme.Theme, status statusline.Model, runner *viewmodel.Runner,
	width, height int, now time.Time) screenLayout {
	available := max(1, height-lipgloss.Height(m.header(th, status, width, now)))
	treeMin := m.treeMinimum(th, runner, width)
	detailWidth := detailWidth(width)
	detailMin := lipgloss.Height(panel.Render(th, "DETAIL", strings.Join([]string{
		th.Muted.Render("body"), strings.Repeat("─", panel.ContentWidth(detailWidth)),
		m.Comment.View(th, panel.ContentWidth(detailWidth))}, "\n"), detailWidth, 0, false))
	mainMin := max(treeMin, detailMin)
	if width < 90 {
		mainMin = treeMin + detailMin
	}
	output := 0
	if m.ShowOutput {
		output = min(14, max(3, available/3))
		output = min(output, max(0, available-mainMin))
	}
	main := max(1, available-output)
	if main < mainMin {
		main = mainMin
	}
	result := screenLayout{mainHeight: main, outputHeight: output}
	if width < 90 {
		target := max(treeMin, min(main/3, main-detailMin))
		treeHeight := min(m.treeNaturalHeight(th, runner, width), target)
		if treeHeight < treeMin {
			treeHeight = treeMin
		}
		result.treeHeight, result.detailHeight = treeHeight, max(detailMin, main-treeHeight)
	} else {
		result.treeHeight, result.detailHeight = main, main
	}
	return result
}

func (m Model) treeNaturalHeight(th theme.Theme, runner *viewmodel.Runner, width int) int {
	inner := panel.ContentWidth(width)
	top := []string{}
	if m.Epic != nil && viewmodel.Drafting(m.Epic.State) {
		detail := "this epic hasn't been drafted."
		if runner != nil {
			detail = fmt.Sprintf("⚙ %s is drafting it · round %d · %s", runner.Run.Agent,
				draftingRound(m.Epic), statusline.Elapsed(runner.Elapsed))
		}
		top = append(top, th.Muted.Render(text.Truncate(detail, inner)), "")
	}
	rows := append([]string(nil), top...)
	if m.ShowProposal {
		rows = append(rows, m.proposalRows(th, inner)...)
	} else {
		issues := m.Rows()
		rows = append(rows, tree.RenderWindow(th, issues, m.Index, 0, max(1, len(issues)), inner))
	}
	rows = append(rows, "", th.Accent.Render("S")+th.Muted.Render(" transition"))
	return lipgloss.Height(panel.Render(th, "ISSUES", strings.Join(rows, "\n"), width, 0, true))
}

func (m Model) treeMinimum(th theme.Theme, runner *viewmodel.Runner, width int) int {
	inner := panel.ContentWidth(width)
	rows := []string{"", th.Accent.Render("S") + th.Muted.Render(" transition")}
	if m.Epic != nil && viewmodel.Drafting(m.Epic.State) {
		rows = append([]string{th.Muted.Render(text.Truncate("this epic hasn't been drafted.", inner)), ""}, rows...)
	}
	issues := m.Rows()
	rows = append([]string{tree.RenderWindow(th, issues, 0, 0, 1, inner)}, rows...)
	return lipgloss.Height(panel.Render(th, "ISSUES", strings.Join(rows, "\n"), width, 0, true))
}

func split(width int) (treeWidth, sideWidth int) {
	treeWidth = max(28, width*2/5)
	sideWidth = max(28, width-treeWidth-1)
	return treeWidth, sideWidth
}

func detailWidth(width int) int {
	if width < 90 {
		return width
	}
	_, sideWidth := split(width)
	return sideWidth
}

func (m Model) detailBudget(th theme.Theme, runner *viewmodel.Runner, width, height int) int {
	return m.layout(th, statusline.Model{}, runner, width, height, time.Time{}).detailHeight
}

func (m Model) DetailWindow(th theme.Theme, runner *viewmodel.Runner, width, height int) int {
	budget := panel.ContentHeight(m.detailBudget(th, runner, width, height))
	return max(1, budget-m.pinnedRows(th, width))
}

func (m Model) DetailRows(th theme.Theme, width int, now time.Time) int {
	return len(m.detailLines(th, panel.ContentWidth(detailWidth(width)), now))
}

func (m Model) pinnedRows(th theme.Theme, width int) int {
	inner := panel.ContentWidth(detailWidth(width))
	return lipgloss.Height(m.Comment.View(th, inner)) + 1
}

func (m Model) header(th theme.Theme, status statusline.Model, width int, now time.Time) string {
	title, state := "Epic", ""
	if m.Epic != nil {
		title = m.Epic.Title
		state = th.Badge(th.EpicStateStyle(m.Epic.State), m.Epic.State)
	}
	left := th.Title.Render(title) + "  " + state
	if m.Epic != nil && m.Epic.BranchPrefix != "" {
		left += th.Muted.Render("  " + branchNamespace + m.Epic.BranchPrefix + "-…")
	}
	if m.ShowProposal {
		left += th.Selected.Render("  proposed")
	}
	return text.Justify(left, statusline.Render(th, status, now), width)
}

func (m Model) loading(th theme.Theme, status statusline.Model, width int, now time.Time) string {
	message := th.Muted.Render("Loading epic...")
	if status.Offline {
		message = th.Error.Render("Could not load epic — no connection to the daemon") +
			"\n" + th.Muted.Render("r retry")
	}
	return panel.Render(th, "EPIC"+status.TitleSuffix(), message, width, 0, false)
}

func (m Model) treePanel(th theme.Theme, runner *viewmodel.Runner, width, height int) string {
	inner := panel.ContentWidth(width)
	contentHeight := panel.ContentHeight(height)
	top := []string{}
	if m.Epic != nil && viewmodel.Drafting(m.Epic.State) {
		detail := "this epic hasn't been drafted."
		if runner != nil {
			detail = fmt.Sprintf("⚙ %s is drafting it · round %d · %s", runner.Run.Agent,
				draftingRound(m.Epic), statusline.Elapsed(runner.Elapsed))
		}
		top = append(top, th.Muted.Render(text.Truncate(detail, inner)), "")
	}
	bottom := []string{"", th.Accent.Render("S") + th.Muted.Render(" transition")}
	rows := append([]string(nil), top...)
	if m.ShowProposal {
		available := max(0, contentHeight-len(top)-len(bottom))
		rows = append(rows, scroll.Window(m.proposalRows(th, inner), available, 0)...)
	} else {
		issues := m.Rows()
		window := max(1, (contentHeight-len(top)-len(bottom))/2)
		start := scroll.Follow(len(issues), window, m.Index)
		rows = append(rows, tree.RenderWindow(th, issues, m.Index, start, window, inner))
	}
	rows = append(rows, bottom...)
	title := "ISSUES"
	if !m.ShowProposal {
		issues := m.Rows()
		window := max(1, (contentHeight-len(top)-len(bottom))/2)
		title += scroll.Mark(len(issues), window, scroll.Follow(len(issues), window, m.Index))
	}
	return panel.Render(th, title, strings.Join(rows, "\n"), width, height, true)
}

func (m Model) proposalRows(th theme.Theme, width int) []string {
	if m.Epic == nil {
		return []string{th.Muted.Render("Nothing proposed.")}
	}
	changes := mock.Proposal(*m.Epic)
	if len(changes) == 0 {
		return []string{th.Muted.Render("Nothing proposed.")}
	}
	rows := []string{th.Muted.Render("Approving would apply:"), ""}
	for _, change := range changes {
		style := th.Success
		if strings.HasPrefix(change.Marker, "~") {
			style = th.Warning
		}
		rows = append(rows, style.Render(text.Pad(change.Marker, 10))+text.Truncate(change.Title, max(1, width-10)))
	}
	return rows
}

func (m Model) sidePanel(th theme.Theme, width, budget int, now time.Time) string {
	issue, ok := m.SelectedIssue()
	if !ok {
		return panel.Render(th, "DETAIL", th.Muted.Render("Select an issue."), width, budget, false)
	}
	inner := panel.ContentWidth(width)
	lines := m.detailLines(th, inner, now)
	window := max(1, panel.ContentHeight(budget)-m.pinnedRows(th, width))
	offset := min(m.detailScroll, max(0, len(lines)-window))
	rows := append(scroll.Window(lines, window, offset), strings.Repeat("─", inner), m.Comment.View(th, inner))
	title := text.Truncate(issue.Title, max(8, inner-8))
	if pr, ok := m.PullRequestFor(issue.ID); ok && pr.Number > 0 {
		title = fmt.Sprintf("%s · #%d", title, pr.Number)
	}
	focused := !m.ShowOutput || !m.outputFocused
	return panel.Render(th, title+scroll.Mark(len(lines), window, offset),
		zones.Mark(zones.EpicDetailPane, strings.Join(rows, "\n")), width, budget, focused)
}

func (m Model) detailLines(th theme.Theme, inner int, now time.Time) []string {
	issue, ok := m.SelectedIssue()
	if !ok {
		return nil
	}
	meta := []string{}
	if issue.Repository != "" {
		meta = append(meta, issue.Repository)
	}
	meta = append(meta, issue.State)
	if created := parseTime(issue.CreatedAt); !created.IsZero() {
		meta = append(meta, "created "+statusline.Age(created, now)+" ago")
	}
	if pr, ok := m.PullRequestFor(issue.ID); ok && pr.Number > 0 {
		meta = append(meta, fmt.Sprintf("PR #%d", pr.Number))
	}
	lines := []string{th.Muted.Render(text.Truncate(strings.Join(meta, " · "), inner)), strings.Repeat("─", inner)}
	body := m.renderMarkdown(issue.Body, inner)
	if body != "" {
		lines = append(lines, strings.Split(body, "\n")...)
	} else {
		lines = append(lines, th.Muted.Render("No description."))
	}
	lines = append(lines, strings.Repeat("─", inner))
	return append(lines, comments.Render(th, nil, issue.Comments, inner)...)
}

func (m Model) renderMarkdown(source string, width int) string {
	if m.markdown == nil {
		return source
	}
	m.markdown.SetSource(source)
	m.markdown.SetWidth(width)
	return strings.TrimSpace(m.markdown.Render())
}

func (m Model) outputPanel(th theme.Theme, runner *viewmodel.Runner, entries []transcript.Entry, width, height int) string {
	if runner == nil {
		return panel.Render(th, "AGENT OUTPUT", th.Muted.Render("No run is targeting this epic."), width, height, false)
	}
	body := th.Muted.Render("Waiting for the round to write its first output…")
	if len(entries) > 0 {
		rows := transcript.Render(th, entries, panel.ContentWidth(width))
		window := max(1, panel.ContentHeight(height))
		top := scroll.Clamp(len(rows), window, m.outputTop)
		if !m.outputDetached {
			top = scroll.Clamp(len(rows), window, len(rows)-window)
		}
		rows = rows[top:min(len(rows), top+window)]
		body = strings.Join(rows, "\n")
	}
	focused := m.ShowOutput && m.outputFocused
	return panel.Render(th, "AGENT OUTPUT — "+runner.Run.Agent,
		zones.Mark(zones.EpicOutputPane, body), width, height, focused)
}

func (m Model) OutputWindow(th theme.Theme, status statusline.Model, runner *viewmodel.Runner,
	width, height int, now time.Time) int {
	return panel.ContentHeight(m.layout(th, status, runner, width, height, now).outputHeight)
}

func (m Model) Footer(th theme.Theme, width int) string {
	if m.Commenting() {
		return th.Muted.Render("ctrl+s post   esc cancel")
	}
	footer := keyhelpText(th, "j/k move", "pgup/pgdn scroll detail", "c comment", "S transition")
	if m.ShowOutput {
		focused := "detail"
		if m.outputFocused {
			focused = "output"
		}
		footer += "   " + th.Selected.Render("scroll "+focused)
	}
	return text.Truncate(footer, width)
}

func keyhelpText(th theme.Theme, values ...string) string {
	return th.Muted.Render(strings.Join(values, "   "))
}

func draftingRound(epic *netomatic.Epic) int {
	if epic == nil || epic.DraftingPasses <= 0 {
		return 1
	}
	return epic.DraftingPasses
}

func parseTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	parsed, _ := time.Parse(time.RFC3339Nano, value)
	return parsed
}
