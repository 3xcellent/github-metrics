package runners_test

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/3xcellent/github-metrics/config"
	"github.com/3xcellent/github-metrics/metrics/runners"
	"github.com/3xcellent/github-metrics/metrics/runners/runnersfakes"
	"github.com/3xcellent/github-metrics/models"
	"github.com/3xcellent/github-metrics/tools/testhelpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		expectedHeaders := []string{"Day"}
		expectedHeaders = append(expectedHeaders, object.ColumnNames...)
		assert.Equal(t, expectedHeaders, object.Headers())
	})
	testCtxCancelFunc()
}

func TestColumnsRunner_Values(t *testing.T) {

}

func TestColumnsRunner_Run_Error(t *testing.T) {
	startDate := time.Date(2001, 2, 3, 0, 0, 0, 0, time.Now().Location()) // if not midnight, adds a day

	t.Run("getting project", func(t *testing.T) {
		fakeClient := new(runnersfakes.FakeClient)
		runConfig := config.RunConfig{
			ProjectID: 42,
			StartDate: startDate,
			EndDate:   startDate.AddDate(0, 0, 30),
		}
		object := runners.NewColumnsRunner(runConfig, fakeClient)

		projectsErr := errors.New("project error")
		fakeClient.GetProjectReturns(models.Project{}, projectsErr)
		actualErr := object.Run(testCtx)

		t.Run("returns expected error", func(t *testing.T) {
			assert.Equal(t, projectsErr, actualErr)
		})

		t.Run("is called with correct params", func(t *testing.T) {
			require.Equal(t, 1, fakeClient.GetProjectCallCount(), "client.GetProject never called")
			actCtx, actProjectID := fakeClient.GetProjectArgsForCall(0)
			assert.Equal(t, testCtx, actCtx)
			assert.Equal(t, runConfig.ProjectID, actProjectID)
		})
	})

	t.Run("getting project columns", func(t *testing.T) {
		fakeClient := new(runnersfakes.FakeClient)
		runConfig := config.RunConfig{
			ProjectID: 42,
			StartDate: startDate,
			EndDate:   startDate.AddDate(0, 0, 30),
		}
		object := runners.NewColumnsRunner(runConfig, fakeClient)

		t.Run("client", func(t *testing.T) {
			fakeClient.GetProjectReturns(project, nil)
			projectsColumnsErr := errors.New("project columns error")
			fakeClient.GetProjectColumnsReturns(nil, projectsColumnsErr)
			actualErr := object.Run(testCtx)

			t.Run("returns expected error", func(t *testing.T) {
				assert.Equal(t, projectsColumnsErr, actualErr)
			})

			t.Run("is called with correct params", func(t *testing.T) {
				require.Equal(t, 1, fakeClient.GetProjectColumnsCallCount(), "client.GetProject never called")
				actCtx, actProjectID := fakeClient.GetProjectColumnsArgsForCall(0)
				assert.Equal(t, testCtx, actCtx)
				assert.Equal(t, runConfig.ProjectID, actProjectID)
			})
		})
		t.Run("with no project columns", func(t *testing.T) {
			fakeClient.GetProjectReturns(project, nil)
			fakeClient.GetProjectColumnsReturns(nil, nil)
			actualErr := object.Run(context.Background())

			t.Run("returns expected error", func(t *testing.T) {
				assert.Equal(t, runners.ErrEmptyProjectColumns, actualErr)
			})
		})
	})

	t.Run("getting repos", func(t *testing.T) {
		fakeClient := new(runnersfakes.FakeClient)
		runConfig := config.RunConfig{
			ProjectID: 42,
			StartDate: startDate,
			EndDate:   startDate.AddDate(0, 0, 30),
		}
		object := runners.NewColumnsRunner(runConfig, fakeClient)

		t.Run("client", func(t *testing.T) {
			fakeClient.GetProjectReturns(project, nil)
			fakeClient.GetProjectColumnsReturns(projectColumns, nil)

			reposErr := errors.New("repos error")
			fakeClient.GetReposFromProjectColumnReturns(nil, reposErr)
			actualErr := object.Run(testCtx)
			t.Run("returns expected error", func(t *testing.T) {
				assert.Equal(t, reposErr, actualErr)
			})

			t.Run("is called with correct params", func(t *testing.T) {
				require.Equal(t, 1, fakeClient.GetReposFromProjectColumnCallCount(), "client.GetReposFromProjectColumn never called")
				actCtx, actColumnID := fakeClient.GetReposFromProjectColumnArgsForCall(0)
				assert.Equal(t, testCtx, actCtx)
				assert.Equal(t, object.EndColumnID, actColumnID)
			})
		})
	})

	t.Run("getting issues", func(t *testing.T) {
		fakeClient := new(runnersfakes.FakeClient)
		runConfig := config.RunConfig{
			ProjectID: 42,
			StartDate: startDate,
			EndDate:   startDate.AddDate(0, 0, 30),
		}
		object := runners.NewColumnsRunner(runConfig, fakeClient)

		t.Run("client", func(t *testing.T) {
			fakeClient.GetProjectReturns(project, nil)
			fakeClient.GetProjectColumnsReturns(projectColumns, nil)

			fakeClient.GetReposFromProjectColumnReturns(repos, nil)

			issuesErr := errors.New("issues error")
			fakeClient.GetIssuesReturns(nil, issuesErr)

			actualErr := object.Run(testCtx)

			t.Run("returns expected error", func(t *testing.T) {
				assert.Equal(t, issuesErr, actualErr)
			})

			t.Run("is called with correct params", func(t *testing.T) {
				require.Equal(t, 1, fakeClient.GetIssuesCallCount(), "client.GetIssues never called")
				actCtx, actOwner, actRepoNames, actStartDate, actEndDate := fakeClient.GetIssuesArgsForCall(0)
				assert.Equal(t, testCtx, actCtx)
				assert.Equal(t, object.Owner, actOwner)
				assert.Equal(t, repos.Names(), actRepoNames)
				assert.Equal(t, object.StartDate, actStartDate)
				assert.Equal(t, object.EndDate, actEndDate)
			})
		})
	})

	t.Run("getting issue events", func(t *testing.T) {
		fakeClient := new(runnersfakes.FakeClient)
		runConfig := config.RunConfig{
			ProjectID: 42,
			StartDate: startDate,
			EndDate:   startDate.AddDate(0, 0, 30),
		}
		object := runners.NewColumnsRunner(runConfig, fakeClient)

		t.Run("client", func(t *testing.T) {
			fakeClient.GetProjectReturns(project, nil)
			fakeClient.GetProjectColumnsReturns(projectColumns, nil)
			fakeClient.GetReposFromProjectColumnReturns(repos, nil)
			fakeClient.GetIssuesReturns(issues, nil)

			issueEventsErr := errors.New("issue events error")
			fakeClient.GetIssueEventsReturns(nil, issueEventsErr)

			actualErr := object.Run(testCtx)

			t.Run("returns expected error", func(t *testing.T) {
				assert.Equal(t, issueEventsErr, actualErr)
			})

			t.Run("is called with correct params", func(t *testing.T) {
				require.Equal(t, 1, fakeClient.GetIssueEventsCallCount(), "client.GetIssues never called")
				actCtx, actOwner, actRepoName, actIssueNumber := fakeClient.GetIssueEventsArgsForCall(0)
				assert.Equal(t, testCtx, actCtx)
				assert.Equal(t, issues[0].Owner, actOwner)
				assert.Equal(t, issues[0].RepoName, actRepoName)
				assert.Equal(t, issues[0].Number, actIssueNumber)
			})
		})
	})

	testCtxCancelFunc()
}

