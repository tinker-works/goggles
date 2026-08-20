package viewmodel

import (
	"sort"
	"time"

	"github.com/tinker-works/donsy/netomatic"
)

type Runner struct {
	Run      netomatic.AgentRun
	Subject  string
	Orphaned bool
	Elapsed  time.Duration
	Branch   string
}

func (r Runner) Live() bool {
	switch normalize(r.Run.Status) {
	case "queued", "admitted", "running":
		return true
	default:
		return false
	}
}

func (r Runner) Verdict() string {
	if normalize(r.Run.Status) == "succeeded" && (normalize(r.Run.Agent) == "issue-reviewer" || normalize(r.Run.Agent) == "pr-reviewer") {
		return "approve"
	}
	return "—"
}

// Runners derives rail rows from public run DTOs, newest first. The public run
// contract does not expose the epic or issue a run operates on, so its tracker
// identifier must not be presented as a subject.
func Runners(runs []netomatic.AgentRun, _ []netomatic.Epic, now time.Time) []Runner {
	out := make([]Runner, 0, len(runs))
	for _, run := range runs {
		out = append(out, Runner{Run: run, Subject: run.ID, Elapsed: elapsed(run, now)})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Live() != out[j].Live() {
			return out[i].Live()
		}
		return runTime(out[i].Run).After(runTime(out[j].Run))
	})
	return out
}

// Usage is the screen-facing aggregate of daemon token accounting.
type Usage struct {
	TokensIn  int64
	TokensOut int64
	CostUSD   float64
}

func (u Usage) Reported() bool { return u.TokensIn != 0 || u.TokensOut != 0 || u.CostUSD != 0 }

func TotalUsage(runners []Runner) Usage {
	var total Usage
	for _, runner := range runners {
		total.TokensIn += runner.Run.InputTokens
		total.TokensOut += runner.Run.OutputTokens
	}
	return total
}

func LiveCount(runners []Runner) int {
	count := 0
	for _, runner := range runners {
		if runner.Live() {
			count++
		}
	}
	return count
}

func elapsed(run netomatic.AgentRun, now time.Time) time.Duration {
	start := parseTime(run.StartedAt)
	if start.IsZero() {
		start = parseTime(run.FinishedAt)
	}
	if start.IsZero() {
		return 0
	}
	end := now
	if finished := parseTime(run.FinishedAt); !finished.IsZero() {
		end = finished
	}
	if end.Before(start) {
		return 0
	}
	return end.Sub(start)
}

func runTime(run netomatic.AgentRun) time.Time {
	if started := parseTime(run.StartedAt); !started.IsZero() {
		return started
	}
	return parseTime(run.FinishedAt)
}

func parseTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02 15:04:05Z07:00"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed
		}
	}
	return time.Time{}
}
