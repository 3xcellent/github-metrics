package runners

import (
	"context"
	"strconv"
	"time"

	"github.com/3xcellent/github-metrics/config"
	"github.com/3xcellent/github-metrics/metrics"
	"github.com/3xcellent/github-metrics/models"
	"github.com/sirupsen/logrus"
)

// ColumnsRunner - contains all data needed to run and maintain state for the Columns Metric
type ColumnsRunner struct {
	*Runner
	Cols metrics.DateColMap
}

var _ MetricsRunner = new(ColumnsRunner)

// NewColumnsRunner - returns metric runner for running the columns metric, requires a project id and client
func NewColumnsRunner(metricsCfg config.RunConfig, client Client) *ColumnsRunner {
	m := ColumnsRunner{
		Runner: NewBaseRunner(metricsCfg, client),
		Cols:   metrics.NewDateColumnMap(metricsCfg.StartDate, metricsCfg.EndDate),
	}

	return &m
}

// Headers returns list of headers column names
func (r *ColumnsRunner) Headers() []string {
	headers := []string{"Date"}
	headers = append(headers, r.ColumnNames...)
	return headers
}

// Values - returns CSV data (rows and cols) as two-deimensional slice [][]string
// * headers with be included unless ColumnsRunner.NoHeaders is true
func (r *ColumnsRunner) Values() [][]string {
	rows := make([][]string, 0)

	if !r.NoHeaders {
		logrus.Debugf("option: headers")
		rows = append(rows, r.Headers())
	}

	for currentDate := r.StartDate; currentDate.Before(r.EndDate); currentDate = currentDate.AddDate(0, 0, 1) {
		logrus.Debugf("date: %s | %d -> %d", metrics.DateKey(currentDate), r.StartColumnIndex, r.EndColumnIndex)
		dateRow := []string{metrics.DateKey(currentDate)}
		for i := r.StartColumnIndex; i <= r.EndColumnIndex; i++ {
			appendVal := "0"
			val, found := r.Cols.DateColumn(currentDate, r.ColumnNames[i])
			if found {
				appendVal = strconv.Itoa(val)
			}

			dateRow = append(dateRow, appendVal)
		}
		rows = append(rows, dateRow)
	}
	logrus.Debugf("\treturning RowValues: %s", rows)
	return rows
}

// Run - Runs Columns Mwtric (gathers data from github and processes repos, issues, and events)
func (r *ColumnsRunner) Run(ctx context.Context) error {
	logrus.Debug("Starting ColumnsRunner")
	r.Debug()
	ghIssues, _, err := r.GetIssuesAndColumns(ctx)
	if err != nil {
		return err
	}

	var issues metrics.Issues
	for _, ghIssue := range ghIssues {
		issue := metrics.Issue{
			Issue:     &ghIssue,
			Type:      metrics.Type(ghIssue.Labels),
			IsFeature: metrics.HasFeatureLabel(ghIssue.Labels),
		}

		logrus.Debugf("getting events for issue: %s/%d", issue.RepoName, issue.Number)
		events, err := r.Client.GetIssueEvents(ctx, issue.Owner, issue.RepoName, issue.Number)
		logrus.Debugf("\tevents found: %d", len(events))
		if err != nil {
			return err
		}

		r.processIssueEvents(events)
		issues = append(issues, issue)
	}
	if r.after != nil {
		err = r.after(r.Values())
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ColumnsRunner) processIssueEvents(events models.IssueEvents) {
	logrus.Debugf("ColumnsRunner.ProcessEvents: %d", len(events))
	shouldIncludeData := false
	issueDateMap := metrics.DateColMap{}

	var prevDate time.Time
	var prevColumn string

	for eventIdx, event := range events {
		createdAt := event.CreatedAt
		switch event.Type {
		case models.AddedToProject:
			shouldIncludeData = event.ProjectID == r.ProjectID

			if event.ColumnName == "" {
				logrus.Warn("ProjectCard ColumnName not set.")
				continue
			}
			logrus.Debugf("Event @ %s: created %d - %s", event.CreatedAt.String(), event.ProjectID, event.ColumnName)
			fallthrough // Must fallthrough to MovedColumns for handling of case where card is dropped into column, and has not moved; expecting GetColumnName to be set
		case models.MovedColumns:
			eventDate := time.Date(createdAt.Year(), createdAt.Month(), createdAt.Day(), 0, 0, 0, 0, createdAt.Location())
			logrus.Debugf("Event @ %s: setting column to \"%s\"", metrics.DateKey(eventDate), event.ColumnName)

			// if prevDate was set, fill in dates between the prevDate and this date.
			if !prevDate.IsZero() && prevDate.Before(eventDate) {
				fillDate := prevDate
				for fillDate.Before(eventDate) {
					issueDateMap[metrics.DateKey(fillDate)] = map[string]int{prevColumn: 1}
					fillDate = fillDate.AddDate(0, 0, 1)
				}
			}

			issueDateMap[metrics.DateKey(eventDate)] = map[string]int{event.ColumnName: 1}

			// set for next MovedColumns event to check
			prevDate = eventDate
			prevColumn = event.ColumnName
		case models.Labeled:
			logrus.Debugf("Event @ %s: labeled %s", event.CreatedAt.String(), event.Label)
		case models.Unlabeled:
			logrus.Debugf("Event @ %s: unlabeled %s", event.CreatedAt.String(), event.Label)
		default:
			logrus.Debugf("Event @ %s: \"%s\"", event.CreatedAt.String(), event.Event)
		}

		// account for issues not done yet by 'filling-in' date ColumnsRunner until the endDate
		if eventIdx == len(events)-1 && prevColumn != r.EndColumn && !prevDate.IsZero() {
			fillDate := prevDate.AddDate(0, 0, 1)
			for fillDate.Before(r.EndDate) {
				issueDateMap[metrics.DateKey(fillDate)] = map[string]int{prevColumn: 1}
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
				if _, found := r.Cols[dateKey]; !found {
					r.Cols[dateKey] = map[string]int{}
				}
				r.Cols[dateKey][colKey] += colVal
				logrus.Debugf("dateMap updates : [%s][%s]: %d", dateKey, colKey, r.Cols[dateKey][colKey])
			}
		}
	}
}
