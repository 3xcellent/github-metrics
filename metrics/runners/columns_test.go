package runners_test

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/3xcellent/github-metrics/config"
	"github.com/3xcellent/github-metrics/metrics/runners"
	"github.com/3xcellent/github-metrics/metrics/runners/runnersfakes"
	"github.com/3xcellent/github-metrics/models"
	"github.com/3xcellent/github-metrics/tools/testhelpers"
	"github.com/stretchr/testify/assert"
)

var (
	testCtx, testCtxCancelFunc = context.WithCancel(context.Background())
	numDays                    = rand.Intn(30) + 2
	numColumns                 = rand.Intn(10) + 2
	startDate                  = time.Date(2001, 2, 3, 0, 0, 0, 0, time.Now().Location()) // must be midnight
	endDate                    = startDate.AddDate(0, 0, numDays)
	projectID                  = int64(123)

	fakeClient = new(runnersfakes.FakeClient)

	runConfig = config.RunConfig{
		ProjectID: projectID,
		StartDate: startDate,
		EndDate:   endDate,
	}
	projectName = "Project Name"
	project     = models.Project{
		Name: projectName,
		ID:   projectID,
	}
	projectColumns = testhelpers.NewProjectColumns(numColumns)
	repos          = models.Repositories{
		{Name: "repo 1", Owner: project.Owner, ID: 1, URL: "some url"},
		{Name: "repo 2", Owner: project.Owner, ID: 2, URL: "some url"},
	}
	issues = models.Issues{
		{Owner: repos[0].Owner, RepoName: repos[0].Name, Number: rand.Intn(100)},
	}
	issueEvents = models.IssueEvents{
		{ProjectID: projectID, Type: models.AddedToProject, ColumnName: projectColumns[0].Name, CreatedAt: startDate.AddDate(0, 0, 1)},
		{ProjectID: projectID, Type: models.MovedColumns, ColumnName: projectColumns[1].Name, CreatedAt: startDate.AddDate(0, 0, 2)},
	}
)

func TestColumnsRunner_NewColumnsRunner(t *testing.T) {
	t.Run("NewColumnsRunner", func(t *testing.T) {
		object := runners.NewColumnsRunner(runConfig, fakeClient)

		t.Run("creates a new base runner and assigns the client", func(t *testing.T) {
			assert.NotNil(t, object.Runner)
			assert.Equal(t, fakeClient, object.Client)
		})

		t.Run("creates a new date column map with the expected number of dates", func(t *testing.T) {
			assert.Len(t, object.Cols, numDays)
		})
	})
}

func TestColumnsRunner_RunName(t *testing.T) {
	numDays := rand.Intn(100)
	startDate := time.Date(2001, 2, 3, 0, 0, 0, 0, time.Now().Location()) // if not midnight, adds a day
	endDate := startDate.AddDate(0, 0, numDays)
	fakeClient := new(runnersfakes.FakeClient)
	runConfig := config.RunConfig{
		StartDate: startDate,
		EndDate:   endDate,
	}

	object := runners.NewColumnsRunner(runConfig, fakeClient)
	object.ProjectName = "Some Project Name"

	t.Run("returns expected", func(t *testing.T) {
		expectedRunName := "Some_Project_Name_columns_2001-02.csv"
		assert.Equal(t, expectedRunName, object.RunName())
	})
}

func TestColumnsRunner_Headers(t *testing.T) {
	fakeClient := new(runnersfakes.FakeClient)
	runConfig := config.RunConfig{
		ProjectID: 42,
		StartDate: startDate,
		EndDate:   startDate.AddDate(0, 0, 30),
	}
	object := runners.NewColumnsRunner(runConfig, fakeClient)
	fakeClient.GetProjectReturns(models.Project{Name: projectName}, nil)
	fakeClient.GetProjectColumnsReturns(projectColumns, nil)
	fakeClient.GetReposFromProjectColumnReturns(repos, nil)
	fakeClient.GetIssuesReturns(issues, nil)
	fakeClient.GetIssueEventsReturns(issueEvents, nil)

	actualErr := object.Run(testCtx)
	assert.NoError(t, actualErr)

	t.Run("returns expected list of header column names", func(t *testing.T) {
		expectedHeaders := []string{"Date"}
		expectedHeaders = append(expectedHeaders, object.ColumnNames...)
		assert.Equal(t, expectedHeaders, object.Headers())
	})
	testCtxCancelFunc()
}

func TestColumnsRunner_Values(t *testing.T) {

}
