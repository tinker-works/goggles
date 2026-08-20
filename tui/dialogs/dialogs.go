// Package dialogs owns shared modal specifications and maps submissions to
// public-daemon actions.
package dialogs

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/tinker-works/donsy/netomatic"
	"github.com/tinker-works/goggles/tui/actions"
	"github.com/tinker-works/goggles/tui/components/modal"
	"github.com/tinker-works/goggles/tui/viewmodel"
)

const (
	ModalEpic         = "epic"
	ModalIssue        = "issue"
	ModalPullRequest  = "pull-request"
	ModalComment      = "comment"
	ModalState        = "state"
	ModalForceState   = "force-state"
	ModalCloseIssue   = "close-issue"
	ModalCloseEpic    = "close-epic"
	ModalMerge        = "merge-pull-request"
	ModalRetry        = "retry-pull-request"
	ModalKillRun      = "kill-run"
	ModalResetIssue   = "reset-issue"
	ModalAgentRole    = "agent-role"
	ModalBranchPrefix = "branch-prefix"
)

var IDs = []string{
	ModalEpic, ModalIssue, ModalPullRequest, ModalComment, ModalState,
	ModalForceState, ModalCloseIssue, ModalCloseEpic, ModalMerge, ModalKillRun,
	ModalResetIssue, ModalRetry, ModalAgentRole, ModalBranchPrefix,
}

func Owns(id string) bool {
	for _, candidate := range IDs {
		if candidate == id {
			return true
		}
	}
	return false
}

var EpicStates = []string{"Concept", "Refine", "Review", "ChangesRequested", "Proposed", "Ready", "Done", "Closed", "Failed"}

const MaxCodingRounds = 5

func NewEpic(repositories []string) modal.Spec {
	return modal.Spec{ID: ModalEpic, Title: "New epic", Fields: []modal.Field{
		{Prompt: "Title"}, {Prompt: "Branch prefix (optional)"},
	}, Body: true, Submit: "Create", Options: repositories, OptionsPrompt: "Repositories (none checked = all)"}
}

func BranchPrefix(epic *netomatic.Epic, approving bool) modal.Spec {
	spec := modal.Spec{ID: ModalBranchPrefix, Title: "Branch prefix",
		Explain: "Names the tracker item this epic's branches belong to.\nLeave it empty to name branches after the issue alone.", Submit: "Save"}
	if epic != nil {
		spec.Fields = []modal.Field{{Prompt: "Branch prefix", Value: epic.BranchPrefix}}
	}
	if approving {
		spec.Title, spec.Submit = "Approve epic — branch prefix", "Approve"
	}
	return spec
}

func NewIssue() modal.Spec {
	return modal.Spec{ID: ModalIssue, Title: "New issue", Fields: []modal.Field{
		{Prompt: "Title"}, {Prompt: "Repository"},
	}, Body: true, Submit: "Create"}
}

func PullRequest() modal.Spec {
	return modal.Spec{ID: ModalPullRequest, Title: "Record pull request", Fields: []modal.Field{
		{Prompt: "Title"}, {Prompt: "Repository"}, {Prompt: "Head"}, {Prompt: "Base"},
	}, Submit: "Record"}
}

func CloseIssue() modal.Spec {
	return modal.Spec{ID: ModalCloseIssue, Title: "Close issue", Message: "Close the selected issue? Any open pull request on it closes too.", Submit: "Close"}
}

func CloseEpic() modal.Spec {
	return modal.Spec{ID: ModalCloseEpic, Title: "Close epic", Message: "Close the selected epic?", Submit: "Close"}
}

func Merge(title, repository, base string) modal.Spec {
	if base == "" {
		base = "its base branch"
	}
	target := title
	if repository != "" {
		target += " (" + repository + ")"
	}
	return modal.Spec{ID: ModalMerge, Title: "Merge pull request",
		Message: "Merge " + target + " into " + base + "?\nThis cannot be undone.", Submit: "Merge"}
}

func Retry(title string, codingRounds, granted int) modal.Spec {
	if title == "" {
		title = "this pull request"
	}
	return modal.Spec{ID: ModalRetry, Title: "Retry pull request", Message: fmt.Sprintf(
		"Grant %s another coding round?\n%d spent of %d allowed.", title, codingRounds, MaxCodingRounds+granted), Submit: "Grant"}
}

func KillRun(role, subject string) modal.Spec {
	target := role
	if subject != "" {
		target += " round on " + subject
	}
	return modal.Spec{ID: ModalKillRun, Title: "Kill run", Message: "Stop the " + target + "?\nWork done so far in this round is lost.", Submit: "Kill"}
}

func ResetIssue(title string) modal.Spec {
	if title == "" {
		title = "this issue"
	}
	return modal.Spec{ID: ModalResetIssue, Title: "Reset issue", Message: "Reset " + title + "?\nThis cancels agents, deletes sandboxes, branches, transcripts, and run history. The current pull request closes and the issue reopens.", Submit: "Reset"}
}

func AgentRole(role, agentName, variant string) modal.Spec {
	return modal.Spec{ID: ModalAgentRole, Title: "Agent role — " + role, Fields: []modal.Field{
		{Prompt: "Agent", Value: agentName}, {Prompt: "Variant", Value: variant},
	}, Submit: "Save"}
}

type Context struct {
	Epic           *netomatic.Epic
	IssueID        string
	PullRequestID  string
	CommentTargets []viewmodel.CommentTarget
	StateOptions   []string
	RunID          string
	Role           string
	ReadyOnSubmit  bool
}

