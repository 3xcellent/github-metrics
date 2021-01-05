package metrics

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/3xcellent/github-metrics/config"
	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
)

const DateKeyFmt = "2006-01-02"

type ColumnsMetric struct {
	Cols             ColsMap
	BoardID          int64
	Owner            string
	StartColumn      string
	StartDate        time.Time
	StartColumnIndex int
	EndDate          time.Time
	EndColumn        string
	EndColumnIndex   int
	EndColumnID      int64
	ColumnNames      []string
	LogFunc          func(args ...interface{})

	RepoCount  int
	IssueCount int
	EventCount int
}

func NewColumnsMetric(boardCfg config.BoardConfig, projectColumns []*github.ProjectColumn, logFunc func(args ...interface{})) ColumnsMetric {
	logrus.Infof("NewColumnsMetric-boardCfg.Owner: %s", boardCfg.Owner)
	if boardCfg.Owner == "" {
		panic("wtf")
	}
	metric := ColumnsMetric{
		Cols:        NewDateColumnMap(boardCfg.StartDate, boardCfg.EndDate),
		BoardID:     boardCfg.BoardID,
		Owner:       boardCfg.Owner,
		StartDate:   boardCfg.StartDate,
		EndDate:     boardCfg.EndDate,
		StartColumn: boardCfg.StartColumn,
		EndColumn:   boardCfg.EndColumn,
		LogFunc:     logFunc,
	}
	for _, col := range projectColumns {
		metric.ColumnNames = append(metric.ColumnNames, col.GetName())
	}
	metric.Log(fmt.Sprintf("Columns: %s", strings.Join(metric.ColumnNames, ", ")))

	for i, col := range projectColumns {
		colName := col.GetName()
		if colName == boardCfg.StartColumn {
			metric.Log(fmt.Sprintf("StartColumnIndex: %d", i))
			metric.StartColumnIndex = i
		}

		if colName == boardCfg.EndColumn {
			metric.Log(fmt.Sprintf("EndColumnIndex: %d", i))
			metric.EndColumnIndex = i
		}
	}
	if metric.EndColumnIndex == 0 {
		metric.EndColumnIndex = len(projectColumns) - 1
	}
	metric.EndColumnID = projectColumns[metric.EndColumnIndex].GetID()

	return metric
}

func (m *ColumnsMetric) Log(values ...interface{}) {
	for _, value := range values {
		m.LogFunc(value.(string))
	}
}

func (m *ColumnsMetric) RowValues() [][]string {
	rows := make([][]string, 0)
	for currentDate := m.StartDate; currentDate.Before(m.EndDate); currentDate = currentDate.AddDate(0, 0, 1) {
		m.Log(fmt.Sprintf("date: %s | %d -> %d", currentDate.Format(DateKeyFmt), m.StartColumnIndex, m.EndColumnIndex))
		dateRow := []string{currentDate.Format(DateKeyFmt)}
		for i := m.StartColumnIndex; i <= m.EndColumnIndex; i++ {
			appendVal := "0"
			val, found := m.DateColumn(currentDate, m.ColumnNames[i])
			if found {
				appendVal = strconv.Itoa(val)
			}

			dateRow = append(dateRow, appendVal)
		}
		rows = append(rows, dateRow)
	}
	m.Log(fmt.Sprintf("\treturning RowValues: %s", rows))
	return rows
}

func (m *ColumnsMetric) GatherAndProcessIssues(ctx context.Context, client config.MetricsClient) {
	var issues Issues
	repos := client.GetRepos(ctx, m.EndColumnID)
	m.Log(fmt.Sprintf("-----  Found Repos: %s", strings.Join(repos, ",")))
	repoIssues := client.GetIssues(ctx, m.Owner, repos, m.StartDate, m.EndDate)
	m.Log(fmt.Sprintf("-----  Found repoIssues: %d", len(repoIssues)))
	for _, repoIssue := range repoIssues {
		issue := &Issue{
			Owner:     m.Owner,
			RepoName:  repoIssue.GetRepository().GetName(),
			Number:    repoIssue.GetNumber(),
			Title:     repoIssue.GetTitle(),
			Type:      Type(repoIssue.Labels),
			IsFeature: HasFeatureLabel(repoIssue.Labels),
		}

		m.Log(fmt.Sprintf("Getting events for issue: %s/%d", issue.RepoName, issue.Number))
		events, err := client.GetIssueEvents(ctx, issue.Owner, issue.RepoName, issue.Number)
		m.Log(fmt.Sprintf("\tevents found: %d", len(events)))
		if err != nil {
			m.Log(fmt.Sprintf("error getting issue: %v", err))
			continue
		}

		m.ProcessEvents(events)
		issues = append(issues, issue)
	}
}

