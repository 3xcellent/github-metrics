package metrics

import (
	"testing"
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/stretchr/testify/assert"
)

func TestIssueMetric_ProcessIssueEvents(t *testing.T) {
	beginColumnIndex := 0
	t.Run("consecutive column progression", func(t *testing.T) {
		date1 := time.Date(2001, 2, 3, 4, 5, 6, 7, time.Now().Location())
		date2 := date1.Add(time.Hour * 24)
		date3 := date1.Add(time.Hour * 24)
		date4 := date1.Add(time.Hour * 24)

		col1 := "col 1"
		col2 := "col 2"
		col3 := "col 3"
		col4 := "col 4"

		movedColumnsEvent := string(MovedColumns)

		events := []*github.IssueEvent{
			{Event: &movedColumnsEvent, CreatedAt: &date1, ProjectCard: &github.ProjectCard{ColumnName: &col1}},
			{Event: &movedColumnsEvent, CreatedAt: &date2, ProjectCard: &github.ProjectCard{ColumnName: &col2}},
			{Event: &movedColumnsEvent, CreatedAt: &date3, ProjectCard: &github.ProjectCard{ColumnName: &col3}},
			{Event: &movedColumnsEvent, CreatedAt: &date4, ProjectCard: &github.ProjectCard{ColumnName: &col4}},
		}

		issue := Issue{
			//beginColumnIdx: 0,
			//endColumnIdx:   3,
			ColumnDates: BoardColumns{
				{Name: col1},
				{Name: col2},
				{Name: col3},
				{Name: col4},
			}}
		issue.ProcessIssueEvents(events, beginColumnIndex)

		expectedIssue := Issue{
			//beginColumnIdx: 0,
			//endColumnIdx:   3,
			ColumnDates: BoardColumns{
				{Name: col1, Date: date1},
				{Name: col2, Date: date2},
				{Name: col3, Date: date3},
				{Name: col4, Date: date4},
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

		col1 := "col 1"
		col2 := "col 2"
		col3 := "col 3"
		col4 := "col 4"

		movedColumnsEvent := string(MovedColumns)

		events := []*github.IssueEvent{
			{Event: &movedColumnsEvent, CreatedAt: &date1, ProjectCard: &github.ProjectCard{ColumnName: &col1}},
			{Event: &movedColumnsEvent, CreatedAt: &date2, ProjectCard: &github.ProjectCard{ColumnName: &col2}},
			{Event: &movedColumnsEvent, CreatedAt: &date3, ProjectCard: &github.ProjectCard{ColumnName: &col3}},
			{Event: &movedColumnsEvent, CreatedAt: &date4, ProjectCard: &github.ProjectCard{ColumnName: &col4}},
			{Event: &movedColumnsEvent, CreatedAt: &date5, ProjectCard: &github.ProjectCard{ColumnName: &col1}},
			{Event: &movedColumnsEvent, CreatedAt: &date6, ProjectCard: &github.ProjectCard{ColumnName: &col2}},
			{Event: &movedColumnsEvent, CreatedAt: &date7, ProjectCard: &github.ProjectCard{ColumnName: &col3}},
			{Event: &movedColumnsEvent, CreatedAt: &date8, ProjectCard: &github.ProjectCard{ColumnName: &col4}},
		}

		issue := Issue{
			//beginColumnIdx: 0,
			//endColumnIdx:   3,
			ColumnDates: BoardColumns{
				{Name: col1},
				{Name: col2},
				{Name: col3},
				{Name: col4},
			}}
		issue.ProcessIssueEvents(events, beginColumnIndex)

		expectedIssue := Issue{
			//beginColumnIdx: 0,
			//endColumnIdx:   3,
			ColumnDates: BoardColumns{
				{Name: col1, Date: date5},
				{Name: col2, Date: date6},
				{Name: col3, Date: date7},
				{Name: col4, Date: date8},
			}}

		assert.Equal(t, expectedIssue.ColumnDates, issue.ColumnDates)
	})
	t.Run("when column is skipped, it will have same date as next column", func(t *testing.T) {
		date1 := time.Date(2001, 2, 3, 4, 5, 6, 7, time.Now().Location())
		date2 := date1.Add(time.Hour * 24)
		date3 := date1.Add(time.Hour * 24)

		col1 := "col 1"
		col2 := "col 2"
		col3 := "col 3"
		col4 := "col 4"

		movedColumnsEvent := string(MovedColumns)

		events := []*github.IssueEvent{
			{Event: &movedColumnsEvent, CreatedAt: &date1, ProjectCard: &github.ProjectCard{ColumnName: &col1}},
			{Event: &movedColumnsEvent, CreatedAt: &date2, ProjectCard: &github.ProjectCard{ColumnName: &col3}},
			{Event: &movedColumnsEvent, CreatedAt: &date3, ProjectCard: &github.ProjectCard{ColumnName: &col4}},
		}

		issue := Issue{
			//beginColumnIdx: 0,
			//endColumnIdx:   3,
			ColumnDates: BoardColumns{
				{Name: col1},
				{Name: col2},
				{Name: col3},
				{Name: col4},
			}}
		issue.ProcessIssueEvents(events, beginColumnIndex)

		expectedIssue := Issue{
			//beginColumnIdx: 0,
			//endColumnIdx:   3,
			ColumnDates: BoardColumns{
				{Name: col1, Date: date1},
				{Name: col2, Date: date2},
				{Name: col3, Date: date2},
				{Name: col4, Date: date3},
			}}

		assert.Equal(t, expectedIssue.ColumnDates, issue.ColumnDates)
	})
	t.Run("when column is skipped is last column, it gets assigned last event date", func(t *testing.T) {
		date1 := time.Date(2001, 2, 3, 4, 5, 6, 7, time.Now().Location())
		date2 := date1.Add(time.Hour * 24)
		date3 := date1.Add(time.Hour * 24)

		col1 := "col 1"
		col2 := "col 2"
		col3 := "col 3"
		col4 := "col 4"

		movedColumnsEvent := string(MovedColumns)
		someEvent := "some event"
		events := []*github.IssueEvent{
			{Event: &movedColumnsEvent, CreatedAt: &date1, ProjectCard: &github.ProjectCard{ColumnName: &col1}},
			{Event: &movedColumnsEvent, CreatedAt: &date2, ProjectCard: &github.ProjectCard{ColumnName: &col2}},
			{Event: &someEvent, CreatedAt: &date3, ProjectCard: &github.ProjectCard{ColumnName: &col3}},
		}

		issue := Issue{
			//beginColumnIdx: 0,
			//endColumnIdx:   3,
			ColumnDates: BoardColumns{
				{Name: col1},
				{Name: col2},
				{Name: col3},
				{Name: col4},
			}}
		issue.ProcessIssueEvents(events, beginColumnIndex)

		expectedIssue := Issue{
			//beginColumnIdx: 0,
			//endColumnIdx:   3,
			ColumnDates: BoardColumns{
				{Name: col1, Date: date1},
				{Name: col2, Date: date2},
				{Name: col3, Date: date3},
				{Name: col4, Date: date3},
			}}

		assert.Equal(t, expectedIssue.ColumnDates, issue.ColumnDates)
	})
	t.Run("when event column is not in list of board columns", func(t *testing.T) {
		date1 := time.Date(2001, 2, 3, 4, 5, 6, 7, time.Now().Location())
		date2 := date1.Add(time.Hour * 24)
		date3 := date1.Add(time.Hour * 24)

		col1 := "col 1"
		col2 := "I don't belong here"
		col3 := "col 3"
		col4 := "col 4"

		movedColumnsEvent := string(MovedColumns)
		someEvent := "some event"
		events := []*github.IssueEvent{
			{Event: &movedColumnsEvent, CreatedAt: &date1, ProjectCard: &github.ProjectCard{ColumnName: &col1}},
			{Event: &movedColumnsEvent, CreatedAt: &date2, ProjectCard: &github.ProjectCard{ColumnName: &col2}},
			{Event: &someEvent, CreatedAt: &date3, ProjectCard: &github.ProjectCard{ColumnName: &col3}},
		}

		issue := Issue{
			//beginColumnIdx: 0,
			//endColumnIdx:   3,
			ColumnDates: BoardColumns{
				{Name: col1},
				{Name: col3},
				{Name: col4},
			}}
		issue.ProcessIssueEvents(events, beginColumnIndex)

		expectedIssue := Issue{
			//beginColumnIdx: 0,
			//endColumnIdx:   3,
			ColumnDates: BoardColumns{
				{Name: col1, Date: date1},
				{Name: col3, Date: date3},
				{Name: col4, Date: date3},
			}}

		assert.Equal(t, expectedIssue.ColumnDates, issue.ColumnDates)
	})
}
