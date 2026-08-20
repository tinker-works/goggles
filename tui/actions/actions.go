// Package actions wraps public daemon calls as tea.Cmd values and defines the
// messages screens consume.
package actions

import (
	"context"
	"log/slog"
	"strconv"

	tea "charm.land/bubbletea/v2"
	"github.com/tinker-works/donsy/netomatic"
	"github.com/tinker-works/goggles/tui/components/transcript"
	"github.com/tinker-works/goggles/tui/viewmodel"
)

type LogLoadedMsg struct {
	Lines []string
	Next  int64
	Err   error
}

const maxLogLines = netomatic.MaxDaemonLogLines

// ReadLog polls complete daemon-log records. The daemon holds an unterminated
// final record for the next offset, so the next call naturally completes it.
func ReadLog(client netomatic.Client, from int64) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return LogLoadedMsg{Err: ErrUnsupported{Reason: "no daemon client is configured"}}
		}
		page, err := client.ReadDaemonLog(context.Background(), from, maxLogLines)
		return LogLoadedMsg{Lines: page.Lines, Next: page.NextOffset, Err: err}
	}
}

type StartedMsg struct{}
type DoneMsg struct{}

func track(work func() tea.Msg) tea.Cmd {
	return tea.Sequence(
		func() tea.Msg { return StartedMsg{} },
		func() tea.Msg { return work() },
		func() tea.Msg { return DoneMsg{} },
	)
}

func unavailable(what string) tea.Msg {
	return FinishedMsg{Status: what + " is not available", Err: ErrUnsupported{Reason: "no daemon client is configured"}}
}

type Reload uint8

const (
	ReloadNone Reload = iota
	ReloadProjects
	ReloadEpics
	ReloadEpic
	ReloadOrganisations
)

type FinishedMsg struct {
	Status string
	Err    error
	Reload Reload
	EpicID string
	Synced bool
	Silent bool
}

type ProjectsLoadedMsg struct {
	Projects []netomatic.Project
	Err      error
}

type EpicsLoadedMsg struct {
	Epics  []netomatic.Epic
	Err    error
	Silent bool
}

type EpicLoadedMsg struct {
	Epic   netomatic.Epic
	Err    error
	Silent bool
}

func projectKey(project netomatic.Project) string {
	if project.Name != "" {
		return project.Name
	}
	return strconv.FormatUint(uint64(project.ID), 10)
}

func LoadProjects(client netomatic.Client) tea.Cmd {
	return track(func() tea.Msg {
		if client == nil {
			return ProjectsLoadedMsg{}
		}
		projects, err := client.ListProjects(context.Background())
		return ProjectsLoadedMsg{Projects: projects, Err: err}
	})
}

func LoadEpics(client netomatic.Client, project netomatic.Project) tea.Cmd {
	return track(loadEpics(client, project, false))
}

func PollEpics(client netomatic.Client, project netomatic.Project) tea.Cmd {
	return loadEpics(client, project, true)
}

func loadEpics(client netomatic.Client, project netomatic.Project, silent bool) func() tea.Msg {
	return func() tea.Msg {
		if client == nil {
			return EpicsLoadedMsg{Silent: silent}
		}
		response, err := client.ListEpics(context.Background(), netomatic.ListEpicsRequest{Project: projectKey(project)})
		if err != nil {
			slog.Error("load epics failed", "project", project.Name, "error", err)
		}
		return EpicsLoadedMsg{Epics: response.Epics, Err: err, Silent: silent}
	}
}

func LoadEpic(client netomatic.Client, project netomatic.Project, id string) tea.Cmd {
	return track(loadEpic(client, project, id, false))
}

func PollEpic(client netomatic.Client, project netomatic.Project, id string) tea.Cmd {
	return loadEpic(client, project, id, true)
}

