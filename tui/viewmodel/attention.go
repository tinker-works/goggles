package viewmodel

import (
	"fmt"
	"strings"
	"time"

	"github.com/tinker-works/donsy/netomatic"
)

const StallThreshold = 10 * time.Minute
const FailureWindow = time.Hour

type AttentionKind uint8

const (
	AttentionEpicFailed AttentionKind = iota
	AttentionRunFailed
	AttentionRunStuck
	AttentionOrphanedRun
	AttentionMissingProfile
	AttentionSandboxBroken
)

type AttentionItem struct {
	Kind    AttentionKind
	Subject string
	Detail  string
	EpicID  string
}

// AttentionItems identifies daemon states that need an explanation in the UI.
// The settings argument remains part of the presentation boundary so callers
// can supply the settings response as the public contract grows role metadata.
func AttentionItems(epics []netomatic.Epic, runners []Runner, sandboxes []netomatic.Sandbox, _ any, now time.Time) []AttentionItem {
	items := []AttentionItem{}
	for _, epic := range epics {
		if normalize(epic.State) == "failed" {
			items = append(items, AttentionItem{Kind: AttentionEpicFailed, Subject: epic.Title,
				Detail: "epic failed", EpicID: epic.ID})
		}
	}
	for _, runner := range runners {
		status := normalize(runner.Run.Status)
		subject := runner.Subject
		switch {
		case status == "failed" && failureAge(runner.Run, now) <= FailureWindow:
			items = append(items, AttentionItem{Kind: AttentionRunFailed, Subject: subject,
				Detail: failureDetail(runner.Run)})
		case status == "stalled" && failureAge(runner.Run, now) >= StallThreshold:
			items = append(items, AttentionItem{Kind: AttentionRunFailed, Subject: subject,
				Detail: fmt.Sprintf("agent run stalled — no progress for %s", roughDuration(failureAge(runner.Run, now)))})
		case runner.Orphaned:
			items = append(items, AttentionItem{Kind: AttentionOrphanedRun, Subject: subject,
				Detail: "agent run outlived its subject — safe to ignore"})
		case (status == "queued" || status == "admitted") && failureAge(runner.Run, now) >= StallThreshold:
			items = append(items, AttentionItem{Kind: AttentionRunStuck, Subject: subject,
				Detail: fmt.Sprintf("agent run stuck for %s", roughDuration(failureAge(runner.Run, now)))})
		}
	}
	for _, sandbox := range sandboxes {
		if normalize(sandbox.Status) == "broken" {
			items = append(items, AttentionItem{Kind: AttentionSandboxBroken, Subject: sandbox.Name,
				Detail: "Sandbox broken — reconcile will recreate it"})
		}
	}
	return items
}

func failureAge(run netomatic.AgentRun, now time.Time) time.Duration {
	at := parseTime(run.FinishedAt)
	if at.IsZero() {
		at = parseTime(run.StartedAt)
	}
	if at.IsZero() {
		return 0
	}
	return now.Sub(at)
}

func failureDetail(run netomatic.AgentRun) string {
	if run.Error != "" {
		return "agent run failed — " + run.Error
	}
	return "agent run failed"
}

func roughDuration(duration time.Duration) string {
	if duration < time.Hour {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	}
	return fmt.Sprintf("%dh", int(duration.Hours()))
}

func normalize(value string) string {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(value)), "_", "-")
}
