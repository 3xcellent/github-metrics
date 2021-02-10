package runners

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
func NewIssuesRunner(metricsCfg config.RunConfig, client Client) IssuesRunner {
	m := IssuesRunner{
		Runner: NewBaseRunner(metricsCfg, client),
	}

	return m
}

// Filename - returns formatted filename including the .csv extension
func (r *IssuesRunner) Filename() string {
	return fmt.Sprintf("%s_%s_%d-%02d.csv",
		strings.Replace(r.ProjectName, " ", "_", -1),
		"issues",
		r.StartDate.Year(),
		r.StartDate.Month(),
	)
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
	project, err := r.Client.GetProject(ctx, r.ProjectID)
	if err != nil {
		return err
	}
	r.ProjectName = project.Name

	projectColumns, err := r.Client.GetProjectColumns(ctx, r.ProjectID)
	if err != nil {
		return err
	}
	r.setColumnParams(projectColumns)
	dateCols := make(metrics.IssuesDateColumns, 0)
	for _, col := range projectColumns {
		dateCols = append(dateCols, metrics.IssuesDateColumn{ProjectColumn: &models.ProjectColumn{Name: col.Name, ID: col.ID}})
	}
	logrus.Debugf("dateCols: %#v", dateCols)

	r.Issues = metrics.Issues{}
	if r.IssueNumber != 0 && r.RepoName != "" {
		ghIssue, err := r.Client.GetIssue(ctx, r.Owner, r.RepoName, r.IssueNumber)
		if err != nil {
			return err
		}
		metricsIssue, err := r.newMetricsIssue(ctx, ghIssue, dateCols)
		if err != nil {
			return err
		}
		r.Issues = append(r.Issues, metricsIssue)
	} else {
		repos, err := r.Client.GetRepos(ctx, projectColumns[len(projectColumns)-1].ID)
		if err != nil {
			return err
		}
		for _, ghIssue := range r.Client.GetIssues(ctx,
			r.Owner,
			repos,
			r.StartDate,
			r.EndDate,
		) {
			metricsIssue, err := r.newMetricsIssue(ctx, ghIssue, dateCols)
			if err != nil {
				return err
			}

			r.Issues = append(r.Issues, metricsIssue)
		}
	}

	for _, issue := range r.Issues {
		events, err := r.Client.GetIssueEvents(ctx, r.Owner, issue.RepoName, issue.Number)
		if err != nil {
			return err
		}
		issue.ProcessIssueEvents(events)
	}

	logrus.Debugf("done processing %d issues", len(r.Issues))
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