func loadEpic(client netomatic.Client, project netomatic.Project, id string, silent bool) func() tea.Msg {
	return func() tea.Msg {
		if client == nil {
			return EpicLoadedMsg{Silent: silent}
		}
		response, err := client.GetEpic(context.Background(), netomatic.GetEpicRequest{Project: projectKey(project), Epic: id})
		if err != nil {
			slog.Error("load epic failed", "project", project.Name, "epic", id, "error", err)
		}
		return EpicLoadedMsg{Epic: response.Epic, Err: err, Silent: silent}
	}
}

func TouchProject(client netomatic.Client, project netomatic.Project) tea.Cmd {
	return func() tea.Msg {
		if client != nil {
			_ = client.OpenProject(context.Background(), netomatic.ProjectPath{ProjectID: project.ID})
		}
		return nil
	}
}

type ProjectCreatedMsg struct {
	Project netomatic.Project
	Err     error
}

func CreateProject(client netomatic.Client, name string) tea.Cmd {
	return track(func() tea.Msg {
		if client == nil {
			return unavailable("Creating projects")
		}
		response, err := client.CreateProject(context.Background(), netomatic.CreateProjectRequest{Name: name})
		return ProjectCreatedMsg{Project: response, Err: err}
	})
}

func CreateEpic(client netomatic.Client, project netomatic.Project, title, assignee, body, branchPrefix string, repositories []string) tea.Cmd {
	return track(func() tea.Msg {
		if client == nil {
			return unavailable("Creating epics")
		}
		response, err := client.CreateEpic(context.Background(), netomatic.CreateEpicRequest{
			Project: projectKey(project), Title: title, Description: body,
		})
		if err == nil && branchPrefix != "" {
			_, err = client.PrefixEpic(context.Background(), netomatic.PrefixEpicRequest{
				Project: projectKey(project), Epic: response.Epic.ID, Prefix: branchPrefix,
			})
		}
		return FinishedMsg{Status: "Epic created", Err: err, Reload: ReloadEpics}
	})
}

func SetBranchPrefix(client netomatic.Client, project netomatic.Project, epicID, prefix string) tea.Cmd {
	return track(func() tea.Msg {
		if client == nil {
			return unavailable("Setting a branch prefix")
		}
		_, err := client.PrefixEpic(context.Background(), netomatic.PrefixEpicRequest{Project: projectKey(project), Epic: epicID, Prefix: prefix})
		return FinishedMsg{Status: "Branch prefix saved", Err: err, Reload: ReloadEpic, EpicID: epicID}
	})
}

func CreateIssue(client netomatic.Client, project netomatic.Project, epicID, parentID, title, body, repository string) tea.Cmd {
	return track(func() tea.Msg {
		if client == nil {
			return unavailable("Creating issues")
		}
		_, err := client.CreateIssue(context.Background(), netomatic.CreateIssueRequest{
			Project: projectKey(project), Epic: epicID, Title: title, Description: body,
		})
		return FinishedMsg{Status: "Issue created", Err: err, Reload: ReloadEpic, EpicID: epicID}
	})
}

func TransitionEpic(client netomatic.Client, project netomatic.Project, epicID, state string) tea.Cmd {
	return transitionEpic(client, project, epicID, state, false)
}

func ForceEpicState(client netomatic.Client, project netomatic.Project, epicID, state string) tea.Cmd {
	return transitionEpic(client, project, epicID, state, true)
}

func transitionEpic(client netomatic.Client, project netomatic.Project, epicID, state string, forced bool) tea.Cmd {
	return track(func() tea.Msg {
		if client == nil {
			return unavailable("Epic transitions")
		}
		_, err := client.TransitionEpic(context.Background(), netomatic.TransitionEpicRequest{Project: projectKey(project), Epic: epicID, Status: state})
		status := "Epic state set to " + state
		if forced {
			status = "Epic state forced to " + state
		}
		return FinishedMsg{Status: status, Err: err, Reload: ReloadEpic, EpicID: epicID}
	})
}

