// Package epicdetail implements the single-epic issue tree and detail view.
package epicdetail

import (
	tea "charm.land/bubbletea/v2"

	"github.com/tinker-works/donsy/netomatic"
	"github.com/tinker-works/goggles/tui/components/commentbox"
	"github.com/tinker-works/goggles/tui/components/markdown"
	"github.com/tinker-works/goggles/tui/components/scroll"
	"github.com/tinker-works/goggles/tui/components/tree"
	"github.com/tinker-works/goggles/tui/viewmodel"
)

const branchNamespace = "gm/"

type Model struct {
	Epic  *netomatic.Epic
	Index int

	Comment commentbox.Model

	ShowProposal   bool
	ShowOutput     bool
	outputTop      int
	outputDetached bool
	outputFocused  bool

	detailScroll int
	tree         viewmodel.IssueTree
	markdown     *markdown.Model
}

func New() Model {
	renderer := markdown.New()
	return Model{markdown: &renderer, Comment: commentbox.New()}
}

// SetEpic loads an epic while retaining the cursor and reader position across a
// refresh of the same record.
func (m Model) SetEpic(epic netomatic.Epic) Model {
	same := m.Epic != nil && m.Epic.ID == epic.ID
	m.Epic = &epic
	m.tree = viewmodel.BuildIssueTree(&epic)
	if !same {
		m.Index, m.detailScroll = 0, 0
	}
	m.Index = max(0, min(m.Index, max(0, m.tree.Len()-1)))
	if m.markdown == nil {
		renderer := markdown.New()
		m.markdown = &renderer
	}
	return m
}

func (m *Model) Resize(width int) { m.Comment.Resize(width) }

func (m Model) Reset() Model {
	m.Epic = nil
	m.tree = viewmodel.IssueTree{}
	m.Index, m.detailScroll = 0, 0
	m.Comment = commentbox.New()
	m.ShowProposal, m.ShowOutput = false, false
	m.outputTop, m.outputDetached, m.outputFocused = 0, false, false
	return m
}

func (m Model) Loaded() bool { return m.Epic != nil }

func (m Model) MoveUp() Model {
	m.Index, m.detailScroll = max(0, m.Index-1), 0
	return m
}

func (m Model) MoveDown() Model {
	m.Index = min(max(0, m.tree.Len()-1), m.Index+1)
	m.detailScroll = 0
	return m
}

func (m Model) SelectRow(index int) Model {
	if index < 0 || index >= m.tree.Len() {
		return m
	}
	m.Index, m.detailScroll = index, 0
	return m
}

func (m Model) DetailScroll() int { return m.detailScroll }

func (m Model) ScrollDetail(rows, total, window int) Model {
	m.detailScroll = max(0, min(m.detailScroll+rows, max(0, total-window)))
	return m
}

func (m Model) SelectedIssue() (netomatic.Issue, bool) { return m.tree.At(m.Index) }

func (m Model) SelectedIssueID() string {
	issue, ok := m.SelectedIssue()
	if !ok {
		return ""
	}
	return issue.ID
}

func (m Model) AtRoot() bool {
	issue, ok := m.SelectedIssue()
	return ok && issue.ParentID == ""
}

func (m Model) PullRequestFor(issueID string) (netomatic.PullRequest, bool) {
	if m.Epic == nil {
		return netomatic.PullRequest{}, false
	}
	for _, pullRequest := range m.Epic.PullRequests {
		if pullRequest.IssueID == issueID {
			return pullRequest, true
		}
	}
	return netomatic.PullRequest{}, false
}

func (m Model) Commenting() bool { return m.Comment.Focused() }

func (m Model) StartComment() (Model, tea.Cmd) {
	box, cmd := m.Comment.Focus()
	m.Comment = box
	return m, cmd
}

func (m Model) EndComment(keep bool) Model {
	m.Comment = m.Comment.Blur(keep)
	return m
}

func (m Model) UpdateComment(msg tea.Msg) (Model, tea.Cmd) {
	box, cmd := m.Comment.Update(msg)
	m.Comment = box
	return m, cmd
}

func (m Model) ToggleProposal() Model {
	m.ShowProposal = !m.ShowProposal
	return m
}

func (m Model) ToggleOutput() Model {
	m.ShowOutput = !m.ShowOutput
	m.outputTop, m.outputDetached = 0, false
	m.outputFocused = m.ShowOutput
	return m
}

func (m Model) OutputTop() int        { return m.outputTop }
func (m Model) FollowingOutput() bool { return !m.outputDetached }
func (m Model) OutputFocused() bool   { return m.outputFocused }

func (m Model) ToggleOutputFocus() Model {
	if m.ShowOutput {
		m.outputFocused = !m.outputFocused
	}
	return m
}

func (m Model) FocusOutput() Model {
	if m.ShowOutput {
		m.outputFocused = true
	}
	return m
}

func (m Model) FocusDetail() Model {
	m.outputFocused = false
	return m
}

func (m Model) ScrollOutput(rows, total, window int) Model {
	end := max(0, total-window)
	top := end
	if m.outputDetached {
		top = scroll.Clamp(total, window, m.outputTop)
	}
	top = scroll.Clamp(total, window, top+rows)
	m.outputTop, m.outputDetached = top, top < end
	return m
}

func (m Model) Rows() []tree.Row {
	if m.Epic == nil {
		return nil
	}
	depths := map[string]int{}
	children := map[string]int{}
	for _, issue := range m.Epic.Issues {
		if issue.ParentID != "" {
			children[issue.ParentID]++
		}
	}
	numbers := map[string]int{}
	for _, pullRequest := range m.Epic.PullRequests {
		if pullRequest.Number > 0 {
			numbers[pullRequest.IssueID] = pullRequest.Number
		}
	}
	rows := make([]tree.Row, 0, m.tree.Len())
	for _, issue := range m.tree.Rows {
		depth := 0
		if issue.ParentID != "" {
			depth = depths[issue.ParentID] + 1
		}
		depths[issue.ID] = depth
		rows = append(rows, tree.Row{Issue: issue, Depth: depth,
			HasChild: children[issue.ID] > 0, Number: numbers[issue.ID]})
	}
	return rows
}
