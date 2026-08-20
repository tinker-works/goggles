package dialogs

import (
	"strings"
	"testing"

	"github.com/tinker-works/donsy/netomatic"
)

func TestState_ShouldOfferOnlyKnownTransitionsAndForceEscape(t *testing.T) {
	spec, states := State(&netomatic.Epic{State: "Review"})
	if len(states) == 0 || len(spec.Choices) != len(states)+1 || !strings.Contains(spec.Choices[len(spec.Choices)-1], "Force") {
		t.Fatalf("unexpected state choices: states=%v spec=%+v", states, spec)
	}
}

func TestComment_ShouldCycleIssueAndMatchingPullRequests(t *testing.T) {
	spec, targets := Comment(&netomatic.Epic{PullRequests: []netomatic.PullRequest{{ID: "pr", IssueID: "issue", Title: "Fix"}}}, "issue")
	if len(targets) != 2 || len(spec.Cycle) != 2 || targets[1].Kind != netomatic.PullRequestCommentTarget {
		t.Fatalf("unexpected comment dialog: spec=%+v targets=%+v", spec, targets)
	}
}
