package runners

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/3xcellent/github-metrics/config"
	"github.com/3xcellent/github-metrics/models"
	"github.com/sirupsen/logrus"
)

type afterFunc func([][]string) error

// MetricsRunner is the main runner interface all runners must honor
type MetricsRunner interface {
	Run(context.Context) error
	RunName() string
	Values() [][]string
	Headers() []string
	After(afterFunc)
}

// Client - wrapper for github api
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Client
type Client interface {
	GetIssue(ctx context.Context, repoOwner, repoName string, issueNumber int) (models.Issue, error)
	GetProject(ctx context.Context, projectID int64) (models.Project, error)
	GetProjects(ctx context.Context, owner string) (models.Projects, error)
	GetProjectColumns(ctx context.Context, projectID int64) (models.ProjectColumns, error)
	GetPullRequests(ctx context.Context, repoOwner, repoName string) (models.PullRequests, error)
	GetIssues(ctx context.Context, repoOwner string, reposNames []string, beginDate, endDate time.Time) (models.Issues, error)
	GetIssueEvents(ctx context.Context, repoOwner, repoName string, issueNumber int) (models.IssueEvents, error)
	GetReposFromProjectColumn(ctx context.Context, columnID int64) (models.Repositories, error)
}

// Runner - provides a metricsClient, and must honor the CSVRunner interface to allow
// running metrics and running the afterFunc if set
type Runner struct {
	MetricsRunner
	Client  Client
	after   afterFunc
	LogFunc func(args ...interface{})

	NoHeaders bool

	MetricName  string
	ProjectName string
	ProjectID   int64
	Owner       string

	StartColumn      string
	StartDate        time.Time
	StartColumnIndex int
	EndDate          time.Time
	EndColumn        string
	EndColumnIndex   int
	EndColumnID      int64
	ColumnNames      []string
}

// After - sets the afterFunc to one provided
func (r *Runner) After(af afterFunc) {
	r.after = af
}

func New(metricsCfg config.RunConfig, client Client) (MetricsRunner, error) {
	switch metricsCfg.MetricName {
	case "columns":
		return NewColumnsRunner(metricsCfg, client), nil
	case "issues":
		return NewIssuesRunner(metricsCfg, client), nil
	}
	return nil, errors.New("runner name unkonwn")
}

// NewBaseRunner - creates the base runner from the config and set the client client
func NewBaseRunner(metricsCfg config.RunConfig, client Client) *Runner {
	logrus.Debugf("initializing new runner with %#v:", metricsCfg)
	return &Runner{
		Client:      client,
		MetricName:  metricsCfg.MetricName,
		ProjectID:   metricsCfg.ProjectID,
		Owner:       metricsCfg.Owner,
		StartDate:   metricsCfg.StartDate,
		EndDate:     metricsCfg.EndDate,
		StartColumn: metricsCfg.StartColumn,
		EndColumn:   metricsCfg.EndColumn,
		NoHeaders:   metricsCfg.NoHeaders,
	}
}

// errors
var (
	ErrEmptyProjectColumns = errors.New("cannot set indexes: ProjectColumns is empty")
)

// SetColumnParams - sets runner Start/EmdColumnIndec based ProjectColumns and runner.StartColumn/EndColumn values (from RunConfig)
func (r *Runner) setColumnParams(projectColumns models.ProjectColumns) error {
	if len(projectColumns) == 0 {
		return ErrEmptyProjectColumns
	}
	colNames := make([]string, 0)
	for i, col := range projectColumns {
		colNames = append(colNames, col.Name)
		if col.Name == r.StartColumn {
			r.StartColumnIndex = i
			logrus.Debugf("\t index of %q: %d", r.StartColumn, r.StartColumnIndex)
		}

		if col.Name == r.EndColumn {
			r.EndColumnIndex = i
			logrus.Debugf("\t index of %q: %d", r.EndColumn, r.EndColumnIndex)
		}
	}
	if r.EndColumnIndex == 0 {
		r.EndColumnIndex = len(projectColumns) - 1
		logrus.Debugf("\t setting end column: %d", r.EndColumnIndex)
	}
	r.EndColumnID = projectColumns[r.EndColumnIndex].ID
	r.ColumnNames = colNames[r.StartColumnIndex : r.EndColumnIndex+1]
	logrus.Debugf("\tcalculating for columns [%d:%d]: %s", r.StartColumnIndex, r.EndColumnIndex, strings.Join(r.ColumnNames, ","))
	return nil
}

// GetIssuesAndColumns returns the issues and columns for a project
func (r *Runner) GetIssuesAndColumns(ctx context.Context) (models.Issues, models.ProjectColumns, error) {
	var issues models.Issues

	project, err := r.Client.GetProject(ctx, r.ProjectID)
	if err != nil {
		return nil, nil, err
	}
	r.ProjectName = project.Name

	projectColumns, err := r.Client.GetProjectColumns(ctx, r.ProjectID)
	if err != nil {
		return nil, nil, err
	}

	err = r.setColumnParams(projectColumns)
	if err != nil {
		return nil, nil, err
	}

	logrus.Debugf("getting repos: %#v", r)
	repos, err := r.Client.GetReposFromProjectColumn(ctx, r.EndColumnID)
	if err != nil {
		return nil, nil, err
	}
	logrus.Debugf("\trepos found: %s", strings.Join(repos.Names(), ","))

	repoIssues, err := r.Client.GetIssues(ctx, r.Owner, repos.Names(), r.StartDate, r.EndDate)
	if err != nil {
		return nil, nil, err
	}
	logrus.Debugf("\ttotal repo issues found: %d", len(repoIssues))

	for _, issue := range repoIssues {
		issue.Events, err = r.Client.GetIssueEvents(ctx, issue.Owner, issue.RepoName, issue.Number)
		logrus.Debugf("\t %d events for: %s/%d", len(issue.Events), issue.RepoName, issue.Number)
		if err != nil {
			return nil, nil, err
		}

		issues = append(issues, issue)
	}
	return issues, projectColumns, nil
}

// RunName - returns formatted filename including the .csv extension
func (r *Runner) RunName() string {
	return fmt.Sprintf("%s_%s_%d-%02d.csv",
		strings.Replace(r.ProjectName, " ", "_", -1),
		r.MetricName,
		r.StartDate.Year(),
		r.StartDate.Month(),
	)
}

func (r *Runner) Debug() {
	logrus.Debugf("\t MetricName: %q", r.MetricName)
	logrus.Debugf("\t Owner: %q", r.Owner)
	logrus.Debugf("\t ProjectName: %q", r.ProjectName)
	logrus.Debugf("\t ProjectID: %d", r.ProjectID)
	logrus.Debugf("\t StartColumn: %q", r.StartColumn)
	logrus.Debugf("\t EndColumn: %q", r.EndColumn)
	logrus.Debugf("\t StartDate: %q", r.StartDate.String())
	logrus.Debugf("\t EndDate: %q", r.EndDate.String())
	logrus.Debugf("\t StartColumnIndex: %d", r.StartColumnIndex)
	logrus.Debugf("\t EndColumnIndex: %d", r.EndColumnIndex)
	logrus.Debugf("\t EndColumnID: %d", r.EndColumnID)
}
