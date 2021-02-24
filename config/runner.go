package config

import (
	"sort"
	"time"
)

// RunConfig - settings used to run metrics
type RunConfig struct {
	Name        string
	MetricName  string
	Owner       string
	ProjectID   int64
	RepoName    string
	IssueNumber int
	CreateFile  bool
	NoHeaders   bool

	StartColumn      string
	StartDate        time.Time
	StartColumnIndex int
	EndDate          time.Time
	EndColumn        string
	EndColumnIndex   int
	EndColumnID      int64
	ColumnNames      []string
}

// RunConfigs - provides access to getting a RunCofnig by ID or Name
type RunConfigs []RunConfig

// SortedNames - returns the list of project names available sorted by name
func (runConfigs RunConfigs) SortedNames() []string {
	names := make([]string, 0, len(runConfigs))
	for _, rc := range runConfigs {
		names = append(names, rc.Name)
	}
	sort.Strings(names)
	return names
}