func Comment(epic *netomatic.Epic, issueID string) (modal.Spec, []viewmodel.CommentTarget) {
	targets := viewmodel.CommentTargets(epic, issueID)
	labels := make([]string, len(targets))
	for i, target := range targets {
		labels[i] = target.Label
	}
	return modal.Spec{ID: ModalComment, Title: "Add comment", Body: true, Cycle: labels, Submit: "Post"}, targets
}

const forceStateChoice = "Force a state… (debug)"

func legalTransitions(state string) []string {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "concept":
		return []string{"Refine", "Closed", "Failed"}
	case "refine":
		return []string{"Review", "Closed", "Failed"}
	case "review":
		return []string{"ChangesRequested", "Proposed", "Closed", "Failed"}
	case "changesrequested", "changes-requested":
		return []string{"Refine", "Review", "Closed", "Failed"}
	case "proposed":
		return []string{"ChangesRequested", "Ready", "Closed", "Failed"}
	case "ready":
		return []string{"ChangesRequested", "Done", "Closed", "Failed"}
	case "done":
		return []string{"Closed", "Failed"}
	case "failed":
		return []string{"Concept", "Closed"}
	default:
		return nil
	}
}

func State(epic *netomatic.Epic) (modal.Spec, []string) {
	var states []string
	if epic != nil {
		states = legalTransitions(epic.State)
	}
	choices := append(append([]string(nil), states...), forceStateChoice)
	spec := modal.Spec{ID: ModalState, Title: "Set epic state", Choices: choices}
	if epic != nil && len(states) == 0 {
		spec.Explain = "No legal moves from " + epic.State + "."
	}
	return spec, states
}

func ForceState(epic *netomatic.Epic) modal.Spec {
	selected := 0
	choices := append([]string(nil), EpicStates...)
	if epic != nil {
		for i, state := range choices {
			if state == epic.State {
				selected = i
			}
		}
	}
	return modal.Spec{ID: ModalForceState, Title: "Force epic state (debug)", Explain: "Skips the state machine. For an epic the loop has stranded; the loop may not expect what this produces.", Choices: choices, Selected: selected}
}

func ResolveSubmit(client netomatic.Client, project netomatic.Project, ctx Context, msg modal.SubmittedMsg, _ ...string) tea.Cmd {
	switch msg.ID {
	case ModalEpic:
		return actions.CreateEpic(client, project, valueAt(msg.Values, 0), "", msg.Body, valueAt(msg.Values, 1), msg.Options)
	case ModalBranchPrefix:
		if ctx.Epic == nil {
			return nil
		}
		save := actions.SetBranchPrefix(client, project, ctx.Epic.ID, valueAt(msg.Values, 0))
		if !ctx.ReadyOnSubmit {
			return save
		}
		return actions.ApproveEpic(client, project, ctx.Epic.ID, valueAt(msg.Values, 0))
	case ModalIssue:
		if ctx.Epic == nil {
			return nil
		}
		return actions.CreateIssue(client, project, ctx.Epic.ID, ctx.IssueID, valueAt(msg.Values, 0), msg.Body, valueAt(msg.Values, 1))
	case ModalPullRequest:
		if ctx.Epic == nil {
			return nil
		}
		return actions.CreatePullRequest(client, project, ctx.Epic.ID, ctx.IssueID, valueAt(msg.Values, 0), valueAt(msg.Values, 1), valueAt(msg.Values, 2), valueAt(msg.Values, 3))
	case ModalComment:
		if ctx.Epic == nil || msg.Cycle < 0 || msg.Cycle >= len(ctx.CommentTargets) {
			return nil
		}
		target := ctx.CommentTargets[msg.Cycle]
		return actions.AddComment(client, project, ctx.Epic.ID, target.ID, target.Kind, msg.Body)
	case ModalState:
		if ctx.Epic == nil || msg.Choice < 0 {
			return nil
		}
		if msg.Choice < len(ctx.StateOptions) {
			return actions.TransitionEpic(client, project, ctx.Epic.ID, ctx.StateOptions[msg.Choice])
		}
		return modal.Open(ForceState(ctx.Epic))
	case ModalForceState:
		if ctx.Epic == nil || msg.Choice < 0 || msg.Choice >= len(EpicStates) {
			return nil
		}
		return actions.ForceEpicState(client, project, ctx.Epic.ID, EpicStates[msg.Choice])
	case ModalCloseIssue:
		if ctx.Epic == nil {
			return nil
		}
		return actions.CloseIssue(client, project, ctx.Epic.ID, ctx.IssueID)
	case ModalCloseEpic:
		if ctx.Epic == nil {
			return nil
		}
		return actions.CloseEpic(client, project, ctx.Epic.ID)
	case ModalMerge:
		if ctx.Epic == nil || ctx.PullRequestID == "" {
			return nil
		}
		return actions.MergePullRequest(client, project, ctx.Epic.ID, ctx.PullRequestID)
	case ModalRetry:
		if ctx.Epic == nil || ctx.PullRequestID == "" {
			return nil
		}
		return actions.RetryPullRequest(client, project, ctx.Epic.ID, ctx.PullRequestID)
	case ModalResetIssue:
		if ctx.Epic == nil || ctx.PullRequestID == "" {
			return nil
		}
		return actions.ResetIssue(client, project, ctx.Epic.ID, ctx.PullRequestID)
	case ModalKillRun:
		if ctx.RunID == "" {
			return nil
		}
		return actions.CancelAgentRun(client, ctx.RunID)
	case ModalAgentRole:
		if ctx.Role == "" {
			return nil
		}
		return actions.SetAgentRole(client, project, ctx.Role, valueAt(msg.Values, 0), valueAt(msg.Values, 1))
	default:
		return nil
	}
}

func valueAt(values []string, index int) string {
	if index < 0 || index >= len(values) {
		return ""
	}
	return values[index]
}
