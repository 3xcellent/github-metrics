package testhelpers

import (
	"fmt"
	"time"

	"github.com/3xcellent/github-metrics/models"
)

// NewIssue returns a new models.Issue
func NewIssue() *models.Issue {
	return &models.Issue{}
}

// NewDates returns a slice consecutive dates of num length
func NewDates(num int) []time.Time {
	dates := make([]time.Time, 0, num)
	for i := 0; i < num; i++ {
		if i == 0 {
			dates = append(dates, time.Date(2001, 2, 3, 4, 5, 6, 7, time.Now().Location()))
		} else {
			dates = append(dates, dates[0].AddDate(0, 0, i))
		}
	}
	return dates
}

// NewProjectColumns returns a slice of models.ProjectColumn of num length with consecutive
// names like "col 1", "col 2", etc.
func NewProjectColumns(num int) models.ProjectColumns {
	cols := make(models.ProjectColumns, 0, num)
	for i := 0; i < num; i++ {
		cols = append(cols, models.ProjectColumn{Name: fmt.Sprintf("col %d", i), ID: int64(i), Index: i})
	}
	return cols
}
