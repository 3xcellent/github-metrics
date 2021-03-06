package runners

import (
	"context"
	"errors"

	"github.com/3xcellent/github-metrics/config"
	"github.com/3xcellent/github-metrics/metrics"
	"github.com/3xcellent/github-metrics/models"
	"github.com/sirupsen/logrus"
)

// IssuesRunner - contains all data needed to run and maintain state for the Issues Metric
type IssuesRunner struct {
	*Runner
	IssueNumber int
	RepoName    string
	Issues      metrics.Issues
}

// NewIssuesRunner - returns metric runner for running the columns metric, requires a project id and client
func NewIssuesRunner(metricsCfg config.RunConfig, client Client) *IssuesRunner {
	m := IssuesRunner{
		Runner: NewBaseRunner(metricsCfg, client),
	}

	return &m
}

// Values - returns CSV data (rows and cols) as two-deimensional slice [][]string
// * headers with be included unless ColumnsRunner.NoHeaders is true
func (r *IssuesRunner) Values() [][]string {
	// logrus.Debug("IssuesRunner.Values: %#v", r.Issues)

	var rowColumns [][]string
	if !r.NoHeaders && len(r.Issues) > 0 {
		rowColumns = append(rowColumns, r.Issues[0].CSVHeaders())
	}
	for _, issue := range r.Issues {
		if issue.ProjectID == r.ProjectID &&
			issue.ColumnDates[r.EndColumnIndex].Date.After(r.StartDate) &&
			issue.ColumnDates[r.EndColumnIndex].Date.Before(r.EndDate) {

			if issue.CalcDays() > 0.01 {
				issueValues := issue.Values()
				rowColumns = append(rowColumns, issueValues)
			}
		}
	}
	return rowColumns
}

// Run - Runs Columns Mwtric (gathers data from github and processes repos, issues, and events)
func (r *IssuesRunner) Run(ctx context.Context) error {
	logrus.Debug("Starting IssuesRunner")
	r.Debug()

	ghIssues, projectColumns, err := r.GetIssuesAndColumns(ctx)
	if err != nil {
		return err
	}

	dateCols := make(metrics.IssuesDateColumns, 0)
	for _, col := range projectColumns {
		dateCols = append(dateCols, metrics.IssuesDateColumn{ProjectColumn: &models.ProjectColumn{Name: col.Name, ID: col.ID}})
	}

	logrus.Debugf("dateCols: %#v", dateCols)
	for _, ghIssue := range ghIssues {
		metricsIssue, err := r.newMetricsIssue(ctx, ghIssue, dateCols)
		if err != nil {
			return err
		}
		metricsIssue.ProcessIssueEvents()
		r.Issues = append(r.Issues, metricsIssue)
	}

	if r.after != nil {
		err = r.after(r.Values())
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *IssuesRunner) newMetricsIssue(ctx context.Context, ghIssue models.Issue, dateColumns metrics.IssuesDateColumns) (metrics.Issue, error) {
	issue := metrics.Issue{
		ProjectID:        r.ProjectID,
		Issue:            &ghIssue,
		StartColumnIndex: r.StartColumnIndex,
		EndColumnIndex:   r.EndColumnIndex,
	}
	issue.ProcessLabels(ghIssue.Labels)
	dates, err := newDateColumns(dateColumns)
	if err != nil {
		return metrics.Issue{}, err
	}
	issue.ColumnDates = dates

	return issue, nil
}

func newDateColumns(dateColumns metrics.IssuesDateColumns) (metrics.IssuesDateColumns, error) {
	if len(dateColumns) == 0 {
		return nil, errors.New("dateColumns cannot be empty")
	}
	newDateColumns := make(metrics.IssuesDateColumns, 0, len(dateColumns))
	for _, dc := range dateColumns {
		idc := metrics.IssuesDateColumn{ProjectColumn: &models.ProjectColumn{Name: dc.Name, ID: dc.ID}}
		newDateColumns = append(newDateColumns, idc)
	}
	return newDateColumns, nil
}
