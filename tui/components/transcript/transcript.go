// Package transcript renders structured agent output for the run screens.
package transcript

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/tinker-works/goggles/tui/theme"
)

// Kind identifies the public presentation kind of a transcript entry.
type Kind string

const (
	Text                 Kind = "text"
	ToolUse              Kind = "tool_use"
	ToolOutput           Kind = "tool_output"
	Reasoning            Kind = "reasoning"
	Error                Kind = "error"
	Unknown              Kind = "unknown"
	TranscriptText            = Text
	TranscriptToolUse         = ToolUse
	TranscriptToolOutput      = ToolOutput
	TranscriptReasoning       = Reasoning
	TranscriptError           = Error
	TranscriptUnknown         = Unknown
)

// TranscriptKind and TranscriptEntry are the screen-facing names used by
// callers that do not need to know the implementation's shorter names.
type TranscriptKind = Kind

type Entry struct {
	Kind   Kind
	Tool   string
	CallID string
	Text   string
}

type TranscriptEntry = Entry

const OutputLines = 10

// Line is one unstyled display row.
type Line struct {
	Kind Kind
	Text string
}

// Render lays out entries as styled rows no wider than width.
func Render(th theme.Theme, entries []Entry, width int) []string {
	lines := Lines(entries, width)
	rendered := make([]string, 0, len(lines))
	for _, line := range lines {
		rendered = append(rendered, style(th, line.Kind).Render(line.Text))
	}
	return rendered
}

func style(th theme.Theme, kind Kind) lipgloss.Style {
	switch kind {
	case ToolUse:
		return th.Tool
	case ToolOutput, Reasoning:
		return th.Muted
	case Error:
		return th.Error
	case Unknown:
		return th.Warning
	default:
		return th.Body
	}
}

// Lines wraps and caps entries without styling them.
func Lines(entries []Entry, width int) []Line {
	if width <= 0 {
		return nil
	}
	var lines []Line
	for _, entry := range entries {
		lines = append(lines, entryLines(entry, width)...)
	}
	return lines
}

func entryLines(entry Entry, width int) []Line {
	switch entry.Kind {
	case ToolUse:
		return wrap(entry.Kind, toolHeader(entry), width)
	case ToolOutput:
		return outputLines(entry, width)
	case Reasoning:
		return wrap(entry.Kind, "thinking  "+entry.Text, width)
	case Error:
		return wrap(entry.Kind, "! "+entry.Text, width)
	case Unknown:
		return wrap(entry.Kind, "unknown json output", width)
	default:
		return wrap(entry.Kind, entry.Text, width)
	}
}

func toolHeader(entry Entry) string {
	name := entry.Tool
	if name == "" {
		name = "tool"
	}
	header := "→ " + name
	if summary := strings.TrimSpace(entry.Text); summary != "" {
		header += "  " + summary
	}
	return header
}

func outputLines(entry Entry, width int) []Line {
	source := strings.Split(strings.TrimRight(entry.Text, "\n"), "\n")
	kept := source
	if len(kept) > OutputLines {
		kept = kept[:OutputLines]
	}
	var lines []Line
	for _, line := range kept {
		lines = append(lines, wrap(entry.Kind, "  "+line, width)...)
	}
	if withheld := len(source) - len(kept); withheld > 0 {
		lines = append(lines, wrap(entry.Kind,
			fmt.Sprintf("  … %s not shown", plural(withheld, "line")), width)...)
	}
	return lines
}

func wrap(kind Kind, body string, width int) []Line {
	var lines []Line
	for _, paragraph := range strings.Split(strings.TrimRight(body, "\n"), "\n") {
		if strings.TrimSpace(paragraph) == "" {
			lines = append(lines, Line{Kind: kind})
			continue
		}
		for _, row := range strings.Split(ansi.Wrap(paragraph, width, ""), "\n") {
			lines = append(lines, Line{Kind: kind, Text: row})
		}
	}
	return lines
}

func plural(count int, noun string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, noun)
	}
	return fmt.Sprintf("%d %ss", count, noun)
}