func CreatePullRequest(client netomatic.Client, project netomatic.Project, epicID, issueID, title, repository, head, base string) tea.Cmd {
	return track(func() tea.Msg {
		if client == nil {
			return unavailable("Recording pull requests")
		}
		err := client.CreatePullRequest(context.Background(), netomatic.CreatePullRequestPath{ProjectID: project.ID, EpicID: epicID}, netomatic.CreatePullRequestRequest{
			IssueID: issueID, Title: title, Repository: repository, Head: head, Base: base,
		})
		return FinishedMsg{Status: "Pull request recorded", Err: err, Reload: ReloadEpic, EpicID: epicID}
	})
}

func AddComment(client netomatic.Client, project netomatic.Project, epicID, targetID string, target netomatic.CommentTarget, body string) tea.Cmd {
	return track(func() tea.Msg {
		if client == nil {
			return unavailable("Commenting")
		}
		err := client.AddComment(context.Background(), netomatic.AddCommentPath{ProjectID: project.ID, EpicID: epicID}, netomatic.AddCommentRequest{TargetID: targetID, Target: target, Body: body})
		return FinishedMsg{Status: "Comment added", Err: err, Reload: ReloadEpic, EpicID: epicID}
	})
}

func CloseIssue(client netomatic.Client, project netomatic.Project, epicID, issueID string) tea.Cmd {
	return track(func() tea.Msg {
		if client == nil {
			return unavailable("Closing issues")
		}
		_, err := client.CloseIssue(context.Background(), netomatic.CloseIssueRequest{Project: projectKey(project), Epic: epicID, Issue: issueID})
		return FinishedMsg{Status: "Issue closed", Err: err, Reload: ReloadEpic, EpicID: epicID}
	})
}

func CloseEpic(client netomatic.Client, project netomatic.Project, epicID string) tea.Cmd {
	return track(func() tea.Msg {
		if client == nil {
			return unavailable("Closing epics")
		}
		_, err := client.CloseEpic(context.Background(), netomatic.CloseEpicRequest{Project: projectKey(project), Epic: epicID})
		return FinishedMsg{Status: "Epic closed", Err: err, Reload: ReloadEpic, EpicID: epicID}
	})
}

func MergePullRequest(client netomatic.Client, project netomatic.Project, epicID, pullRequestID string) tea.Cmd {
	return track(func() tea.Msg {
		if client == nil {
			return unavailable("Merging")
		}
		response, err := client.MergePullRequest(context.Background(), netomatic.MergePullRequestPath{ProjectID: project.ID, EpicID: epicID, PullRequestID: pullRequestID})
		status := "Pull request merged"
		if response.Outcome == netomatic.MergeOutcomeReturnedToCoding {
			status = "Branch is behind — sent back to coding"
		}
		return FinishedMsg{Status: status, Err: err, Reload: ReloadEpic, EpicID: epicID}
	})
}

func ClosePullRequest(client netomatic.Client, project netomatic.Project, epicID, pullRequestID string) tea.Cmd {
	return track(func() tea.Msg {
		if client == nil {
			return unavailable("Closing pull requests")
		}
		err := client.TransitionPullRequest(context.Background(), netomatic.TransitionPullRequestPath{ProjectID: project.ID, EpicID: epicID, PullRequestID: pullRequestID}, netomatic.TransitionPullRequestRequest{Status: "closed"})
		return FinishedMsg{Status: "Pull request closed", Err: err, Reload: ReloadEpic, EpicID: epicID}
	})
}

func ResetIssue(client netomatic.Client, project netomatic.Project, epicID, pullRequestID string) tea.Cmd {
	return track(func() tea.Msg {
		if client == nil {
			return unavailable("Resetting issues")
		}
		err := client.ResetIssue(context.Background(), netomatic.ResetIssuePath{ProjectID: project.ID, EpicID: epicID, PullRequestID: pullRequestID})
		return FinishedMsg{Status: "Issue reset", Err: err, Reload: ReloadEpic, EpicID: epicID}
	})
}

