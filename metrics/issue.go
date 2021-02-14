package metrics

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/3xcellent/github-metrics/models"
	"github.com/sirupsen/logrus"
)

// IssuesDateColumns - slice of IssuesDateColumn
type IssuesDateColumns []IssuesDateColumn

// IssuesDateColumn - adds date to models.ProjectColumn for IssuesRunner
type IssuesDateColumn struct {
	*models.ProjectColumn
	Date time.Time
}

//ColumnNames - returns the slice of column names
func (cols IssuesDateColumns) ColumnNames() []string {
	names := make([]string, 0)
	for _, col := range cols {
		names = append(names, col.Name)
	}
	return names
}

// Issues - slice of metrics.Issue
type Issues []Issue

// Issue - used to calculate metrics for an issue
type Issue struct {
	*models.Issue
	ProjectID        int64
	StartColumnIndex int
	EndColumnIndex   int

	Type      string
	IsFeature bool

	// IsCompleted      bool // TODO: can this be determined but card entering column?
	ColumnDates      IssuesDateColumns
	TotalTimeBlocked time.Duration
	BlockedTime      time.Duration
	DevTime          time.Duration
}

func (i *Issue) CalcDays() float64 {
	// logrus.Debugf("\t %s/%s/%d - calcuting: %s - %s", i.Owner, i.RepoName, i.Number, i.ColumnDates[i.EndColumnIndex].Date.String(), i.ColumnDates[i.StartColumnIndex].Date.String())
	return float64(i.ColumnDates[i.EndColumnIndex].Date.Sub(i.ColumnDates[i.StartColumnIndex].Date)) / float64(time.Hour) / 24
}

//CSVHeaders - returns list of colun headers
func (i *Issue) CSVHeaders() []string {
	var cols = []string{
		"Card #",
		"Team",
		"Type",
		"Description",
	}

	for idx := i.StartColumnIndex; idx <= i.EndColumnIndex; idx++ {
		cols = append(cols, i.ColumnDates[idx].Name)
	}

	return append(cols,
		"Development Days",
		"Feature?",
		"Blocked?",
		"Blocked Days")
}

// Values - returns a row of csv values for a single issue
func (i *Issue) Values() []string {
	row := []string{
		fmt.Sprint(i.Number), // 	"Card #"
		i.RepoName,           // 		"Repo"
		i.Type,               // Type
		i.Title,              // Description
	}
	for colInx := i.StartColumnIndex; colInx <= i.EndColumnIndex; colInx++ {
		row = append(row, i.ColumnDates[colInx].Date.Format("01/02/06"))
	}

	return append(row,
		fmt.Sprintf("%.1f", i.CalcDays()),
		strconv.FormatBool(i.IsFeature),
		fmt.Sprintf("%t", math.Ceil(float64(i.TotalTimeBlocked/time.Hour/24)) > 0), // was blocked over 24 hours?
		FmtDays(i.TotalTimeBlocked),                                                // time blocked over 24 hours
	)
}

func (i *Issue) ProcessLabels(labels []string) {
	for _, l := range labels {
		labelName := ToIssueLabel(l)
		switch labelName {
		case Bug:
			i.Type = "Bug"
		case TechDebt:
			i.Type = "Tech Debt"
		case Feature:
			i.IsFeature = true
		}
	}
	if i.Type == "" {
		i.Type = "Enhancement"
	}
}

func Type(labels []string) string {
	for _, l := range labels {
		labelName := ToIssueLabel(l)
		switch labelName {
		case Bug:
			return "Bug"
		case TechDebt:
			return "Tech Debt"
		}
	}
	return ""
}

func HasFeatureLabel(labels []string) bool {
	for _, l := range labels {
		labelName := ToIssueLabel(l)
		switch labelName {
		case Feature:
			return true
		}
	}
	return false
}

// ProcessIssueEvents sets column dates based on its events
func (i *Issue) ProcessIssueEvents() {
	logrus.Debugf("Events: %s/%s/%d - %s", i.Owner, i.RepoName, i.Number, i.Title)
	if len(i.Events) == 0 {
		return
	}
	i.setColumnDates()
	i.setEmptyColumnDates()
}

