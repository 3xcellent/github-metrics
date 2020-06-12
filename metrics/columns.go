package metrics

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
)

type columns struct {
	cols      Columns
	boardID   int64
	endDate   time.Time
	endColumn string
}

func NewColumns(boardID int64, beginDate, endDate time.Time, endColumn string) columns {
	return columns{
		cols:      newDateColumnMap(beginDate, endDate),
		boardID:   boardID,
		endDate:   endDate,
		endColumn: endColumn,
	}
}

type Columns map[string]map[string]int

func (c *columns) ProcessEvents(events []*github.IssueEvent) {
	shouldIncludeData := false
	issueDateMap := Columns{}

	var prevDate time.Time
	var prevColumn string

	for eventIdx, event := range events {
		eventType := ToIssueEvent(event.GetEvent())
		createdAt := event.GetCreatedAt().Local()
		switch eventType {
		case AddedToProject:
			shouldIncludeData = event.GetProjectCard().GetProjectID() == c.boardID

			if event.GetProjectCard().GetColumnName() == "" {
				logrus.Warn("ProjectCard ColumnName not set.")
				continue
			}
			logrus.Debugf("Event @ %s: created %d - %s", event.GetCreatedAt().String(), event.GetProjectCard().GetProjectID(), event.GetProjectCard().GetColumnName())
			fallthrough // Must fallthrough to MovedColumns for handling of case where card is dropped into column, and has not moved; expecting GetColumnName to be set
		case MovedColumns:
			eventDate := time.Date(createdAt.Year(), createdAt.Month(), createdAt.Day(), 0, 0, 0, 0, createdAt.Location())
			logrus.Debugf("Event @ %s: setting column to \"%s\"", eventDate.Format(fmtDateKey), event.GetProjectCard().GetColumnName())

			// if prevDate was set, fill in dates between the prevDate and this date.
			if !prevDate.IsZero() && prevDate.Before(eventDate) {
				fillDate := prevDate
				for fillDate.Before(eventDate) {
					issueDateMap[fillDate.Format(fmtDateKey)] = map[string]int{prevColumn: 1}
					fillDate = fillDate.AddDate(0, 0, 1)
				}
			}

			issueDateMap[eventDate.Format(fmtDateKey)] = map[string]int{event.GetProjectCard().GetColumnName(): 1}

			// set for next MovedColumns event to check
			prevDate = eventDate
			prevColumn = event.GetProjectCard().GetColumnName()
		//case Labeled:
		//	cardStatus := ToIssueLabel(event.GetLabel().GetName())
		//	switch cardStatus {
		//	case Blocked:
		//		logrus.Debugf("Event @ %s: blocked", event.GetCreatedAt().String())
		//	default:
		//		logrus.Debugf("Event @ %s: labeled %s", event.GetCreatedAt().String(), cardStatus)
		//	}
		//case Unlabeled:
		//	cardStatus := ToIssueLabel(event.GetLabel().GetName())
		//	switch cardStatus {
		//	case Blocked:
		//		//logrus.Debugf("Event @ %s: unblocked", event.GetCreatedAt().String())
		//	default:
		//		//logrus.Debugf("Event @ %s: unlabeled %s", event.GetCreatedAt().String(), cardStatus)
		//	}
		default:
			logrus.Debugf("Event @ %s: \"%s\"", event.GetCreatedAt().String(), event.GetEvent())
		}

		// account for issues not done yet by 'filling-in' date columns until the endDate
		if eventIdx == len(events)-1 && prevColumn != c.endColumn && !prevDate.IsZero() {
			fillDate := prevDate.AddDate(0, 0, 1)
			for fillDate.Before(c.endDate) {
				issueDateMap[fillDate.Format(fmtDateKey)] = map[string]int{prevColumn: 1}
				fillDate = fillDate.AddDate(0, 0, 1)
			}
		}
	}

	if len(issueDateMap) == 0 {
		return
	}
	if shouldIncludeData {

		for dateKey, dateColumns := range issueDateMap {
			for colKey, colVal := range dateColumns {
				if _, found := c.cols[dateKey]; !found {
					c.cols[dateKey] = map[string]int{}
				}
				c.cols[dateKey][colKey] += colVal
				logrus.Debugf("dateMap updates : [%s][%s]: %d", dateKey, colKey, c.cols[dateKey][colKey])
			}
		}
	}
}

func (c *columns) DateColumn(date time.Time, columnName string) (int, bool) {
	val, found := c.cols[date.Format(fmtDateKey)][columnName]
	return val, found
}

func (c *columns) Dump() string {
	lines := make([]string, 0)
	keys := make([]string, 0, len(c.cols))
	for k := range c.cols {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		lines = append(lines, fmt.Sprintf("%s: %#v", key, c.cols[key]))
	}
	return strings.Join(lines, "\n")
}

func newDateColumnMap(beginDate, endDate time.Time) Columns {
	current := time.Date(beginDate.Year(), beginDate.Month(), beginDate.Day(), 0, 0, 0, 0, beginDate.Location())
	dateMap := Columns{}
	for current.Before(endDate) {
		dateMap[current.Format(fmtDateKey)] = map[string]int{}
		current = current.AddDate(0, 0, 1)
	}
	return dateMap
}
