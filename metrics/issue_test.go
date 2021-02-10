package metrics

import (
	"testing"
	"time"

	"github.com/3xcellent/github-metrics/models"
	"github.com/stretchr/testify/assert"
)

func TestIssueMetric_ProcessIssueEvents(t *testing.T) {
	beginColumnIndex := 0
	ghIssue := models.Issue{}

	t.Run("consecutive column progression", func(t *testing.T) {
		date1 := time.Date(2001, 2, 3, 4, 5, 6, 7, time.Now().Location())
		date2 := date1.Add(time.Hour * 24)
		date3 := date1.Add(time.Hour * 24)
		date4 := date1.Add(time.Hour * 24)

		col1 := &models.ProjectColumn{Name: "col 1", ID: 1}
		col2 := &models.ProjectColumn{Name: "col 2", ID: 1}
		col3 := &models.ProjectColumn{Name: "col 3", ID: 1}
		col4 := &models.ProjectColumn{Name: "col 4", ID: 1}

		events := models.IssueEvents{
			{Type: models.MovedColumns, CreatedAt: date1, ColumnName: col1.Name},
			{Type: models.MovedColumns, CreatedAt: date2, ColumnName: col2.Name},
			{Type: models.MovedColumns, CreatedAt: date3, ColumnName: col3.Name},
			{Type: models.MovedColumns, CreatedAt: date4, ColumnName: col4.Name},
		}

		issue := Issue{
			Issue:            &ghIssue,
			ProjectID:        42,
			StartColumnIndex: beginColumnIndex,
			EndColumnIndex:   3,
			ColumnDates: IssuesDateColumns{
				{ProjectColumn: col1},
				{ProjectColumn: col2},
				{ProjectColumn: col3},
				{ProjectColumn: col4},
			},
		}
		issue.ProcessIssueEvents(events)

		expectedIssue := Issue{
			ColumnDates: IssuesDateColumns{
				{ProjectColumn: col1, Date: date1},
				{ProjectColumn: col2, Date: date2},
				{ProjectColumn: col3, Date: date3},
				{ProjectColumn: col4, Date: date4},
			}}

		assert.Equal(t, expectedIssue.ColumnDates, issue.ColumnDates)
	})
	t.Run("column goes back a column before progressing", func(t *testing.T) {
		date1 := time.Date(2001, 2, 3, 4, 5, 6, 7, time.Now().Location())
		date2 := date1.Add(time.Hour * 24)
		date3 := date1.Add(time.Hour * 24)
		date4 := date1.Add(time.Hour * 24)
		date5 := date1.Add(time.Hour * 24)
		date6 := date1.Add(time.Hour * 24)
		date7 := date1.Add(time.Hour * 24)
		date8 := date1.Add(time.Hour * 24)

		col1 := &models.ProjectColumn{Name: "col 1", ID: 1}
		col2 := &models.ProjectColumn{Name: "col 2", ID: 1}
		col3 := &models.ProjectColumn{Name: "col 3", ID: 1}
		col4 := &models.ProjectColumn{Name: "col 4", ID: 1}

		events := models.IssueEvents{
			{Type: models.MovedColumns, CreatedAt: date1, ColumnName: col1.Name},
			{Type: models.MovedColumns, CreatedAt: date2, ColumnName: col2.Name},
			{Type: models.MovedColumns, CreatedAt: date3, ColumnName: col3.Name},
			{Type: models.MovedColumns, CreatedAt: date4, ColumnName: col4.Name},
			{Type: models.MovedColumns, CreatedAt: date5, ColumnName: col1.Name},
			{Type: models.MovedColumns, CreatedAt: date6, ColumnName: col2.Name},
			{Type: models.MovedColumns, CreatedAt: date7, ColumnName: col3.Name},
			{Type: models.MovedColumns, CreatedAt: date8, ColumnName: col4.Name},
		}

		issue := Issue{
			Issue:            &ghIssue,
			ProjectID:        42,
			StartColumnIndex: beginColumnIndex,
			EndColumnIndex:   3,
			ColumnDates: IssuesDateColumns{
				{ProjectColumn: col1},
				{ProjectColumn: col2},
				{ProjectColumn: col3},
				{ProjectColumn: col4},
			},
		}
		issue.ProcessIssueEvents(events)

		expectedIssue := Issue{
			ColumnDates: IssuesDateColumns{
				{ProjectColumn: col1, Date: date5},
				{ProjectColumn: col2, Date: date6},
				{ProjectColumn: col3, Date: date7},
				{ProjectColumn: col4, Date: date8},
			}}

		assert.Equal(t, expectedIssue.ColumnDates, issue.ColumnDates)
	})
	t.Run("when column is skipped, it will have same date as next column", func(t *testing.T) {
		date1 := time.Date(2001, 2, 3, 4, 5, 6, 7, time.Now().Location())
		date2 := date1.Add(time.Hour * 24)
		date3 := date1.Add(time.Hour * 24)

		col1 := &models.ProjectColumn{Name: "col 1", ID: 1}
		col2 := &models.ProjectColumn{Name: "col 2", ID: 1}
		col3 := &models.ProjectColumn{Name: "col 3", ID: 1}
		col4 := &models.ProjectColumn{Name: "col 4", ID: 1}

		events := models.IssueEvents{
			{Type: models.MovedColumns, CreatedAt: date1, ColumnName: col1.Name},
			{Type: models.MovedColumns, CreatedAt: date2, ColumnName: col3.Name},
			{Type: models.MovedColumns, CreatedAt: date3, ColumnName: col4.Name},
		}

		issue := Issue{
			Issue:            &ghIssue,
			ProjectID:        42,
			StartColumnIndex: beginColumnIndex,
			EndColumnIndex:   3,
			ColumnDates: IssuesDateColumns{
				{ProjectColumn: col1},
				{ProjectColumn: col2},
				{ProjectColumn: col3},
				{ProjectColumn: col4},
			},
		}
		issue.ProcessIssueEvents(events)

		expectedIssue := Issue{
			ColumnDates: IssuesDateColumns{
				{ProjectColumn: col1, Date: date1},
				{ProjectColumn: col2, Date: date2},
				{ProjectColumn: col3, Date: date2},
				{ProjectColumn: col4, Date: date3},
			}}

		assert.Equal(t, expectedIssue.ColumnDates, issue.ColumnDates)
	})
	t.Run("when column is skipped is last column, it gets assigned last event date", func(t *testing.T) {
		date1 := time.Date(2001, 2, 3, 4, 5, 6, 7, time.Now().Location())
		date2 := date1.Add(time.Hour * 24)
		date3 := date1.Add(time.Hour * 24)

		col1 := &models.ProjectColumn{Name: "col 1", ID: 1}
		col2 := &models.ProjectColumn{Name: "col 2", ID: 1}
		col3 := &models.ProjectColumn{Name: "col 3", ID: 1}
		col4 := &models.ProjectColumn{Name: "col 4", ID: 1}

		events := models.IssueEvents{
			{Type: models.MovedColumns, CreatedAt: date1, ColumnName: col1.Name},
			{Type: models.MovedColumns, CreatedAt: date2, ColumnName: col3.Name},
			{Type: "some event", CreatedAt: date3, ColumnName: col4.Name},
		}

		issue := Issue{
			Issue:            &ghIssue,
			ProjectID:        42,
			StartColumnIndex: beginColumnIndex,
			EndColumnIndex:   3,
			ColumnDates: IssuesDateColumns{
				{ProjectColumn: col1},
				{ProjectColumn: col2},
				{ProjectColumn: col3},
				{ProjectColumn: col4},
			},
		}
		issue.ProcessIssueEvents(events)

		expectedIssue := Issue{
			//beginColumnIdx: 0,
			//endColumnIdx:   3,
			ColumnDates: IssuesDateColumns{
				{ProjectColumn: col1, Date: date1},
				{ProjectColumn: col2, Date: date2},
				{ProjectColumn: col3, Date: date3},
				{ProjectColumn: col4, Date: date3},
			}}

		assert.Equal(t, expectedIssue.ColumnDates, issue.ColumnDates)
	})
	t.Run("when event column is not in list of board columns", func(t *testing.T) {
		date1 := time.Date(2001, 2, 3, 4, 5, 6, 7, time.Now().Location())
		date2 := date1.Add(time.Hour * 24)
		date3 := date1.Add(time.Hour * 24)

		col1 := &models.ProjectColumn{Name: "col 1", ID: 1}
		col2 := &models.ProjectColumn{Name: "I don't belong here", ID: 2}
		col3 := &models.ProjectColumn{Name: "col 3", ID: 3}
		col4 := &models.ProjectColumn{Name: "col 4", ID: 4}

		events := models.IssueEvents{
			{Type: models.MovedColumns, CreatedAt: date1, ColumnName: col1.Name},
			{Type: models.MovedColumns, CreatedAt: date2, ColumnName: col2.Name},
			{Type: "some event", CreatedAt: date1, ColumnName: col3.Name},
		}

		issue := Issue{
			Issue:            &ghIssue,
			ProjectID:        42,
			StartColumnIndex: beginColumnIndex,
			EndColumnIndex:   2,
			ColumnDates: IssuesDateColumns{
				{ProjectColumn: col1},
				{ProjectColumn: col3},
				{ProjectColumn: col4},
			},
		}
		issue.ProcessIssueEvents(events)

		expectedIssue := Issue{
			//beginColumnIdx: 0,
			//endColumnIdx:   3,
			ColumnDates: IssuesDateColumns{
				{ProjectColumn: col1, Date: date1},
				{ProjectColumn: col3, Date: date3},
				{ProjectColumn: col4, Date: date3},
			}}

		assert.Equal(t, expectedIssue.ColumnDates, issue.ColumnDates)
	})
}
