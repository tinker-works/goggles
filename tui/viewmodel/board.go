package viewmodel

import (
	"sort"
	"strings"

	"github.com/tinker-works/donsy/netomatic"
	uifilter "github.com/tinker-works/goggles/tui/components/filter"
)

type Column uint8

const (
	EpicColumn Column = iota
	OpenColumn
	CodingColumn
	ReviewColumn
	PRColumn
	StaleColumn
	MergedColumn
)

const LastColumn = MergedColumn

var columnStates = map[Column]string{
	OpenColumn: "open", CodingColumn: "coding", ReviewColumn: "review",
	PRColumn: "pr", StaleColumn: "stale", MergedColumn: "merged",
}

var ColumnTitles = []string{"EPIC", "OPEN", "CODING", "REVIEW", "IN PR", "STALE", "MERGED"}

type GroupBy uint8

const (
	GroupByEpic GroupBy = iota
	GroupByRepository
	GroupByAssignee
)

func (g GroupBy) String() string {
	switch g {
	case GroupByRepository:
		return "repository"
	case GroupByAssignee:
		return "assignee"
	default:
		return "epic"
	}
}

func (g GroupBy) Next() GroupBy {
	if g >= GroupByAssignee {
		return GroupByEpic
	}
	return g + 1
}

type BoardIssue struct {
	Issue   netomatic.Issue
	EpicID  string
	Number  int
	Matched bool
}

type Lane struct {
	Key      string
	Title    string
	Epic     *netomatic.Epic
	State    string
	Assignee string
	Columns  [LastColumn][]BoardIssue
	Merged   int
	Total    int
	Drafting bool
	Matched  bool
}

func (l Lane) Issues(column Column) []BoardIssue {
	if column == EpicColumn || column > LastColumn {
		return nil
	}
	return l.Columns[column-1]
}

func (l Lane) Count(column Column) int { return len(l.Issues(column)) }

func Drafting(state string) bool {
	switch normalize(state) {
	case "ready", "done", "closed", "failed":
		return false
	default:
		return true
	}
}

// BoardLanes groups epics into fixed lanes along the selected axis.
func BoardLanes(epics []netomatic.Epic, groupBy GroupBy, filter string) []Lane {
	epics = withoutClosedEpics(epics)
	switch groupBy {
	case GroupByRepository:
		return groupedLanes(epics, filter, repositoryGrouping)
	case GroupByAssignee:
		return groupedLanes(epics, filter, assigneeGrouping)
	default:
		return epicLanes(epics, filter)
	}
}

func withoutClosedEpics(epics []netomatic.Epic) []netomatic.Epic {
	kept := make([]netomatic.Epic, 0, len(epics))
	for _, epic := range epics {
		if normalize(epic.State) == "closed" {
			continue
		}
		kept = append(kept, epic)
	}
	return kept
}

func epicLanes(epics []netomatic.Epic, filter string) []Lane {
	lanes := make([]Lane, 0, len(epics))
	for i := range epics {
		epic := epics[i]
		lane := Lane{Key: epic.ID, Title: epic.Title, Epic: &epics[i], State: epic.State,
			Assignee: epic.Assignee, Drafting: Drafting(epic.State)}
		lane.Matched = matches(filter, epic.Title, epic.Assignee) || matches(filter, epic.Repositories...)
		addIssues(&lane, epic, filter)
		lanes = append(lanes, lane)
	}
	return lanes
}

type grouping struct {
	epicKeys func(netomatic.Epic) []string
	issueKey func(netomatic.Epic, netomatic.Issue) string
}

var repositoryGrouping = grouping{
	epicKeys: func(epic netomatic.Epic) []string {
		keys := map[string]struct{}{}
		for _, repository := range epic.Repositories {
			keys[repository] = struct{}{}
		}
		for _, issue := range epic.Issues {
			if issue.Repository != "" {
				keys[issue.Repository] = struct{}{}
			}
		}
		if len(keys) == 0 {
			return []string{noRepository}
		}
		return sorted(keys)
	},
	issueKey: func(_ netomatic.Epic, issue netomatic.Issue) string {
		if issue.Repository == "" {
			return noRepository
		}
		return issue.Repository
	},
}

var assigneeGrouping = grouping{
	epicKeys: func(epic netomatic.Epic) []string { return []string{assigneeOf(epic)} },
	issueKey: func(epic netomatic.Epic, _ netomatic.Issue) string { return assigneeOf(epic) },
}

const (
	noRepository = "(no repository)"
	unassigned   = "(unassigned)"
)

func assigneeOf(epic netomatic.Epic) string {
	if strings.TrimSpace(epic.Assignee) == "" {
		return unassigned
	}
	return epic.Assignee
}

func groupedLanes(epics []netomatic.Epic, filter string, by grouping) []Lane {
	order := []string{}
	byKey := map[string]*Lane{}
	lane := func(key, state string) *Lane {
		if existing, ok := byKey[key]; ok {
			return existing
		}
		order = append(order, key)
		byKey[key] = &Lane{Key: key, Title: key, State: state}
		return byKey[key]
	}
	for _, epic := range epics {
		epicMatches := matches(filter, epic.Title, epic.Assignee)
		for _, key := range by.epicKeys(epic) {
			if epicMatches || matches(filter, key) {
				lane(key, epic.State).Matched = true
			} else {
				lane(key, epic.State)
			}
		}
		numbers := pullRequestNumbers(epic)
		for _, issue := range epic.Issues {
			if issue.ParentID == "" {
				continue
			}
			key := by.issueKey(epic, issue)
			if key != "" {
				addIssue(lane(key, epic.State), epic, issue, numbers, filter)
			}
		}
	}
	sort.Strings(order)
	lanes := make([]Lane, 0, len(order))
	for _, key := range order {
		value := *byKey[key]
		value.Drafting = value.Total == 0
		lanes = append(lanes, value)
	}
	return lanes
}

func sorted(keys map[string]struct{}) []string {
	out := make([]string, 0, len(keys))
	for key := range keys {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

func addIssues(lane *Lane, epic netomatic.Epic, filter string) {
	numbers := pullRequestNumbers(epic)
	for _, issue := range epic.Issues {
		if issue.ParentID != "" {
			addIssue(lane, epic, issue, numbers, filter)
		}
	}
}

func addIssue(lane *Lane, epic netomatic.Epic, issue netomatic.Issue, numbers map[string]int, filter string) {
	lane.Total++
	if normalize(issue.State) == "merged" {
		lane.Merged++
	}
	column, ok := columnFor(issue.State)
	if !ok {
		return
	}
	card := BoardIssue{Issue: issue, EpicID: epic.ID, Number: numbers[issue.ID],
		Matched: matches(filter, issue.Title, issue.Repository)}
	if card.Matched {
		lane.Matched = true
	}
	lane.Columns[column-1] = append(lane.Columns[column-1], card)
}

func columnFor(state string) (Column, bool) {
	for column, want := range columnStates {
		if normalize(want) == normalize(state) {
			return column, true
		}
	}
	return EpicColumn, false
}

func pullRequestNumbers(epic netomatic.Epic) map[string]int {
	numbers := map[string]int{}
	for _, pullRequest := range epic.PullRequests {
		if pullRequest.Number > 0 {
			numbers[pullRequest.IssueID] = pullRequest.Number
		}
	}
	return numbers
}

func matches(value string, candidates ...string) bool { return uifilter.Matches(value, candidates...) }
