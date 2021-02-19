package metrics

import "strings"

// IssueLabel - used to label issues
type IssueLabel string

// issue labels
const (
	Blocked  IssueLabel = "BLOCKED"
	Bug      IssueLabel = "BUG"
	TechDebt IssueLabel = "TECH DEBT"
	Feature  IssueLabel = "FEATURE"
)

// ToIssueLabel - returns the IssueLabel for a string
func ToIssueLabel(s string) IssueLabel {
	return IssueLabel(strings.ToUpper(s))
}
