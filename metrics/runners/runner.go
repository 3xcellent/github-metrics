package runners

import (
	"strings"
	"time"

	"github.com/3xcellent/github-metrics/client"
	"github.com/3xcellent/github-metrics/config"
	"github.com/3xcellent/github-metrics/models"
	"github.com/sirupsen/logrus"
)

type afterFunc func([][]string) error

type csvRunner interface {
	Run() error
	RunName() string
	Values() [][]string
	Headers() []string
	After(afterFunc)
}

// Runner - provides a metricsClient, and must honor the CSVRunner interface to allow
// running metrics and running the afterFunc if set
type Runner struct {
	csvRunner
	Client  client.Client
	after   afterFunc
	LogFunc func(args ...interface{})

	NoHeaders bool

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
func (r *Runner) After(afterFunc func([][]string) error) {
	r.after = afterFunc
}

// NewBaseRunner - creates the base runner from the config and set the client client
func NewBaseRunner(metricsCfg config.RunConfig, client client.Client) *Runner {
	logrus.Debugf("initializing new runner with %#v:", metricsCfg)
	return &Runner{
		Client:      client,
		ProjectID:   metricsCfg.ProjectID,
		Owner:       metricsCfg.Owner,
		StartDate:   metricsCfg.StartDate,
		EndDate:     metricsCfg.EndDate,
		StartColumn: metricsCfg.StartColumn,
		EndColumn:   metricsCfg.EndColumn,
		NoHeaders:   metricsCfg.NoHeaders,
	}
}

func (r *Runner) setColumnParams(projectColumns models.ProjectColumns) {
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
}

// Log - uses the provided logFunc for log text
func (r *Runner) Log(values ...interface{}) {
	if r.LogFunc == nil {
		return
	}
	for _, value := range values {
		r.LogFunc(value.(string))
	}
}
