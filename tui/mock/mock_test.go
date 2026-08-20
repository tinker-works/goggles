package mock

import (
	"testing"

	"github.com/tinker-works/donsy/netomatic"
)

func TestProposal_ShouldConsumePublicEpicAndSkipRoot(t *testing.T) {
	changes := Proposal(netomatic.Epic{Issues: []netomatic.Issue{{ID: "root"}, {ID: "child", ParentID: "root", Title: "Child"}}})
	if len(changes) != 1 || changes[0].Title != "Child" || changes[0].Marker == "" {
		t.Fatalf("unexpected proposal: %+v", changes)
	}
}
