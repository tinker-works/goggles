// Package pullrequest implements the review conversation and diff reader.
package pullrequest

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/tinker-works/donsy/netomatic"
	"github.com/tinker-works/goggles/tui/components/commentbox"
	"github.com/tinker-works/goggles/tui/components/filter"
	"github.com/tinker-works/goggles/tui/components/markdown"
	"github.com/tinker-works/goggles/tui/components/scroll"
)

type Tab uint8

const (
	Conversation Tab = iota
	Diff
)

var TabLabels = []string{"Conversation", "Diff"}

type DiffLayout uint8

const (
	Inline DiffLayout = iota
	TwoColumn
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
	PullRequest *netomatic.PullRequest
	Issue       *netomatic.Issue
	EpicID      string

	Tab        Tab
	DiffLayout DiffLayout

	FileIndex  int
	treeScroll int
	diffScroll int
	diffWindow int
	diffTotal  int
	Filter     filterState
	Comment    commentbox.Model
	scroll     int

	files      []DiffFile
	diffLoaded bool
	diffErr    error
	markdown   *markdown.Model
}

func New() Model {
	renderer := markdown.New()
	return Model{markdown: &renderer, Comment: commentbox.New()}
}

func (m Model) Load(epicID string, pr netomatic.PullRequest, issue *netomatic.Issue) Model {
	same := m.PullRequest != nil && m.PullRequest.ID == pr.ID
	m.EpicID, m.PullRequest, m.Issue = epicID, &pr, issue
	if !same {
		m.Tab, m.FileIndex, m.scroll, m.DiffLayout = Conversation, 0, 0, Inline
		m.treeScroll, m.diffScroll, m.diffWindow, m.diffTotal = 0, 0, 0, 0
		m.Filter = filterState{}
		m.Comment = commentbox.New()
		m.files, m.diffLoaded, m.diffErr = nil, false, nil
	}
	m.FileIndex = max(0, min(m.FileIndex, max(0, len(m.Files())-1)))
	if m.markdown == nil {
		renderer := markdown.New()
		m.markdown = &renderer
	}
	return m
}

func (m Model) Reset() Model { return New() }
func (m Model) Loaded() bool { return m.PullRequest != nil }

func (m Model) Terminal() bool {
	return m.PullRequest != nil && normalize(m.PullRequest.Status) != "open"
}

func (m Model) SwitchTab() Model {
	if m.Tab == Conversation {
		m.Tab = Diff
	} else {
		m.Tab = Conversation
	}
	return m
}

func (m Model) DiffScroll() int     { return m.diffScroll }
func (m Model) DiffTreeScroll() int { return m.treeScroll }

func (m Model) ScrollDiff(rows, total, window int) Model {
	m.diffScroll = scroll.Clamp(total, window, m.diffScroll)
	m.diffScroll = scroll.Clamp(total, window, m.diffScroll+rows)
	return m
}

func (m Model) ToggleDiffLayout() Model {
	if m.DiffLayout == Inline {
		m.DiffLayout = TwoColumn
	} else {
		m.DiffLayout = Inline
	}
	m.diffScroll, m.diffTotal = 0, 0
	return m
}

func (m Model) Scroll() int { return m.scroll }

func (m Model) ScrollConversation(rows, total, window int) Model {
	m.scroll = max(0, min(m.scroll+rows, max(0, total-window)))
	return m
}

func (m Model) SelectTab(index int) Model {
	if index == int(Diff) {
		m.Tab = Diff
	} else {
		m.Tab = Conversation
	}
	return m
}

func (m Model) SetDiff(pullRequestID string, files []DiffFile, err error) Model {
	if m.PullRequest == nil || m.PullRequest.ID != pullRequestID {
		return m
	}
	selected := ""
	if file, ok := m.SelectedFile(); ok {
		selected = file.Path
	}
	m.files, m.diffLoaded, m.diffErr = files, true, err
	m.retainFile(selected)
	if len(m.Files()) == 0 {
		m.FileIndex, m.diffScroll = 0, 0
	} else if selected == "" {
		m.FileIndex = max(0, min(m.FileIndex, len(m.Files())-1))
	}
	m.clampDiffScroll(m.diffWindow)
	return m
}

func (m Model) DiffState() (bool, error) { return m.diffLoaded, m.diffErr }

func (m Model) Files() []DiffFile {
	kept := make([]DiffFile, 0, len(m.files))
	for _, file := range m.files {
		if filter.Matches(m.Filter.Value, file.Path) {
			kept = append(kept, file)
		}
	}
	return kept
}

func (m Model) SelectedFile() (DiffFile, bool) {
	files := m.Files()
	if m.FileIndex < 0 || m.FileIndex >= len(files) {
		return DiffFile{}, false
	}
	return files[m.FileIndex], true
}

func (m Model) MoveUp() Model {
	m.FileIndex = max(0, m.FileIndex-1)
	m.diffScroll = 0
	return m
}

func (m Model) MoveDown() Model {
	m.FileIndex = min(max(0, len(m.Files())-1), m.FileIndex+1)
	m.diffScroll = 0
	return m
}

func (m Model) SelectFile(index int) Model {
	if index < 0 || index >= len(m.Files()) {
		return m
	}
	m.FileIndex, m.diffScroll = index, 0
	return m
}

func (m Model) StartFilter() Model {
	m.Filter = m.Filter.Start()
	m.Tab = Diff
	return m
}

func (m Model) ApplyFilterKey(msg tea.KeyPressMsg) Model {
	selected := ""
	if file, ok := m.SelectedFile(); ok {
		selected = file.Path
	}
	m.Filter = applyFilterKey(m.Filter, msg)
	m.retainFile(selected)
	if len(m.Files()) == 0 {
		m.FileIndex, m.diffScroll = 0, 0
	} else if selected == "" {
		m.FileIndex = max(0, min(m.FileIndex, len(m.Files())-1))
	}
	m.clampDiffScroll(m.diffWindow)
	return m
}

func (m Model) SetFilter(value string) Model {
	selected := ""
	if file, ok := m.SelectedFile(); ok {
		selected = file.Path
	}
	m.Filter = m.Filter.Set(value)
	m.retainFile(selected)
	if len(m.Files()) == 0 {
		m.FileIndex, m.diffScroll = 0, 0
	} else if selected == "" {
		m.FileIndex = max(0, min(m.FileIndex, len(m.Files())-1))
	}
	m.clampDiffScroll(m.diffWindow)
	return m
}

func (m *Model) clampDiffScroll(window int) {
	if window <= 0 {
		window = 1
	}
	m.diffTotal = m.renderedDiffRows()
	m.diffScroll = scroll.Clamp(m.diffTotal, window, m.diffScroll)
}

func (m Model) renderedDiffRows() int {
	file, ok := m.SelectedFile()
	if !ok {
		return 0
	}
	if m.DiffLayout == TwoColumn {
		return 2 + len(twoColumnRows(file.Hunks))
	}
	return 2 + len(file.Hunks)
}

func (m *Model) retainFile(path string) {
	if path == "" {
		return
	}
	for i, file := range m.Files() {
		if file.Path == path {
			m.FileIndex = i
			return
		}
	}
	m.FileIndex, m.diffScroll = 0, 0
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

func (m *Model) Resize(width int) { m.Comment.Resize(width) }

func (m Model) Stale() bool {
	if m.PullRequest == nil || m.PullRequest.Reviews == 0 {
		return false
	}
	return m.PullRequest.ReviewedHead != m.PullRequest.Head || m.PullRequest.ReviewedBase != m.PullRequest.Base
}

func normalize(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.ReplaceAll(value, "_", "-")
}
