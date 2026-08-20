package dialogs

import (
	"context"
	"io"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/tinker-works/donsy/netomatic"
	"github.com/tinker-works/goggles/tui/actions"
	"github.com/tinker-works/goggles/tui/components/modal"
)

type dialogClient struct {
	netomatic.Client
	request netomatic.CreateEpicRequest
}

func (c *dialogClient) CreateEpic(_ context.Context, request netomatic.CreateEpicRequest) (netomatic.CreateEpicResponse, error) {
	c.request = request
	return netomatic.CreateEpicResponse{}, nil
}

type dialogCommandModel struct {
	command  tea.Cmd
	finished *actions.FinishedMsg
}

func (m *dialogCommandModel) Init() tea.Cmd { return m.command }

func (m *dialogCommandModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if finished, ok := msg.(actions.FinishedMsg); ok {
		m.finished = &finished
		return m, func() tea.Msg { return tea.Quit() }
	}
	return m, nil
}

func (m *dialogCommandModel) View() tea.View { return tea.NewView("") }

func runDialogFinished(t *testing.T, command tea.Cmd) actions.FinishedMsg {
	t.Helper()
	model := &dialogCommandModel{command: command}
	program := tea.NewProgram(model, tea.WithInput(nil), tea.WithOutput(io.Discard), tea.WithoutRenderer(), tea.WithoutSignalHandler())
	if _, err := program.Run(); err != nil {
		t.Fatalf("run command: %v", err)
	}
	if model.finished == nil {
		t.Fatal("command did not finish")
	}
	return *model.finished
}

func TestState_ShouldOfferOnlyKnownTransitionsAndForceEscape(t *testing.T) {
	spec, states := State(&netomatic.Epic{State: "Review"})
	if len(states) == 0 || len(spec.Choices) != len(states)+1 || !strings.Contains(spec.Choices[len(spec.Choices)-1], "Force") {
		t.Fatalf("unexpected state choices: states=%v spec=%+v", states, spec)
	}
}

func TestState_ShouldAllowClosingFailedEpics(t *testing.T) {
	_, states := State(&netomatic.Epic{State: "Failed"})
	if !hasState(states, "Closed") {
		t.Fatalf("expected failed epics to offer Closed, got %v", states)
	}
}

func TestComment_ShouldCycleIssueAndMatchingPullRequests(t *testing.T) {
	spec, targets := Comment(&netomatic.Epic{PullRequests: []netomatic.PullRequest{{ID: "pr", IssueID: "issue", Title: "Fix"}}}, "issue")
	if len(targets) != 2 || len(spec.Cycle) != 2 || targets[1].Kind != netomatic.PullRequestCommentTarget {
		t.Fatalf("unexpected comment dialog: spec=%+v targets=%+v", spec, targets)
	}
}

func TestResolveSubmit_ShouldNotTreatTheCurrentUserAsAnEpicAssignee(t *testing.T) {
	client := &dialogClient{}
	msg := modal.SubmittedMsg{ID: ModalEpic, Values: []string{"Epic", ""}, Body: "body"}

	result := runDialogFinished(t, ResolveSubmit(client, netomatic.Project{Name: "demo"}, Context{}, msg, "alice"))
	if result.Err != nil {
		t.Fatalf("expected epic creation to succeed: %v", result.Err)
	}
	if client.request != (netomatic.CreateEpicRequest{Project: "demo", Title: "Epic", Description: "body"}) {
		t.Fatalf("unexpected create request: %+v", client.request)
	}
}

func hasState(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
