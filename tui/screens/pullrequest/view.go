package pullrequest

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/tinker-works/donsy/netomatic"
	"github.com/tinker-works/goggles/tui/components/comments"
	"github.com/tinker-works/goggles/tui/components/panel"
	"github.com/tinker-works/goggles/tui/components/scroll"
	"github.com/tinker-works/goggles/tui/components/statusline"
	"github.com/tinker-works/goggles/tui/components/tabs"
	"github.com/tinker-works/goggles/tui/components/text"
	"github.com/tinker-works/goggles/tui/theme"
	"github.com/tinker-works/goggles/tui/zones"
)

func (m Model) View(th theme.Theme, status statusline.Model, width, height int) string {
	if !m.Loaded() {
		return panel.Render(th, "PULL REQUEST"+status.TitleSuffix(), th.Muted.Render("No pull request selected."), width, 0, false)
	}
	inner := panel.ContentWidth(width)
	head, foot := m.chrome(th, inner)
	rows := append([]string(nil), head...)
	title := m.title()
	if m.Tab == Conversation {
		lines := m.conversationLines(th, inner)
		window := window(th, height, head, foot)
		offset := min(m.scroll, max(0, len(lines)-window))
		rows = append(rows, zones.Mark(zones.ConversationPane, strings.Join(scroll.Window(lines, window, offset), "\n")))
		title += scroll.Mark(len(lines), window, offset)
	} else {
		body := window(th, height, head, foot)
		windows := diffWindows(width, body, len(m.Files()))
		rows = append(rows, m.diff(th, inner, width, body))
		if files := m.Files(); len(files) > windows.tree {
			title += scroll.Mark(len(files), windows.tree, scroll.Follow(len(files), windows.tree, m.FileIndex))
		}
		if total := m.DiffRows(th, width); total > windows.code {
			title += scroll.Mark(total, windows.code, scroll.Clamp(total, windows.code, m.diffScroll))
		}
	}
	rows = append(rows, foot...)
	return panel.Render(th, title+status.TitleSuffix(), strings.Join(rows, "\n"), width, height, false)
}

func (m Model) chrome(th theme.Theme, inner int) (head, foot []string) {
	rule := strings.Repeat("─", inner)
	head = []string{m.headerLine(th, inner)}
	if m.Stale() {
		head = append(head, th.Warning.Render("⚠ Review is stale: the branch moved since the last verdict"))
	}
	head = append(head, rule, tabs.Render(th, TabLabels, int(m.Tab)), rule)
	if m.Tab == Conversation {
		foot = append(foot, rule, m.Comment.View(th, inner))
	}
	return head, append(foot, rule, m.actions(th, inner))
}

func (m Model) Footer(th theme.Theme, width int) string {
	footer := "d conversation/diff   c comment   R retry   V request review   esc back"
	if m.Tab == Diff {
		footer = "d conversation/diff   v inline/two-column   f filter files   esc back"
	}
	return text.Truncate(th.Muted.Render(footer), width)
}

func window(_ theme.Theme, height int, head, foot []string) int {
	used := 0
	for _, section := range append(append([]string(nil), head...), foot...) {
		used += lipgloss.Height(section)
	}
	return max(1, panel.ContentHeight(height)-used)
}

func (m Model) ConversationRows(th theme.Theme, width int) int {
	return len(m.conversationLines(th, panel.ContentWidth(width)))
}

func (m Model) ConversationWindow(th theme.Theme, width, height int) int {
	head, foot := m.chrome(th, panel.ContentWidth(width))
	return window(th, height, head, foot)
}

func (m *Model) DiffWindow(th theme.Theme, width, height int) int {
	head, foot := m.chrome(th, panel.ContentWidth(width))
	result := diffWindows(width, window(th, height, head, foot), len(m.Files()))
	m.diffWindow = result.code
	m.diffTotal = m.DiffRows(th, width)
	m.clampDiffScroll(result.code)
	return result.code
}

