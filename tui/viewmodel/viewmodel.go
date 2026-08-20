// Package viewmodel derives screen-ready presentation data from netomatic DTOs.
package viewmodel

import "github.com/tinker-works/donsy/netomatic"

// Public contract aliases keep screen code from importing transport details
// when it only needs to hold a presentation DTO.
type Project = netomatic.Project
type Epic = netomatic.Epic
type Issue = netomatic.Issue
type PullRequest = netomatic.PullRequest
type Comment = netomatic.Comment
type AgentRun = netomatic.AgentRun
type Sandbox = netomatic.Sandbox

// IssueTree is an epic's issues flattened into display order.
type IssueTree struct {
	Rows []netomatic.Issue
}

// BuildIssueTree flattens an epic's issues depth first. Orphans are omitted and
// an epic without a root produces an empty tree.
func BuildIssueTree(e *netomatic.Epic) IssueTree {
	if e == nil {
		return IssueTree{}
	}
	var root netomatic.Issue
	found := false
	for _, issue := range e.Issues {
		if issue.ParentID == "" {
			root, found = issue, true
			break
		}
	}
	if !found {
		return IssueTree{}
	}
	byParent := map[string][]netomatic.Issue{}
	for _, issue := range e.Issues {
		if issue.ID != root.ID {
			byParent[issue.ParentID] = append(byParent[issue.ParentID], issue)
		}
	}
	rows := []netomatic.Issue{root}
	var add func(string)
	add = func(parentID string) {
		for _, issue := range byParent[parentID] {
			rows = append(rows, issue)
			add(issue.ID)
		}
	}
	add(root.ID)
	return IssueTree{Rows: rows}
}

func (t IssueTree) Len() int { return len(t.Rows) }

// At returns the row at index, clamped to the tree's bounds.
func (t IssueTree) At(index int) (netomatic.Issue, bool) {
	if len(t.Rows) == 0 {
		return netomatic.Issue{}, false
	}
	if index < 0 {
		index = 0
	}
	if index >= len(t.Rows) {
		index = len(t.Rows) - 1
	}
	return t.Rows[index], true
}

// CommentTarget is a comment destination shown by the modal cycle.
type CommentTarget struct {
	ID    string
	Kind  netomatic.CommentTarget
	Label string
}

// CommentTargets lists an issue followed by its matching pull requests.
func CommentTargets(e *netomatic.Epic, issueID string) []CommentTarget {
	targets := []CommentTarget{{
		ID: issueID, Kind: netomatic.IssueCommentTarget, Label: "Issue: " + issueID,
	}}
	if e == nil {
		return targets
	}
	for _, pullRequest := range e.PullRequests {
		if pullRequest.IssueID == issueID {
			targets = append(targets, CommentTarget{
				ID: pullRequest.ID, Kind: netomatic.PullRequestCommentTarget,
				Label: "PR: " + pullRequest.Title,
			})
		}
	}
	return targets
}

// ProjectSummary is the compact project-picker row.
type ProjectSummary struct {
	ProjectID uint
	Epics     int
	Running   int
	Err       error
}
