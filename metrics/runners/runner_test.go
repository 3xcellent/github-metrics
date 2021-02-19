package runners_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/3xcellent/github-metrics/config"
	"github.com/3xcellent/github-metrics/metrics/runners"
	"github.com/3xcellent/github-metrics/metrics/runners/runnersfakes"
	"github.com/3xcellent/github-metrics/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