func TestColumnsRunner_Run(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	fakeClient := new(runnersfakes.FakeClient)
	runConfig := config.RunConfig{
		ProjectID: 42,
		StartDate: startDate,
		EndDate:   startDate.AddDate(0, 0, numDays),
	}
	object := runners.NewColumnsRunner(runConfig, fakeClient)
	fakeClient.GetProjectReturns(models.Project{Name: projectName}, nil)
	fakeClient.GetProjectColumnsReturns(projectColumns, nil)
	fakeClient.GetReposFromProjectColumnReturns(repos, nil)
	fakeClient.GetIssuesReturns(issues, nil)
	fakeClient.GetIssueEventsReturns(issueEvents, nil)

	err := object.Run(testCtx)

	t.Run("has correct start date and end dates", func(t *testing.T) {
		assert.Equal(t, startDate, object.StartDate)
		assert.Equal(t, endDate, object.EndDate)
	})

	t.Run("does not return error", func(t *testing.T) {
		require.NoError(t, err)
	})

	t.Run("assigns project name to runner", func(t *testing.T) {
		assert.Equal(t, projectName, object.ProjectName)
	})

	t.Run("sets DateColMap", func(t *testing.T) {
		assert.Len(t, object.Cols, numDays)
	})
	testCtxCancelFunc()
}