type ColsMap map[string]map[string]int

func (m *ColumnsMetric) ProcessEvents(events []*github.IssueEvent) {
	m.Log(fmt.Sprintf("ColumnsMetric.ProcessEvents: %d", len(events)))
	shouldIncludeData := false
	issueDateMap := ColsMap{}

	var prevDate time.Time
	var prevColumn string

	for eventIdx, event := range events {
		eventType := ToIssueEvent(event.GetEvent())
		createdAt := event.GetCreatedAt().Local()
		switch eventType {
		case AddedToProject:
			shouldIncludeData = event.GetProjectCard().GetProjectID() == m.BoardID

			if event.GetProjectCard().GetColumnName() == "" {
				logrus.Warn("ProjectCard ColumnName not set.")
				continue
			}
			m.Log(fmt.Sprintf("Event @ %s: created %d - %s", event.GetCreatedAt().String(), event.GetProjectCard().GetProjectID(), event.GetProjectCard().GetColumnName()))
			fallthrough // Must fallthrough to MovedColumns for handling of case where card is dropped into column, and has not moved; expecting GetColumnName to be set
		case MovedColumns:
			eventDate := time.Date(createdAt.Year(), createdAt.Month(), createdAt.Day(), 0, 0, 0, 0, createdAt.Location())
			m.Log(fmt.Sprintf("Event @ %s: setting column to \"%s\"", eventDate.Format(fmtDateKey), event.GetProjectCard().GetColumnName()))

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
		//		m.Log(fmt.Sprintf("Event @ %s: blocked", event.GetCreatedAt().String()))
		//	default:
		//		m.Log(fmt.Sprintf("Event @ %s: labeled %s", event.GetCreatedAt().String(), cardStatus))
		//	}
		//case Unlabeled:
		//	cardStatus := ToIssueLabel(event.GetLabel().GetName())
		//	switch cardStatus {
		//	case Blocked:
		//		//m.Log(fmt.Sprintf("Event @ %s: unblocked", event.GetCreatedAt().String()))
		//	default:
		//		//m.Log(fmt.Sprintf("Event @ %s: unlabeled %s", event.GetCreatedAt().String(), cardStatus))
		//	}
		default:
			m.Log(fmt.Sprintf("Event @ %s: \"%s\"", event.GetCreatedAt().String(), event.GetEvent()))
		}

		// account for issues not done yet by 'filling-in' date ColumnsMetric until the endDate
		if eventIdx == len(events)-1 && prevColumn != m.EndColumn && !prevDate.IsZero() {
			fillDate := prevDate.AddDate(0, 0, 1)
			for fillDate.Before(m.EndDate) {
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
				if _, found := m.Cols[dateKey]; !found {
					m.Cols[dateKey] = map[string]int{}
				}
				m.Cols[dateKey][colKey] += colVal
				m.Log(fmt.Sprintf("dateMap updates : [%s][%s]: %d", dateKey, colKey, m.Cols[dateKey][colKey]))
			}
		}
	}
}

func (m *ColumnsMetric) DateColumn(date time.Time, columnName string) (int, bool) {
	val, found := m.Cols[date.Format(fmtDateKey)][columnName]
	return val, found
}

func (m *ColumnsMetric) Dump() string {
	lines := make([]string, 0)
	keys := make([]string, 0, len(m.Cols))
	for k := range m.Cols {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		lines = append(lines, fmt.Sprintf("%s: %#v", key, m.Cols[key]))
	}
	return strings.Join(lines, "\n")
}

func NewDateColumnMap(beginDate, endDate time.Time) ColsMap {
	current := time.Date(beginDate.Year(), beginDate.Month(), beginDate.Day(), 0, 0, 0, 0, beginDate.Location())
	dateMap := ColsMap{}
	for current.Before(endDate) {
		dateMap[current.Format(fmtDateKey)] = map[string]int{}
		current = current.AddDate(0, 0, 1)
	}
	return dateMap
}
