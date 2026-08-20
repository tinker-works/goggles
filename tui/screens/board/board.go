// Package board implements the project epic list and its zoomed issue board.
package board

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/tinker-works/donsy/netomatic"
	"github.com/tinker-works/goggles/tui/components/filter"
	"github.com/tinker-works/goggles/tui/viewmodel"
)

type Model struct {
	Epics     []netomatic.Epic
	Runners   []viewmodel.Runner
	Sandboxes []netomatic.Sandbox
	Settings  []netomatic.AgentSettings

	GroupBy viewmodel.GroupBy
	Filter  filter.Model

	Lane   int
	Column viewmodel.Column
	Index  int
	Zoomed bool

	AttentionCollapsed bool

	lanes     []viewmodel.Lane
	RailIndex int
	activity  map[string][]int
	lastSizes map[string]int64
}

const railActivityCells = 6

func New() Model { return Model{AttentionCollapsed: true} }

func (m Model) RecordActivity(sizes map[string]int64) Model {
	if m.activity == nil {
		m.activity, m.lastSizes = map[string][]int{}, map[string]int64{}
	}
	for runID, size := range sizes {
		last, seen := m.lastSizes[runID]
		m.lastSizes[runID] = size
		if !seen {
			continue
		}
		delta := max(0, int(size-last))
		m.activity[runID] = append(m.activity[runID], delta)
		if values := m.activity[runID]; len(values) > railActivityCells {
			m.activity[runID] = values[len(values)-railActivityCells:]
		}
	}
	return m
}

func (m Model) ActivityFor(runID string) []int { return m.activity[runID] }

func (m Model) ToggleAttention() Model {
	m.AttentionCollapsed = !m.AttentionCollapsed
	return m
}

// Reset clears project data while retaining the user's grouping preference.
func (m Model) Reset() Model {
	fresh := New()
	fresh.GroupBy = m.GroupBy
	return fresh
}

func (m Model) SetEpics(epics []netomatic.Epic) Model {
	focused := ""
	if m.Lane >= 0 && m.Lane < len(m.lanes) {
		focused = m.lanes[m.Lane].Key
	}
	m.Epics = append([]netomatic.Epic(nil), epics...)
	m = m.rebuild()
	for i := range m.lanes {
		if m.lanes[i].Key == focused {
			m.Lane = i
			break
		}
	}
	return m.clamp()
}

func (m Model) SetEpic(epic netomatic.Epic) Model {
	for i := range m.Epics {
		if m.Epics[i].ID == epic.ID {
			m.Epics[i] = epic
			return m.rebuild().clamp()
		}
	}
	return m
}

func (m Model) SetRuns(runs []netomatic.AgentRun, now time.Time) Model {
	m.Runners = viewmodel.Runners(runs, m.Epics, now)
	m.RailIndex = min(m.RailIndex, max(0, len(m.Runners)-1))
	return m
}

func (m Model) SetSandboxes(sandboxes []netomatic.Sandbox) Model {
	m.Sandboxes = append([]netomatic.Sandbox(nil), sandboxes...)
	return m
}

func (m Model) SetSettings(settings []netomatic.AgentSettings) Model {
	m.Settings = append([]netomatic.AgentSettings(nil), settings...)
	return m
}

func (m Model) Lanes() []viewmodel.Lane { return m.lanes }

func (m Model) rebuild() Model {
	m.lanes = viewmodel.BoardLanes(m.Epics, m.GroupBy, m.Filter.Value())
	return m
}

func (m Model) clamp() Model {
	m.Lane = max(0, min(m.Lane, max(0, len(m.lanes)-1)))
	if !m.Zoomed {
		m.Column = viewmodel.EpicColumn
	} else if m.Column < viewmodel.OpenColumn {
		m.Column = viewmodel.OpenColumn
	}
	if m.Column > viewmodel.LastColumn {
		m.Column = viewmodel.LastColumn
	}
	m.Index = max(0, min(m.Index, max(0, len(m.focusedIssues())-1)))
	return m
}

func (m Model) focusedIssues() []viewmodel.BoardIssue {
	if !m.Zoomed || m.Lane < 0 || m.Lane >= len(m.lanes) {
		return nil
	}
	return m.lanes[m.Lane].Issues(m.Column)
}

