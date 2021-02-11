package runners_test

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/3xcellent/github-metrics/config"
	"github.com/3xcellent/github-metrics/metrics"
	"github.com/3xcellent/github-metrics/metrics/runners"
	"github.com/3xcellent/github-metrics/metrics/runners/runnersfakes"
	"github.com/3xcellent/github-metrics/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testCtx, testCtxCancelFunc = context.WithCancel(context.Background())
	numDays                    = rand.Intn(100)
	startDate                  = time.Date(2001, 2, 3, 0, 0, 0, 0, time.Now().Location()) // if not midnight, adds a day
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
	projectcolumns = models.ProjectColumns{
		{Name: "Column 1", ID: 1},
		{Name: "Column 2", ID: 2},
	}
	repos = models.Repositories{
		{Name: "repo 1", Owner: project.Owner, ID: 1, URL: "some url"},
		{Name: "repo 2", Owner: project.Owner, ID: 2, URL: "some url"},
	}
	issues = models.Issues{
		{Owner: repos[0].Owner, RepoName: repos[0].Name, Number: rand.Intn(100)},
	}
	issueEvents = models.IssueEvents{
		{ProjectID: projectID, Type: models.AddedToProject, ColumnName: projectcolumns[0].Name, CreatedAt: startDate.AddDate(0, 0, 1)},
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
			fakeClient.GetProjectColumnsReturns(projectcolumns, nil)

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
			fakeClient.GetProjectColumnsReturns(projectcolumns, nil)

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
			fakeClient.GetProjectColumnsReturns(projectcolumns, nil)
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

func TestColumnsRunner_Run_NoError(t *testing.T) {
	fakeClient := new(runnersfakes.FakeClient)
	runConfig := config.RunConfig{
		ProjectID: 42,
		StartDate: startDate,
		EndDate:   startDate.AddDate(0, 0, 30),
	}
	object := runners.NewColumnsRunner(runConfig, fakeClient)
	fakeClient.GetProjectReturns(models.Project{Name: projectName}, nil)
	fakeClient.GetProjectColumnsReturns(projectcolumns, nil)
	fakeClient.GetReposFromProjectColumnReturns(repos, nil)
	fakeClient.GetIssuesReturns(issues, nil)
	fakeClient.GetIssueEventsReturns(issueEvents, nil)

	actualErr := object.Run(testCtx)
	assert.NoError(t, actualErr)

	t.Run("assigns project name to runner", func(t *testing.T) {
		assert.Equal(t, projectName, object.ProjectName)
	})

	t.Run("sets expected DateColMap", func(t *testing.T) {
		expectedDateColMap := metrics.DateColMap{}
		assert.Equal(t, expectedDateColMap, object.Cols)
	})
	testCtxCancelFunc()
}
