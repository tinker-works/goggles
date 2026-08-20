package transcript

import (
	"strings"
	"testing"
)

func TestLines_ShouldCapToolOutputBySourceLines(t *testing.T) {
	entries := []Entry{{Kind: ToolOutput, Text: strings.Repeat("line\n", OutputLines+2)}}
	lines := Lines(entries, 80)
	if len(lines) != OutputLines+1 || !strings.Contains(lines[len(lines)-1].Text, "2 lines not shown") {
		t.Fatalf("unexpected capped transcript: %+v", lines)
	}
}

func TestLines_ShouldMarkUnknownOutput(t *testing.T) {
	lines := Lines([]Entry{{Kind: Unknown}}, 40)
	if len(lines) != 1 || lines[0].Text != "unknown json output" {
		t.Fatalf("unexpected unknown transcript: %+v", lines)
	}
}