func (m Model) Zoom() Model {
	if m.Lane < 0 || m.Lane >= len(m.lanes) {
		return m
	}
	m.Zoomed = true
	m.Column, m.Index = viewmodel.OpenColumn, 0
	return m.clamp()
}

func (m Model) ZoomAt(lane int) Model {
	if lane < 0 || lane >= len(m.lanes) {
		return m
	}
	m.Lane = lane
	return m.Zoom()
}

func (m Model) Unzoom() Model {
	m.Zoomed = false
	m.Column, m.Index = viewmodel.EpicColumn, 0
	return m
}

func (m Model) ZoomedLane() (viewmodel.Lane, bool) {
	if !m.Zoomed || m.Lane < 0 || m.Lane >= len(m.lanes) {
		return viewmodel.Lane{}, false
	}
	return m.lanes[m.Lane], true
}

func (m Model) MoveLeft() Model {
	if m.Column > viewmodel.OpenColumn {
		m.Column--
		m.Index = 0
	}
	return m.clamp()
}

func (m Model) MoveRight() Model {
	if m.Column < viewmodel.LastColumn {
		m.Column++
		m.Index = 0
	}
	return m.clamp()
}

func (m Model) MoveUp() Model {
	if m.Zoomed {
		m.Index = max(0, m.Index-1)
		return m.clamp()
	}
	m.Lane = max(0, m.Lane-1)
	return m.clamp()
}

func (m Model) MoveDown() Model {
	if m.Zoomed {
		m.Index++
		return m.clamp()
	}
	m.Lane++
	return m.clamp()
}

func (m Model) JumpLane(delta int) Model {
	if len(m.lanes) == 0 {
		return m
	}
	m.Lane = max(0, min(len(m.lanes)-1, m.Lane+delta))
	m.Index = 0
	return m.clamp()
}

func (m Model) SelectCard(lane int, column viewmodel.Column, index int) Model {
	if lane < 0 || lane >= len(m.lanes) {
		return m
	}
	m.Lane, m.Column, m.Index = lane, column, max(0, index)
	return m.clamp()
}

func (m Model) CycleGroup() Model {
	m.GroupBy = m.GroupBy.Next()
	m = m.rebuild()
	m.Lane, m.Column, m.Index = 0, viewmodel.EpicColumn, 0
	return m
}

func (m Model) StartFilter() Model {
	m.Filter = filter.New("filter")
	return m
}

func (m Model) ApplyFilterKey(msg tea.KeyPressMsg) Model {
	if msg.Text == "" && msg.Code == tea.KeyEscape {
		m.Filter = filter.Model{}
		return m.rebuild().clamp()
	}
	m.Filter, _ = m.Filter.Update(msg)
	return m.rebuild().clamp()
}

func (m Model) SetFilter(value string) Model {
	m.Filter.SetValue(value)
	return m.rebuild().clamp()
}

func (m Model) FocusedEpic() (netomatic.Epic, bool) {
	if m.Lane < 0 || m.Lane >= len(m.lanes) {
		return netomatic.Epic{}, false
	}
	lane := m.lanes[m.Lane]
	if lane.Epic != nil {
		return *lane.Epic, true
	}
	if issue, ok := m.FocusedIssue(); ok {
		for _, epic := range m.Epics {
			if epic.ID == issue.EpicID {
				return epic, true
			}
		}
	}
	return netomatic.Epic{}, false
}

func (m Model) FocusedIssue() (viewmodel.BoardIssue, bool) {
	issues := m.focusedIssues()
	if m.Column == viewmodel.EpicColumn || m.Index < 0 || m.Index >= len(issues) {
		return viewmodel.BoardIssue{}, false
	}
	return issues[m.Index], true
}

func (m Model) FocusedRunner() (viewmodel.Runner, bool) {
	if m.RailIndex < 0 || m.RailIndex >= len(m.Runners) {
		return viewmodel.Runner{}, false
	}
	return m.Runners[m.RailIndex], true
}

func (m Model) Attention(now time.Time) []viewmodel.AttentionItem {
	return viewmodel.AttentionItems(m.Epics, m.Runners, m.Sandboxes, m.Settings, now)
}

func (m Model) IssueCount() int {
	count := 0
	for _, epic := range m.Epics {
		count += max(0, len(epic.Issues)-1)
	}
	return count
}