func (i *Issue) setColumnDates() {
	var blockedAt time.Time
	// var startColumn = i.ColumnDates[i.StartColumnIndex]
	initTime := time.Time{}

	for idx, event := range i.Events {
		eventNum := idx
		logPrefix := fmt.Sprintf("  [%d]@%s | %s", eventNum, event.CreatedAt.String(), event.Type)
		switch event.Type {
		case models.AddedToProject:
			logrus.Debugf("%s: %d", logPrefix, event.ProjectID)
			i.ProjectID = event.ProjectID
			logrus.Debugf("\t * added to projectID: %d", event.ProjectID)
			fallthrough // Must fallthrough to MovedColumns for handling of case where card is dropped into column, and has not moved; expecting GetColumnName to be set
		case models.MovedColumns:
			logrus.Debugf("%s - moved columns: %q -> %q\n", logPrefix, event.PreviousColumnName, event.ColumnName)

			movedToColumn, err := i.getColumn(event.ColumnName)
			if err != nil {
				logrus.Warnf("error getting column: %s\n", err.Error())
				continue
			}

			if event.PreviousColumnName != "" {
				movedFromColumn, err := i.getColumn(event.PreviousColumnName)
				if err != nil {
					logrus.Warnf("error getting previous column: %s\n", err.Error())
					continue
				}
				if movedFromColumn.Index > movedToColumn.Index {
					logrus.Debugf("\t * card moved back %d columns, leaving date: %s", movedFromColumn.Index-movedToColumn.Index, i.ColumnDates[movedToColumn.Index].Date.String())
					continue
				}
			}

			logrus.Debugf("%s - setting column %q date - %s\n", logPrefix, movedToColumn.Name, event.CreatedAt)
			i.ColumnDates[movedToColumn.Index].Date = event.CreatedAt
			logrus.Debugf("---- verifying i.DateColumns[%d].Date: %s", movedToColumn.Index, i.ColumnDates[movedToColumn.Index].Date.String())

		case models.Labeled:
			logrus.Debugf("%s: %q", logPrefix, event.Label)
			cardStatus := ToIssueLabel(event.Label)
			switch cardStatus {
			case Blocked:
				logrus.Debugf("%s: %q", logPrefix, event.Label)
				if len(i.ColumnDates) < i.StartColumnIndex+1 {
					logrus.Warnf("issue: %#v\n", i)
					logrus.Fatalf("i.IssuesDateColumns does not contain item at index %d\n", i.StartColumnIndex)
				}
				if i.ColumnDates[i.StartColumnIndex].Date != initTime {
					blockedAt = event.CreatedAt
					logrus.Debug("\t * blocked")
				} else {
					logrus.Debug("\t * blocked but not in develop yet")
				}
			default:
			}
		case models.Unlabeled:
			logrus.Debugf("%s: removed %q", logPrefix, event.Label)
			issueStatus := ToIssueLabel(event.Label)
			switch issueStatus {
			case Blocked:
				if blockedAt.After(i.ColumnDates[i.StartColumnIndex].Date) {
					logrus.Debugf("\t * unblocked")
					i.TotalTimeBlocked += event.CreatedAt.Sub(blockedAt)
				} else {
					logrus.Debug("\t * unblocked, ignoring because card was blocked before in development")
				}

				blockedAt = time.Time{}
			default:
			}
		case models.Assigned:
			logrus.Debugf("%s: %q", logPrefix, event.Assignee)
		case models.Unassigned:
			logrus.Debugf("%s: %q", logPrefix, event.Assignee)
		case models.Mentioned:
			logrus.Debugf("%s: %q - %q", logPrefix, event.LoginName, event.Note)
		case models.Closed:
			logrus.Debugf("%s", logPrefix)
		default:
			logrus.Debugf("%s: unrecognized event", logPrefix)
		}
	}
}
func (i *Issue) setEmptyColumnDates() {
	// TODO: handle weekends?
	// https://stackoverflow.com/questions/31327124/how-can-i-exclude-weekends-golang

	// this section attempts to fix missing dates in ColumnsMetric
	// by iterating backwards through the ColumnsMetric and
	// adjusting missing dates to previous column if not set
	for dateIdx := len(i.ColumnDates) - 1; dateIdx >= 0; dateIdx-- {
		// if date was never set, set to next date we know something happened
		logrus.Debugf("idx %d - column: %s - Date: %s", dateIdx, i.ColumnDates[dateIdx].Name, i.ColumnDates[dateIdx].Date.String())
		if i.ColumnDates[dateIdx].Date.IsZero() {
			logrus.Debugf("\t\tDate.IsZero()")
			// if last
			if dateIdx == i.EndColumnIndex {
				i.ColumnDates[dateIdx].Date = i.Events[len(i.Events)-1].CreatedAt // get date from last event
				logrus.Debugf("\t\tsetting to last event date: %s", i.Events[len(i.Events)-1].CreatedAt.String())
			} else if dateIdx < i.EndColumnIndex && dateIdx > i.StartColumnIndex {
				i.ColumnDates[dateIdx].Date = i.ColumnDates[dateIdx+1].Date // get date from date column just set
				logrus.Debugf("\t\tsetting to next column date: %s", i.ColumnDates[dateIdx+1].Date.String())
			} else if dateIdx == i.StartColumnIndex {
				i.ColumnDates[dateIdx].Date = i.Events[0].CreatedAt // get date from date column just set
				logrus.Debugf("\t\tsetting to first event date: %s", i.Events[0].CreatedAt.String())
			}
		}
	}
	for _, date := range i.ColumnDates {
		logrus.Debugf("\t\t\t%s - %s", date.Name, date.Date.String())
	}
}

func (i *Issue) getColumn(name string) (*IssuesDateColumn, error) {
	if len(i.ColumnDates) == 0 {
		return nil, errors.New("ColumnDates is empty")
	}
	lookingFor := strings.ToUpper(name)
	for _, col := range i.ColumnDates {
		if strings.ToUpper(col.Name) == lookingFor {
			return &col, nil
		}
	}
	return nil, errors.New("column not found: " + name)
}
