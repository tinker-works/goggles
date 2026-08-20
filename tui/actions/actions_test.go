package actions

import (
	"context"
	"errors"
	"io"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/tinker-works/donsy/netomatic"
)

type actionClient struct {
	fakeClient

	epicRequest  netomatic.CreateEpicRequest
	epicResponse netomatic.CreateEpicResponse
	epicErr      error
	epicCalls    int

	prefixRequest netomatic.PrefixEpicRequest
	prefixErr     error
	prefixCalls   int

	issueRequest  netomatic.CreateIssueRequest
	issueResponse netomatic.CreateIssueResponse
	issueErr      error
	issueCalls    int

	transitionRequest netomatic.TransitionEpicRequest
	transitionErr     error
	transitionCalls   int
}

func (c *actionClient) CreateEpic(_ context.Context, request netomatic.CreateEpicRequest) (netomatic.CreateEpicResponse, error) {
	c.epicCalls++
	c.epicRequest = request
	return c.epicResponse, c.epicErr
}

func (c *actionClient) PrefixEpic(_ context.Context, request netomatic.PrefixEpicRequest) (netomatic.PrefixEpicResponse, error) {
	c.prefixCalls++
	c.prefixRequest = request
	return netomatic.PrefixEpicResponse{}, c.prefixErr
}

func (c *actionClient) CreateIssue(_ context.Context, request netomatic.CreateIssueRequest) (netomatic.CreateIssueResponse, error) {
	c.issueCalls++
	c.issueRequest = request
	return c.issueResponse, c.issueErr
}

func (c *actionClient) TransitionEpic(_ context.Context, request netomatic.TransitionEpicRequest) (netomatic.TransitionEpicResponse, error) {
	c.transitionCalls++
	c.transitionRequest = request
	return netomatic.TransitionEpicResponse{}, c.transitionErr
}

type actionCommandModel struct {
	command  tea.Cmd
	finished *FinishedMsg
}

func (m *actionCommandModel) Init() tea.Cmd { return m.command }

func (m *actionCommandModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if finished, ok := msg.(FinishedMsg); ok {
		m.finished = &finished
		return m, func() tea.Msg { return tea.Quit() }
	}
	return m, nil
}

func (m *actionCommandModel) View() tea.View { return tea.NewView("") }

func runFinished(t *testing.T, command tea.Cmd) FinishedMsg {
	t.Helper()
	model := &actionCommandModel{command: command}
	program := tea.NewProgram(model, tea.WithInput(nil), tea.WithOutput(io.Discard), tea.WithoutRenderer(), tea.WithoutSignalHandler())
	if _, err := program.Run(); err != nil {
		t.Fatalf("run command: %v", err)
	}
	if model.finished == nil {
		t.Fatal("command did not finish")
	}
	return *model.finished
}

func TestCreateEpic_ShouldRejectUnsupportedMetadataBeforeWriting(t *testing.T) {
	tests := []struct {
		name         string
		assignee     string
		repositories []string
	}{
		{name: "assignee", assignee: "alice"},
		{name: "repositories", repositories: []string{"origin"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &actionClient{}
			msg := runFinished(t, CreateEpic(client, netomatic.Project{Name: "demo"}, "Epic", test.assignee, "body", "", test.repositories))
			if msg.Err == nil || client.epicCalls != 0 {
				t.Fatalf("expected visible rejection without a daemon write: msg=%+v calls=%d", msg, client.epicCalls)
			}
		})
	}
}

