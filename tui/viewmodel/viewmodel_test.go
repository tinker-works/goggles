package viewmodel

import (
	"testing"

	"github.com/tinker-works/donsy/netomatic"
)

func TestBuildIssueTree_ShouldFlattenPublicEpicDepthFirst(t *testing.T) {
	epic := &netomatic.Epic{Issues: []netomatic.Issue{
		{ID: "root"}, {ID: "child", ParentID: "root"},
		{ID: "grandchild", ParentID: "child"}, {ID: "orphan", ParentID: "missing"},
	}}
	tree := BuildIssueTree(epic)
	if tree.Len() != 3 || tree.Rows[1].ID != "child" || tree.Rows[2].ID != "grandchild" {
		t.Fatalf("unexpected issue order: %+v", tree.Rows)
	}
}

func TestBoardLanes_ShouldUsePublicIssueStates(t *testing.T) {
	lanes := BoardLanes([]netomatic.Epic{{ID: "e", State: "Ready", Issues: []netomatic.Issue{
		{ID: "root"}, {ID: "issue", ParentID: "root", State: "open"},
	}}}, GroupByEpic, "")
	if len(lanes) != 1 || len(lanes[0].Issues(OpenColumn)) != 1 {
		t.Fatalf("expected an open issue column, got %+v", lanes)
	}
}

func TestCommentTargets_ShouldUsePublicTargetKinds(t *testing.T) {
	targets := CommentTargets(&netomatic.Epic{PullRequests: []netomatic.PullRequest{{ID: "pr", IssueID: "issue", Title: "Fix"}}}, "issue")
	if len(targets) != 2 || targets[1].Kind != netomatic.PullRequestCommentTarget {
		t.Fatalf("unexpected comment targets: %+v", targets)
	}
}
