package models

import (
	"strings"
	"time"
)

// IssueEvent - model for github issue events
type IssueEvent struct {
	Event              string
	CreatedAt          time.Time
	ProjectID          int64
	Type               IssueEventType
	ColumnName         string
	PreviousColumnName string
	Label              string
	Assignee           string
	Note               string
	LoginName          string
}

// IssueEvents - slice of IssueEvent
type IssueEvents []IssueEvent

// IssueEventType - allows assign of event labels
type IssueEventType string

// labels
const (
	Assigned       IssueEventType = "ASSIGNED"
	Unassigned     IssueEventType = "UNASSIGNED"
	Mentioned      IssueEventType = "MENTIONED"
	Closed         IssueEventType = "CLOSED"
	MovedColumns   IssueEventType = "MOVED_COLUMNS_IN_PROJECT"
	Labeled        IssueEventType = "LABELED"
	Unlabeled      IssueEventType = "UNLABELED"
	AddedToProject IssueEventType = "ADDED_TO_PROJECT"
)

// ToIssueEventType - returns IssueEventType from string
func ToIssueEventType(s string) IssueEventType {
	return IssueEventType(strings.ToUpper(s))
}