func RetryPullRequest(client netomatic.Client, project netomatic.Project, epicID, pullRequestID string) tea.Cmd {
	return track(func() tea.Msg {
		if client == nil {
			return unavailable("Retrying")
		}
		err := client.GrantCodingRound(context.Background(), netomatic.GrantCodingRoundPath{ProjectID: project.ID, EpicID: epicID, PullRequestID: pullRequestID})
		return FinishedMsg{Status: "Granted another coding round", Err: err, Reload: ReloadEpic, EpicID: epicID}
	})
}

func UpdateRepositories(client netomatic.Client, project netomatic.Project, repositories []string) tea.Cmd {
	return track(func() tea.Msg {
		if client == nil {
			return unavailable("Updating repositories")
		}
		err := client.UpdateProjectRepositories(context.Background(), netomatic.UpdateProjectRepositoriesPath{ProjectID: project.ID}, netomatic.UpdateProjectRepositoriesRequest{Repositories: repositories})
		return FinishedMsg{Status: "Repositories updated", Err: err, Reload: ReloadEpics}
	})
}

type ProjectRepositoriesLoadedMsg struct {
	Linked []string
	Err    error
}

func LoadProjectRepositories(client netomatic.Client, project netomatic.Project) tea.Cmd {
	return track(func() tea.Msg {
		if client == nil {
			return ProjectRepositoriesLoadedMsg{}
		}
		linked, err := client.ListProjectRepositories(context.Background(), netomatic.ListProjectRepositoriesPath{ProjectID: project.ID})
		return ProjectRepositoriesLoadedMsg{Linked: linked, Err: err}
	})
}

func SyncRepositories(client netomatic.Client) tea.Cmd {
	return track(func() tea.Msg {
		if client == nil {
			return unavailable("Syncing repositories")
		}
		err := client.SyncRepositories(context.Background())
		return FinishedMsg{Status: "Repositories synced", Err: err, Synced: err == nil}
	})
}

type RepositoriesLoadedMsg struct {
	Repositories []netomatic.Repository
	Err          error
}

func LoadRepositories(client netomatic.Client) tea.Cmd {
	return track(func() tea.Msg {
		if client == nil {
			return RepositoriesLoadedMsg{}
		}
		repositories, err := client.ListRepositories(context.Background())
		return RepositoriesLoadedMsg{Repositories: repositories, Err: err}
	})
}

type OrganisationsLoadedMsg struct {
	Organisations []netomatic.Organisation
	Err           error
}

func LoadOrganisations(client netomatic.Client) tea.Cmd {
	return track(func() tea.Msg {
		if client == nil {
			return OrganisationsLoadedMsg{}
		}
		organisations, err := client.ListOrganisations(context.Background())
		return OrganisationsLoadedMsg{Organisations: organisations, Err: err}
	})
}

func AddOrganisation(client netomatic.Client, name string) tea.Cmd {
	return track(func() tea.Msg {
		if client == nil {
			return unavailable("Adding organisations")
		}
		err := client.AddOrganisation(context.Background(), netomatic.AddOrganisationRequest{Name: name})
		return FinishedMsg{Status: "Organisation added", Err: err, Reload: ReloadOrganisations}
	})
}

func RemoveOrganisation(client netomatic.Client, name string) tea.Cmd {
	return track(func() tea.Msg {
		if client == nil {
			return unavailable("Removing organisations")
		}
		err := client.RemoveOrganisation(context.Background(), netomatic.RemoveOrganisationPath{Name: name})
		return FinishedMsg{Status: "Organisation removed", Err: err, Reload: ReloadOrganisations}
	})
}

func DiscoverOrganisations(client netomatic.Client) tea.Cmd {
	return track(func() tea.Msg {
		if client == nil {
			return unavailable("Discovering organisations")
		}
		_, err := client.DiscoverOrganisations(context.Background())
		return FinishedMsg{Status: "Organisations discovered", Err: err, Reload: ReloadOrganisations}
	})
}

type SetupStateLoadedMsg struct {
	Project netomatic.Project
	State   netomatic.SetupState
	Err     error
}