func TestCreateEpic_ShouldSendSupportedFieldsAndPrefix(t *testing.T) {
	client := &actionClient{epicResponse: netomatic.CreateEpicResponse{Epic: netomatic.Epic{ID: "epic-1"}}}
	msg := runFinished(t, CreateEpic(client, netomatic.Project{Name: "demo"}, "Epic", "", "body", "feature", nil))
	if msg.Err != nil || client.epicCalls != 1 || client.prefixCalls != 1 {
		t.Fatalf("unexpected create result: msg=%+v epicCalls=%d prefixCalls=%d", msg, client.epicCalls, client.prefixCalls)
	}
	if client.epicRequest != (netomatic.CreateEpicRequest{Project: "demo", Title: "Epic", Description: "body"}) {
		t.Fatalf("unexpected epic request: %+v", client.epicRequest)
	}
	if client.prefixRequest != (netomatic.PrefixEpicRequest{Project: "demo", Epic: "epic-1", Prefix: "feature"}) {
		t.Fatalf("unexpected prefix request: %+v", client.prefixRequest)
	}
}

func TestApproveEpic_ShouldNotTransitionWhenPrefixSaveFails(t *testing.T) {
	prefixErr := errors.New("invalid prefix")
	client := &actionClient{prefixErr: prefixErr}

	msg := runFinished(t, ApproveEpic(client, netomatic.Project{Name: "demo"}, "epic-1", "feature"))
	if !errors.Is(msg.Err, prefixErr) || client.prefixCalls != 1 || client.transitionCalls != 0 {
		t.Fatalf("expected prefix failure to stop approval: msg=%+v prefixCalls=%d transitionCalls=%d", msg, client.prefixCalls, client.transitionCalls)
	}
}

func TestApproveEpic_ShouldTransitionAfterSavingPrefix(t *testing.T) {
	client := &actionClient{}

	msg := runFinished(t, ApproveEpic(client, netomatic.Project{Name: "demo"}, "epic-1", "feature"))
	if msg.Err != nil || client.prefixCalls != 1 || client.transitionCalls != 1 {
		t.Fatalf("unexpected approval result: msg=%+v prefixCalls=%d transitionCalls=%d", msg, client.prefixCalls, client.transitionCalls)
	}
	if client.transitionRequest != (netomatic.TransitionEpicRequest{Project: "demo", Epic: "epic-1", Status: "Ready"}) {
		t.Fatalf("unexpected transition request: %+v", client.transitionRequest)
	}
}

func TestCreateIssue_ShouldRejectUnsupportedMetadataBeforeWriting(t *testing.T) {
	tests := []struct {
		name       string
		parentID   string
		repository string
	}{
		{name: "parent", parentID: "parent-1"},
		{name: "repository", repository: "origin"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &actionClient{}
			msg := runFinished(t, CreateIssue(client, netomatic.Project{Name: "demo"}, "epic-1", test.parentID, "Issue", "body", test.repository))
			if msg.Err == nil || client.issueCalls != 0 {
				t.Fatalf("expected visible rejection without a daemon write: msg=%+v calls=%d", msg, client.issueCalls)
			}
		})
	}
}

func TestCreateIssue_ShouldSendSupportedFields(t *testing.T) {
	client := &actionClient{}
	msg := runFinished(t, CreateIssue(client, netomatic.Project{Name: "demo"}, "epic-1", "", "Issue", "body", ""))
	if msg.Err != nil || client.issueCalls != 1 {
		t.Fatalf("unexpected create result: msg=%+v calls=%d", msg, client.issueCalls)
	}
	if client.issueRequest != (netomatic.CreateIssueRequest{Project: "demo", Epic: "epic-1", Title: "Issue", Description: "body"}) {
		t.Fatalf("unexpected issue request: %+v", client.issueRequest)
	}
}

func TestForceEpicState_ShouldReportUnsupportedWithoutTransitioning(t *testing.T) {
	client := &actionClient{}
	msg := runFinished(t, ForceEpicState(client, netomatic.Project{Name: "demo"}, "epic-1", "Done"))
	if msg.Err == nil || client.transitionCalls != 0 {
		t.Fatalf("expected visible rejection without a transition: msg=%+v calls=%d", msg, client.transitionCalls)
	}
	var unsupported ErrUnsupported
	if !errors.As(msg.Err, &unsupported) {
		t.Fatalf("expected unsupported error, got %T: %v", msg.Err, msg.Err)
	}
}