type diffWindowsResult struct{ tree, code int }

func diffWindows(width, body, treeRows int) (result diffWindowsResult) {
	result.tree, result.code = max(1, body), max(1, body)
	if layoutTierNarrow(width) {
		result.tree = min(max(1, body/3), max(1, treeRows))
		result.code = max(1, body-result.tree-1)
	}
	return result
}

func layoutTierNarrow(width int) bool { return width < 90 }

func (m Model) DiffRows(_ theme.Theme, _ int) int { return m.renderedDiffRows() }

func (m Model) title() string {
	title := m.PullRequest.Title
	if m.Issue != nil {
		title = m.Issue.Title
	}
	if m.PullRequest.Repository != "" && m.PullRequest.Number > 0 {
		return fmt.Sprintf("%s · %s#%d", title, m.PullRequest.Repository, m.PullRequest.Number)
	}
	return title
}

func (m Model) headerLine(th theme.Theme, width int) string {
	pr := m.PullRequest
	head, base := pr.Head, pr.Base
	if head == "" {
		head = "(head)"
	}
	if base == "" {
		base = "(base)"
	}
	status := th.Badge(th.AgentRunStatusStyle(pr.Status), pr.Status)
	available := max(0, width-lipgloss.Width(status)-1)
	return text.Justify(th.Muted.Render(text.Truncate(head+"  →  "+base, available)), status, width)
}

func (m Model) conversationLines(th theme.Theme, width int) []string {
	body := ""
	if m.Issue != nil {
		if m.markdown != nil {
			m.markdown.SetSource(m.Issue.Body)
			m.markdown.SetWidth(width)
			body = strings.TrimSpace(m.markdown.Render())
		}
	}
	if body == "" {
		body = th.Muted.Render("No description.")
	}
	rows := append(strings.Split(body, "\n"), "")
	if len(m.PullRequest.Comments) == 0 {
		return append(rows, th.Muted.Render("No review comments yet."))
	}
	return append(rows, comments.Render(th, nil, m.PullRequest.Comments, width)...)
}

func (m Model) diff(th theme.Theme, inner, width, body int) string {
	files := m.Files()
	if len(files) == 0 {
		loaded, err := m.DiffState()
		switch {
		case err != nil:
			return th.Muted.Render(text.Truncate("No diff available — "+err.Error(), inner))
		case !loaded:
			return th.Muted.Render("Computing the diff from the daemon…")
		case strings.TrimSpace(m.Filter.Value) != "":
			return th.Muted.Render("No files match " + m.Filter.Value)
		default:
			return th.Muted.Render("No changes between base and head.")
		}
	}
	treeWidth := min(30, max(18, inner/3))
	paneWidth := diffPaneWidth(width)
	if layoutTierNarrow(width) {
		treeWidth = inner
	}
	treeRows := make([]string, 0, len(files))
	for i, file := range files {
		marker, style := "  ", th.Muted
		if i == m.FileIndex {
			marker, style = th.Selected.Render("› "), th.Selected
		}
		treeRows = append(treeRows, zones.Mark(zones.DiffFile(i), text.Truncate(marker+style.Render(shortPath(file.Path, treeWidth-2)), treeWidth)))
	}
	windows := diffWindows(width, body, len(files))
	treeWindow := min(windows.tree, max(1, len(treeRows)))
	treeOffset := scroll.Follow(len(treeRows), treeWindow, m.FileIndex)
	tree := strings.Join(padRows(scroll.Window(treeRows, treeWindow, treeOffset), treeWindow), "\n")
	pane := th.Muted.Render("Select a file.")
	if file, ok := m.SelectedFile(); ok {
		if m.DiffLayout == TwoColumn {
			pane = twoColumnPane(th, *m.PullRequest, file, paneWidth)
		} else {
			lines := []string{diffControls(th, paneWidth), th.Accent.Render(text.Truncate("@@ "+file.Path, paneWidth))}
			for _, hunk := range file.Hunks {
				lines = append(lines, hunkLine(th, hunk.Kind, hunk.Text, paneWidth))
			}
			pane = strings.Join(lines, "\n")
		}
	}
	codeRows := strings.Split(pane, "\n")
	codeWindow := windows.code
	codeOffset := scroll.Clamp(len(codeRows), codeWindow, m.diffScroll)
	code := strings.Join(padRows(scroll.Window(codeRows, codeWindow, codeOffset), codeWindow), "\n")
	tree = zones.Mark(zones.DiffPane+"-tree", tree)
	code = zones.Mark(zones.DiffPane, code)
	if layoutTierNarrow(width) {
		return strings.Join([]string{tree, strings.Repeat("─", inner), code}, "\n")
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, tree, " ", text.Fit(code, paneWidth))
}