func LoadSetupState(client netomatic.Client, project netomatic.Project) tea.Cmd {
	return track(func() tea.Msg {
		if client == nil {
			return SetupStateLoadedMsg{Project: project}
		}
		state, err := client.StoreSetup(context.Background(), netomatic.ProjectPath{ProjectID: project.ID})
		return SetupStateLoadedMsg{Project: project, State: state, Err: err}
	})
}

type StoreInitialisedMsg struct{ Err error }

func InitialiseStore(client netomatic.Client, project netomatic.Project, model, variant string, repositories []string) tea.Cmd {
	return track(func() tea.Msg {
		if client == nil {
			return StoreInitialisedMsg{Err: ErrUnsupported{Reason: "no daemon client is configured"}}
		}
		err := client.InitialiseStore(context.Background(), netomatic.ProjectPath{ProjectID: project.ID}, netomatic.InitialiseStoreRequest{Model: model, Variant: variant, Repositories: repositories})
		return StoreInitialisedMsg{Err: err}
	})
}

func Unsupported(reason string) tea.Cmd {
	return track(func() tea.Msg { return FinishedMsg{Status: reason, Err: ErrUnsupported{Reason: reason}} })
}

type ErrUnsupported struct{ Reason string }

func (e ErrUnsupported) Error() string { return e.Reason }

type RunsLoadedMsg struct {
	Runs   []netomatic.AgentRun
	Err    error
	Silent bool
}

type SandboxesLoadedMsg struct {
	Sandboxes []netomatic.Sandbox
	Err       error
}

type AgentSettingsLoadedMsg struct {
	Settings []netomatic.AgentSettings
	Err      error
}

func CancelAgentRun(client netomatic.Client, runID string) tea.Cmd {
	return track(func() tea.Msg {
		if client == nil {
			return unavailable("Cancelling runs")
		}
		_, err := client.CancelAgentRun(context.Background(), netomatic.CancelAgentRunRequest{Run: runID})
		return FinishedMsg{Status: "Run cancelled", Err: err}
	})
}

type RunActivityMsg struct{ Sizes map[string]int64 }

func PollRunActivity(client netomatic.Client, runIDs []string) tea.Cmd {
	if len(runIDs) == 0 {
		return nil
	}
	return func() tea.Msg {
		if client == nil {
			return nil
		}
		sizes := map[string]int64{}
		for _, runID := range runIDs {
			response, err := client.AgentActivity(context.Background(), netomatic.AgentActivityRequest{Run: runID})
			if err == nil {
				sizes[runID] = int64(len(response.Activity))
			}
		}
		if len(sizes) == 0 {
			return nil
		}
		return RunActivityMsg{Sizes: sizes}
	}
}

type PullRequestDiffLoadedMsg struct {
	PullRequestID string
	Diff          string
	Err           error
}

func LoadPullRequestDiff(client netomatic.Client, project netomatic.Project, epicID, pullRequestID string) tea.Cmd {
	return track(func() tea.Msg {
		if client == nil {
			return PullRequestDiffLoadedMsg{PullRequestID: pullRequestID, Err: ErrUnsupported{Reason: "no daemon client is configured"}}
		}
		response, err := client.GetPullRequestDiff(context.Background(), netomatic.GetPullRequestDiffPath{ProjectID: project.ID, EpicID: epicID, PullRequestID: pullRequestID})
		return PullRequestDiffLoadedMsg{PullRequestID: pullRequestID, Diff: response.Diff, Err: err}
	})
}

type RunOutputLoadedMsg struct {
	RunID   string
	Entries []transcript.Entry
	Output  string
	Next    int64
	Done    bool
	Err     error
}

func ReadRunOutput(client netomatic.Client, runID string, from int64) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return RunOutputLoadedMsg{RunID: runID, Err: ErrUnsupported{Reason: "no daemon client is configured"}}
		}
		response, err := client.RunOutput(context.Background(), netomatic.RunOutputRequest{Run: runID, Offset: from})
		output := response.Output.Output
		entries := []transcript.Entry{}
		if output != "" {
			entries = append(entries, transcript.Entry{Kind: transcript.Text, Text: output})
		}
		return RunOutputLoadedMsg{RunID: runID, Entries: entries, Output: output, Next: from + int64(len(output)), Done: response.Output.Done, Err: err}
	}
}

