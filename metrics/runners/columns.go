package runners

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/3xcellent/github-metrics/client"
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

// NewColumnsRunner - returns metric runner for running the columns metric, requires a project id and client
func NewColumnsRunner(metricsCfg config.RunConfig, client client.Client) ColumnsRunner {
	m := ColumnsRunner{
		Runner: NewBaseRunner(metricsCfg, client),
		Cols:   metrics.NewDateColumnMap(metricsCfg.StartDate, metricsCfg.EndDate),
	}

	return m
}

// RunName - returns formatted filename including the .csv extension
func (r *ColumnsRunner) RunName() string {
	return fmt.Sprintf("%s_%s_%d-%02d.csv",
		strings.Replace(r.ProjectName, " ", "_", -1),
		"columns",
		r.StartDate.Year(),
		r.StartDate.Month(),
	)
}

// Values - returns CSV data (rows and cols) as two-deimensional slice [][]string
// * headers with be included unless ColumnsRunner.NoHeaders is true
func (r *ColumnsRunner) Values() [][]string {
	rows := make([][]string, 0)

	if !r.NoHeaders {
		logrus.Debugf("option: headers")
		headers := []string{"Day"}
		headers = append(headers, r.ColumnNames...)
		rows = append(rows, headers)
	}

	for currentDate := r.StartDate; currentDate.Before(r.EndDate); currentDate = currentDate.AddDate(0, 0, 1) {
		r.Log(fmt.Sprintf("date: %s | %d -> %d", metrics.DateKey(currentDate), r.StartColumnIndex, r.EndColumnIndex))
		dateRow := []string{metrics.DateKey(currentDate)}
		for i := r.StartColumnIndex; i <= r.EndColumnIndex; i++ {
			appendVal := "0"
			val, found := r.Cols.DateColumn(currentDate, r.ColumnNames[i-(r.EndColumnIndex-r.StartColumnIndex)])
			if found {
				appendVal = strconv.Itoa(val)
			}

			dateRow = append(dateRow, appendVal)
		}
		rows = append(rows, dateRow)
	}
	r.Log(fmt.Sprintf("\treturning RowValues: %s", rows))
	return rows
}

// Run - Runs Columns Mwtric (gathers data from github and processes repos, issues, and events)
func (r *ColumnsRunner) Run(ctx context.Context) error {
	var issues metrics.Issues

	project, err := r.Client.GetProject(ctx, r.ProjectID)
	if err != nil {
		return err
	}
	r.ProjectName = project.Name

	projectColumns, err := r.Client.GetProjectColumns(ctx, r.ProjectID)
	if err != nil {
		return err
	}

	err = r.setColumnParams(projectColumns)
	if err != nil {
		return err
	}

	logrus.Debugf("getting repos: %#v", r)
	repos, err := r.Client.GetReposFromProjectColumn(ctx, r.EndColumnID)
	if err != nil {
		return err
	}
	r.Log(fmt.Sprintf("-----  Found Repos: %s", strings.Join(repos, ",")))

	repoIssues := r.Client.GetIssues(ctx, r.Owner, repos, r.StartDate, r.EndDate)
	r.Log(fmt.Sprintf("-----  Found repoIssues: %d", len(repoIssues)))

	for _, repoIssue := range repoIssues {
		issue := metrics.Issue{
			Issue:     &repoIssue,
			Type:      metrics.Type(repoIssue.Labels),
			IsFeature: metrics.HasFeatureLabel(repoIssue.Labels),
		}

		r.Log(fmt.Sprintf("Getting events for issue: %s/%d", issue.RepoName, issue.Number))
		events, err := r.Client.GetIssueEvents(ctx, issue.Owner, issue.RepoName, issue.Number)
		r.Log(fmt.Sprintf("\tevents found: %d", len(events)))
		if err != nil {
			r.Log(fmt.Sprintf("error getting issue: %v", err))
			continue
		}

		r.processIssueEvents(events)
		issues = append(issues, issue)
	}

	if r.after != nil {
		r.after(r.Values())
	}
	return nil
}

func (r *ColumnsRunner) processIssueEvents(events models.IssueEvents) {
	r.Log(fmt.Sprintf("ColumnsRunner.ProcessEvents: %d", len(events)))
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
			r.Log(fmt.Sprintf("Event @ %s: created %d - %s", event.CreatedAt.String(), event.ProjectID, event.ColumnName))
			fallthrough // Must fallthrough to MovedColumns for handling of case where card is dropped into column, and has not moved; expecting GetColumnName to be set
		case models.MovedColumns:
			eventDate := time.Date(createdAt.Year(), createdAt.Month(), createdAt.Day(), 0, 0, 0, 0, createdAt.Location())
			r.Log(fmt.Sprintf("Event @ %s: setting column to \"%s\"", metrics.DateKey(eventDate), event.ColumnName))

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
			r.Log(fmt.Sprintf("Event @ %s: labeled %s", event.CreatedAt.String(), event.Label))
		case models.Unlabeled:
			r.Log(fmt.Sprintf("Event @ %s: unlabeled %s", event.CreatedAt.String(), event.Label))
		default:
			r.Log(fmt.Sprintf("Event @ %s: \"%s\"", event.CreatedAt.String(), event.Event))
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
				r.Log(fmt.Sprintf("dateMap updates : [%s][%s]: %d", dateKey, colKey, r.Cols[dateKey][colKey]))
			}
		}
	}
}
