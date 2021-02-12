package metrics

import (
	"testing"

	"github.com/3xcellent/github-metrics/models"
	"github.com/3xcellent/github-metrics/tools/testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestIssueMetric_setColumnDates(t *testing.T) {
	t.Run("events with consecutive column progression", func(t *testing.T) {
		dates := testhelpers.NewDates(4)
		cols := testhelpers.NewProjectColumns(4)

		events := models.IssueEvents{
			{Type: models.MovedColumns, CreatedAt: dates[0], ColumnName: cols[0].Name},
			{Type: models.MovedColumns, CreatedAt: dates[1], ColumnName: cols[1].Name},
			{Type: models.MovedColumns, CreatedAt: dates[2], ColumnName: cols[2].Name},
			{Type: models.MovedColumns, CreatedAt: dates[3], ColumnName: cols[3].Name},
		}

		issue := Issue{
			Issue:            testhelpers.NewIssue(),
			ProjectID:        42,
			StartColumnIndex: 0,
			EndColumnIndex:   3,
			ColumnDates: IssuesDateColumns{
				{ProjectColumn: &cols[0]},
				{ProjectColumn: &cols[1]},
				{ProjectColumn: &cols[2]},
				{ProjectColumn: &cols[3]},
			},
		}
		issue.Events = events
		issue.setColumnDates()

		t.Run("assign columndates with those consecutive dates", func(t *testing.T) {
			expectedIssue := Issue{
				ColumnDates: IssuesDateColumns{
					{ProjectColumn: &cols[0], Date: dates[0]},
					{ProjectColumn: &cols[1], Date: dates[1]},
					{ProjectColumn: &cols[2], Date: dates[2]},
					{ProjectColumn: &cols[3], Date: dates[3]},
				},
			}

			assertColumnDates(t, expectedIssue.ColumnDates, issue.ColumnDates)
		})
	})
	t.Run("events with the previous column index greater than the new column", func(t *testing.T) {
		dates := testhelpers.NewDates(8)
		cols := testhelpers.NewProjectColumns(4)

		events := models.IssueEvents{
			{Type: models.MovedColumns, CreatedAt: dates[0], ColumnName: cols[0].Name},
			{Type: models.MovedColumns, CreatedAt: dates[1], ColumnName: cols[1].Name},
			{Type: models.MovedColumns, CreatedAt: dates[2], ColumnName: cols[2].Name},
			{Type: models.MovedColumns, CreatedAt: dates[3], ColumnName: cols[3].Name},
			{Type: models.MovedColumns, CreatedAt: dates[4], ColumnName: cols[0].Name, PreviousColumnName: cols[3].Name}, // important when moving backwards
			{Type: models.MovedColumns, CreatedAt: dates[5], ColumnName: cols[1].Name},
			{Type: models.MovedColumns, CreatedAt: dates[6], ColumnName: cols[2].Name},
			{Type: models.MovedColumns, CreatedAt: dates[7], ColumnName: cols[3].Name},
		}

		issue := Issue{
			Issue:            testhelpers.NewIssue(),
			StartColumnIndex: 0,
			EndColumnIndex:   3,
			ColumnDates: IssuesDateColumns{
				{ProjectColumn: &cols[0]},
				{ProjectColumn: &cols[1]},
				{ProjectColumn: &cols[2]},
				{ProjectColumn: &cols[3]},
			},
		}
		issue.Events = events

		issue.setColumnDates()

		t.Run("do not get assigned on the new column", func(t *testing.T) {
			expectedIssue := Issue{
				ColumnDates: IssuesDateColumns{
					{ProjectColumn: &cols[0], Date: dates[0]}, // not dates[4]
					{ProjectColumn: &cols[1], Date: dates[5]},
					{ProjectColumn: &cols[2], Date: dates[6]},
					{ProjectColumn: &cols[3], Date: dates[7]},
				},
			}

			assertColumnDates(t, expectedIssue.ColumnDates, issue.ColumnDates)
		})
	})
	t.Run("events with columns not in list of project columns", func(t *testing.T) {
		dates := testhelpers.NewDates(4)
		cols := testhelpers.NewProjectColumns(4)

		issue := Issue{
			Issue:            testhelpers.NewIssue(),
			ProjectID:        42,
			StartColumnIndex: 0,
			EndColumnIndex:   2,
			ColumnDates: IssuesDateColumns{
				{ProjectColumn: &cols[0]},
				{ProjectColumn: &cols[1]},
				{ProjectColumn: &cols[2]},
				{ProjectColumn: &cols[3]},
			},
		}

		events := models.IssueEvents{
			{Type: models.MovedColumns, CreatedAt: dates[0], ColumnName: cols[0].Name},
			{Type: models.MovedColumns, CreatedAt: dates[1], ColumnName: "Column that doesn't exist"},
			{Type: models.MovedColumns, CreatedAt: dates[2], ColumnName: cols[2].Name},
			{Type: models.MovedColumns, CreatedAt: dates[3], ColumnName: cols[3].Name},
		}
		issue.Events = events
		t.Run("gets ignored", func(t *testing.T) {
			issue.setColumnDates()

			expectedIssue := Issue{
				ColumnDates: IssuesDateColumns{
					{ProjectColumn: &cols[0], Date: dates[0]},
					{ProjectColumn: &cols[1]}, // never assigned
					{ProjectColumn: &cols[2], Date: dates[2]},
					{ProjectColumn: &cols[3], Date: dates[3]},
				},
			}

			assertColumnDates(t, expectedIssue.ColumnDates, issue.ColumnDates)
		})
	})
}
func TestIssueMetric_setEmptyColumnDates(t *testing.T) {
	t.Run("ColumnDates with empty dates", func(t *testing.T) {
		dates := testhelpers.NewDates(4)
		cols := testhelpers.NewProjectColumns(4)
		issue := Issue{
			Issue:            testhelpers.NewIssue(),
			ProjectID:        42,
			StartColumnIndex: 0,
			EndColumnIndex:   3,
			ColumnDates: IssuesDateColumns{
				{ProjectColumn: &cols[0], Date: dates[0]},
				{ProjectColumn: &cols[1], Date: dates[1]},
				{ProjectColumn: &cols[2]},
				{ProjectColumn: &cols[3], Date: dates[3]},
			},
		}
		t.Run("get assigned next column date", func(t *testing.T) {
			issue.setEmptyColumnDates()

			expectedIssue := Issue{
				ColumnDates: IssuesDateColumns{
					{ProjectColumn: &cols[0], Date: dates[0]},
					{ProjectColumn: &cols[1], Date: dates[1]},
					{ProjectColumn: &cols[2], Date: dates[3]},
					{ProjectColumn: &cols[3], Date: dates[3]},
				},
			}

			assertColumnDates(t, expectedIssue.ColumnDates, issue.ColumnDates)
		})
	})

	t.Run("when last ColumnDate not set", func(t *testing.T) {
		dates := testhelpers.NewDates(4)
		cols := testhelpers.NewProjectColumns(4)

		issue := Issue{
			Issue:            testhelpers.NewIssue(),
			ProjectID:        42,
			StartColumnIndex: 0,
			EndColumnIndex:   3,
			ColumnDates: IssuesDateColumns{
				{ProjectColumn: &cols[0], Date: dates[0]},
				{ProjectColumn: &cols[1], Date: dates[1]},
				{ProjectColumn: &cols[2], Date: dates[2]},
				{ProjectColumn: &cols[3]},
			},
		}
		t.Run("gets assigned last event date", func(t *testing.T) {
			events := models.IssueEvents{
				{Type: "test event", CreatedAt: dates[0], ColumnName: cols[0].Name},
				{Type: "test event", CreatedAt: dates[1], ColumnName: cols[1].Name},
				{Type: "test event", CreatedAt: dates[2], ColumnName: cols[2].Name},
				{Type: "test event", CreatedAt: dates[3], ColumnName: cols[3].Name},
			}
			issue.Events = events
			issue.setEmptyColumnDates()

			expectedIssue := Issue{
				ColumnDates: IssuesDateColumns{
					{ProjectColumn: &cols[0], Date: dates[0]},
					{ProjectColumn: &cols[1], Date: dates[1]},
					{ProjectColumn: &cols[2], Date: dates[2]},
					{ProjectColumn: &cols[3], Date: dates[3]},
				},
			}

			assertColumnDates(t, expectedIssue.ColumnDates, issue.ColumnDates)
		})
	})

}

func assertColumnDates(t *testing.T, expected, actual IssuesDateColumns) {
	for idx, expectedColumnDate := range expected {
		assert.Equal(t, expectedColumnDate, actual[idx], "columnDate[%d] column: %s | was: %s - expected %s", idx, actual[idx].Date.String(), expectedColumnDate.Name, expectedColumnDate.Date.String())
	}
}
