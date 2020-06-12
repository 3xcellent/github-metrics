package metrics

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
)

const fmtDateKey = "2006-01-02"

type Issues []*Issue
type Issue struct {
	Owner            string
	RepoName         string
	Number           int
	Type             string
	Title            string
	ColumnDates      []*BoardColumn
	DateColumnMap    Columns
	IsFeature        bool
	TotalTimeBlocked time.Duration
	ProjectID        int64
}

type IssueLabel string

const (
	Blocked  IssueLabel = "BLOCKED"
	Bug      IssueLabel = "BUG"
	TechDebt IssueLabel = "TECH DEBT"
	Feature  IssueLabel = "FEATURE"
)

func ToIssueLabel(s string) IssueLabel {
	return IssueLabel(strings.ToUpper(s))
}

type IssueEvent string

const (
	Assigned       IssueEvent = "ASSIGNED"
	Unassigned     IssueEvent = "UNASSIGNED"
	Mentioned      IssueEvent = "MENTIONED"
	Closed         IssueEvent = "CLOSED"
	MovedColumns   IssueEvent = "MOVED_COLUMNS_IN_PROJECT"
	Labeled        IssueEvent = "LABELED"
	Unlabeled      IssueEvent = "UNLABELED"
	AddedToProject IssueEvent = "ADDED_TO_PROJECT"
)

func ToIssueEvent(s string) IssueEvent {
	return IssueEvent(strings.ToUpper(s))
}

type BoardColumns []*BoardColumn

func (cd BoardColumns) ColumnNames() []string {
	names := make([]string, 0)
	for _, columnDate := range cd {
		names = append(names, columnDate.Name)
	}
	return names
}

type BoardColumn struct {
	Name string
	ID   int64
	Date time.Time
}

func (i *Issue) CalcDays(startColumnIndex, endColumnIndex int) float64 {
	return float64(i.ColumnDates[endColumnIndex].Date.Sub(i.ColumnDates[startColumnIndex].Date)) / float64(time.Hour) / 24
}

func (i *Issue) CSVHeaders(beginColumnIdx, endColumnIdx int) []string {
	var cols = []string{
		"Card #",
		"Team",
		"Type",
		"Description",
	}

	for idx := beginColumnIdx; idx <= endColumnIdx; idx++ {
		cols = append(cols, i.ColumnDates[idx].Name)
	}

	return append(cols,
		"Development Days",
		"Feature?",
		"Blocked?",
		"Blocked Days")
}

func (i *Issue) CSV(beginColumnIdx, endColumnIdx int) []string {
	row := []string{
		fmt.Sprint(i.Number),
		i.RepoName,
		i.Type,
		i.Title,
	}
	for idx := beginColumnIdx; idx <= endColumnIdx; idx++ {
		row = append(row, i.ColumnDates[idx].Date.Format("01/02/06"))
	}

	return append(row,
		fmt.Sprintf("%.1f", i.CalcDays(beginColumnIdx, endColumnIdx)),
		strconv.FormatBool(i.IsFeature),
		fmt.Sprintf("%t", math.Ceil(float64(i.TotalTimeBlocked/time.Hour/24)) > 0), // was blocked over 24 hours?
		FmtDays(i.TotalTimeBlocked),                                                // time blocked over 24 hours
	)
}

func (i *Issue) ProcessLabels(labels []*github.Label) {
	for _, l := range labels {
		labelName := ToIssueLabel(l.GetName())
		switch labelName {
		case Bug:
			i.Type = "Bug"
		case TechDebt:
			i.Type = "Tech Debt"
		case Feature:
			i.IsFeature = true
		}
	}
}

