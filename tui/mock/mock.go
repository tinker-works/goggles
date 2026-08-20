// Package mock contains deterministic presentation fixtures that have no
// corresponding daemon operation yet.
package mock

import "github.com/tinker-works/donsy/netomatic"

type ProposalChange struct {
	Marker string
	Title  string
}

// Proposal returns a deterministic placeholder diff for an epic's child issues.
func Proposal(epic netomatic.Epic) []ProposalChange {
	changes := []ProposalChange{}
	for i, issue := range epic.Issues {
		if issue.ParentID == "" {
			continue
		}
		marker := "~edited"
		if i%3 == 0 {
			marker = "+new"
		}
		changes = append(changes, ProposalChange{Marker: marker, Title: issue.Title})
	}
	return changes
}