func diffPaneWidth(width int) int {
	inner := panel.ContentWidth(width)
	if layoutTierNarrow(width) {
		return inner
	}
	treeWidth := min(30, max(18, inner/3))
	return max(20, inner-treeWidth-1)
}

func padRows(rows []string, height int) []string {
	for len(rows) < height {
		rows = append(rows, "")
	}
	return rows
}

func diffControls(th theme.Theme, width int) string {
	return text.Truncate(th.Selected.Render("Inline")+th.Muted.Render(" / Two-column   v toggle"), width)
}

func hunkLine(th theme.Theme, kind byte, body string, width int) string {
	line := string(kind) + " " + body
	switch kind {
	case '+':
		line = th.Success.Render(line)
	case '-':
		line = th.Error.Render(line)
	default:
		line = th.Muted.Render(line)
	}
	return text.Truncate(line, width)
}

func twoColumnPane(th theme.Theme, pr netomatic.PullRequest, file DiffFile, width int) string {
	column := max(0, (width-1)/2)
	rows := []string{diffControlsTwoColumn(th, width), text.Pad(th.Accent.Render("BASE "+pr.Base), column) + "│" + text.Pad(th.Accent.Render("HEAD "+pr.Head), column)}
	for _, row := range twoColumnRows(file.Hunks) {
		if row.header {
			rows = append(rows, text.Truncate(th.Accent.Render(row.text), width))
			continue
		}
		left, right := "", ""
		if row.left != nil {
			left = hunkLine(th, row.left.Kind, row.left.Text, column)
		}
		if row.right != nil {
			right = hunkLine(th, row.right.Kind, row.right.Text, column)
		}
		rows = append(rows, text.Pad(left, column)+"│"+text.Pad(right, column))
	}
	return strings.Join(rows, "\n")
}

func diffControlsTwoColumn(th theme.Theme, width int) string {
	return text.Truncate(th.Muted.Render("Inline / ")+th.Selected.Render("Two-column")+th.Muted.Render("   v toggle"), width)
}

func shortPath(path string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(path) <= width {
		return path
	}
	parts := strings.Split(path, "/")
	return text.Truncate("…/"+parts[len(parts)-1], width)
}

func (m Model) actions(th theme.Theme, width int) string {
	if m.Terminal() {
		line := th.Accent.Render("o") + th.Muted.Render(" open on GitHub   "+m.PullRequest.Status)
		if m.Issue != nil && normalize(m.Issue.State) == "pr" {
			return line + "\n" + zones.Mark(zones.MergeButton, th.Selected.Render("[ M Merge ]")) + th.Muted.Render("  current pull request")
		}
		return line
	}
	if m.Filter.Active {
		return th.Muted.Render("filter: " + m.Filter.Value + "   ⏎ apply   esc clear")
	}
	line := strings.Join([]string{"R retry", "K kill retry", "V request review"}, "   ")
	return strings.Join([]string{text.Truncate(line, width), zones.Mark(zones.MergeButton, th.Selected.Render("[ Merge ]")) + th.Muted.Render("  requires click + confirm")}, "\n")
}