type ProjectSummariesLoadedMsg struct {
	Summaries []viewmodel.ProjectSummary
	Err       error
}

func LoadProjectSummaries(client netomatic.Client) tea.Cmd {
	return track(loadProjectSummaries(client))
}

func PollProjectSummaries(client netomatic.Client) tea.Cmd { return loadProjectSummaries(client) }

func loadProjectSummaries(client netomatic.Client) func() tea.Msg {
	return func() tea.Msg {
		if client == nil {
			return ProjectSummariesLoadedMsg{}
		}
		response, err := client.ListProjectSummaries(context.Background())
		rows := make([]viewmodel.ProjectSummary, 0, len(response))
		for _, summary := range response {
			rows = append(rows, viewmodel.ProjectSummary{ProjectID: summary.Project.ID, Epics: summary.Epics, Running: summary.Running, Err: summary.Error})
		}
		return ProjectSummariesLoadedMsg{Summaries: rows, Err: err}
	}
}

func LoadRuns(client netomatic.Client, project netomatic.Project) tea.Cmd {
	return track(loadRuns(client, project, false))
}

func PollRuns(client netomatic.Client, project netomatic.Project) tea.Cmd {
	return loadRuns(client, project, true)
}

func loadRuns(client netomatic.Client, project netomatic.Project, silent bool) func() tea.Msg {
	return func() tea.Msg {
		if client == nil {
			return RunsLoadedMsg{Silent: silent}
		}
		response, err := client.ListAgentRuns(context.Background(), netomatic.ListAgentRunsRequest{Project: projectKey(project)})
		return RunsLoadedMsg{Runs: response.Runs, Err: err, Silent: silent}
	}
}

func LoadSandboxes(client netomatic.Client, project netomatic.Project) tea.Cmd {
	return track(loadSandboxes(client))
}

func PollSandboxes(client netomatic.Client, project netomatic.Project) tea.Cmd {
	return loadSandboxes(client)
}

func loadSandboxes(client netomatic.Client) func() tea.Msg {
	return func() tea.Msg {
		if client == nil {
			return SandboxesLoadedMsg{}
		}
		response, err := client.ListSandboxes(context.Background(), netomatic.ListSandboxesRequest{})
		return SandboxesLoadedMsg{Sandboxes: response.Sandboxes, Err: err}
	}
}

func LoadAgentSettings(client netomatic.Client, project netomatic.Project) tea.Cmd {
	return track(loadAgentSettings(client, project))
}

func PollAgentSettings(client netomatic.Client, project netomatic.Project) tea.Cmd {
	return loadAgentSettings(client, project)
}

func loadAgentSettings(client netomatic.Client, project netomatic.Project) func() tea.Msg {
	return func() tea.Msg {
		if client == nil {
			return AgentSettingsLoadedMsg{}
		}
		response, err := client.GetAgentSettings(context.Background(), netomatic.GetAgentSettingsRequest{Project: projectKey(project)})
		return AgentSettingsLoadedMsg{Settings: response.Settings, Err: err}
	}
}

func SetAgentRole(client netomatic.Client, project netomatic.Project, role, agentName, variant string) tea.Cmd {
	return track(func() tea.Msg {
		return FinishedMsg{Status: "Assigning roles is not available", Err: ErrUnsupported{Reason: "the public daemon contract has no role update operation"}}
	})
}

func ForgetProject(client netomatic.Client, project netomatic.Project) tea.Cmd {
	return track(func() tea.Msg {
		if client == nil {
			return unavailable("Forgetting projects")
		}
		err := client.ForgetProject(context.Background(), netomatic.ProjectPath{ProjectID: project.ID})
		return FinishedMsg{Status: "Forgot " + project.Name, Err: err, Reload: ReloadProjects}
	})
}