func (i *Issue) ProcessIssueEvents(events []*github.IssueEvent, beginColumnIdx int) {
	logrus.Debugf("\n\nbeginColumnIdx: %d\n\n", beginColumnIdx)

	if len(events) == 0 {
		return
	}
	var blockedAt time.Time
	initTime := time.Time{}

	for idx, event := range events {
		eventType := ToIssueEvent(event.GetEvent())
		switch eventType {
		case AddedToProject:
			i.ProjectID = event.GetProjectCard().GetProjectID()
			logrus.Debugf("Event @ %s: added to repoName \"%d\"", event.GetCreatedAt().String(), event.GetProjectCard().GetProjectID())
			fallthrough // Must fallthrough to MovedColumns for handling of case where card is dropped into column, and has not moved; expecting GetColumnName to be set
		case MovedColumns:
			logrus.Debugf("Event[%d] @ %s: moved columns %s -> %s\n", idx, event.GetCreatedAt().String(), event.GetProjectCard().GetPreviousColumnName(), event.GetProjectCard().GetColumnName())
			boardColumn, _, err := i.getColumn(event.GetProjectCard().GetColumnName())
			if err != nil {
				logrus.Debugf("Event: metric column not found: %s.   skip metric...\n", event.GetProjectCard().GetColumnName())
				return
			}
			logrus.Debugf("Event: setting column \"%s\" date - %s\n", boardColumn.Name, event.GetCreatedAt())
			boardColumn.Date = event.GetCreatedAt()
		case Labeled:
			cardStatus := ToIssueLabel(event.GetLabel().GetName())
			switch cardStatus {
			case Blocked:

				if len(i.ColumnDates) < beginColumnIdx+1 {
					logrus.Warningf("issue: %#v\n", i)
					logrus.Fatalf("i.BoardColumns does not contain item at index %d\n", beginColumnIdx)
				}
				if i.ColumnDates[beginColumnIdx].Date != initTime {
					blockedAt = event.GetCreatedAt()
					logrus.Debugf("Event @ %s: blocked", event.GetCreatedAt().String())
				} else {
					logrus.Debugf("Event @ %s: blocked but not in develop yet", event.GetCreatedAt().String())
				}

			default:
			}
		case Unlabeled:
			issueStatus := ToIssueLabel(event.GetLabel().GetName())
			switch issueStatus {
			case Blocked:
				if blockedAt.After(i.ColumnDates[beginColumnIdx].Date) {
					logrus.Debugf("Event @ %s: unblocked", event.GetCreatedAt().String())
					i.TotalTimeBlocked += event.GetCreatedAt().Sub(blockedAt)
				} else {
					logrus.Debugf("Event @ %s: unblocked, ignoring because card was blocked before in development", event.GetCreatedAt().String())
				}

				blockedAt = time.Time{}
			default:
			}
		case Assigned, Unassigned:
			logrus.Debugf("Event @ %s: assigned/unassigned to \"%s\"", event.GetCreatedAt().String(), event.GetAssignee().GetLogin())
		case Mentioned:
			logrus.Debugf("Event @ %s: Mentioned %s, \"%s\"", event.GetCreatedAt().String(), event.GetProjectCard().GetNote(), event.GetActor().GetLogin())
		case Closed:
			logrus.Debugf("Event @ %s: Closed ", event.GetCreatedAt().String())
		default:
			logrus.Debugf("Event @ %s: unrecognized event: %#v", event.GetCreatedAt().String(), eventType)
		}
	}

	// TODO: handle weekends?
	// https://stackoverflow.com/questions/31327124/how-can-i-exclude-weekends-golang

	// this section attempts to fix missing dates in columns
	// by iterating backwards through the columns and
	// adjusting missing dates to previous column if not set
	defaultTime := time.Time{}
	for dateIdx := len(i.ColumnDates) - 1; dateIdx >= 0; dateIdx-- {
		// if date was never set, set to next date we know something happened
		if i.ColumnDates[dateIdx].Date == defaultTime {
			// if last
			if dateIdx == len(i.ColumnDates)-1 {
				i.ColumnDates[dateIdx].Date = events[len(events)-1].GetCreatedAt() // get date from last event
			} else if dateIdx != 0 {
				i.ColumnDates[dateIdx].Date = i.ColumnDates[dateIdx+1].Date // get date from date column just set
			}
		}
	}
}

func (issues Issues) CSVRowColumns(startColumnIndex, endColumnIndex int, boardID int64, headers bool, startDate, endDate time.Time) [][]string {
	var rowColumns [][]string
	if headers {
		rowColumns = append(rowColumns, issues[0].CSVHeaders(startColumnIndex, endColumnIndex))
	}
	for _, issue := range issues {
		if issue.ProjectID == boardID &&
			issue.ColumnDates[endColumnIndex].Date.After(startDate) &&
			issue.ColumnDates[endColumnIndex].Date.Before(endDate) {

			if issue.CalcDays(startColumnIndex, endColumnIndex) > 0 {
				rowColumns = append(rowColumns, issue.CSV(startColumnIndex, endColumnIndex))
			}
		}
	}
	return rowColumns
}

func FmtDays(d time.Duration) string {
	d = d.Round(time.Minute)
	days := int(math.Ceil(float64(d / time.Hour / 24)))
	//h := d / time.Hour
	//d -= h * time.Hour
	//m := d / time.Minute
	return fmt.Sprintf("%d", days)
}

func FmtDaysHours(d time.Duration) string {
	d = d.Round(time.Minute)
	duration := float64(d) / float64(time.Hour) / 24
	return fmt.Sprintf("%.1f", duration)
}

func (i *Issue) getColumn(name string) (*BoardColumn, int, error) {
	lookingFor := strings.ToUpper(name)
	for index, col := range i.ColumnDates {
		if strings.ToUpper(col.Name) == lookingFor {
			return col, index, nil
		}
	}
	return nil, 0, errors.New("column not found: " + name)
}
