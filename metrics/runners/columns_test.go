package runners

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/3xcellent/github-metrics/client/clientfakes"
	"github.com/3xcellent/github-metrics/config"
	"github.com/3xcellent/github-metrics/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestColumnsRunner_NewColumnsRunner(t *testing.T) {
	numDays := rand.Intn(100)
	startDate := time.Date(2001, 2, 3, 0, 0, 0, 0, time.Now().Location()) // if not midnight, adds a day
	endDate := startDate.AddDate(0, 0, numDays)
	fakeClient := new(clientfakes.FakeClient)
	runConfig := config.RunConfig{
		StartDate: startDate,
		EndDate:   endDate,
	}

	t.Run("NewColumnsRunner", func(t *testing.T) {
		object := NewColumnsRunner(runConfig, fakeClient)

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
	fakeClient := new(clientfakes.FakeClient)
	runConfig := config.RunConfig{
		StartDate: startDate,
		EndDate:   endDate,
	}

	object := NewColumnsRunner(runConfig, fakeClient)
	object.ProjectName = "Some Project Name"

	t.Run("RunName", func(t *testing.T) {

		t.Run("returns expected", func(t *testing.T) {
			expectedRunName := "Some_Project_Name_columns_2001-02.csv"
			assert.Equal(t, expectedRunName, object.RunName())
		})

	})
}

func TestColumnsRunner_Run(t *testing.T) {
	testCtx, testCtxCancelFunc := context.WithCancel(context.Background())
	startDate := time.Date(2001, 2, 3, 0, 0, 0, 0, time.Now().Location()) // if not midnight, adds a day

	t.Run("getting project details", func(t *testing.T) {
		fakeClient := new(clientfakes.FakeClient)
		runConfig := config.RunConfig{
			ProjectID: 42,
			StartDate: startDate,
			EndDate:   startDate.AddDate(0, 0, 30),
		}
		object := NewColumnsRunner(runConfig, fakeClient)

		t.Run("client returns error", func(t *testing.T) {
			projectsErr := errors.New("project error")
			fakeClient.GetProjectReturns(models.Project{}, projectsErr)
			actualErr := object.Run(testCtx)

			t.Run("returns expected error", func(t *testing.T) {
				assert.Equal(t, projectsErr, actualErr)
			})

			t.Run("calls client with correct params", func(t *testing.T) {
				require.Equal(t, 1, fakeClient.GetProjectCallCount(), "client.GetProject never called")
				actCtx, actProjectID := fakeClient.GetProjectArgsForCall(0)
				assert.Equal(t, testCtx, actCtx)
				assert.Equal(t, runConfig.ProjectID, actProjectID)
			})

		})

	})

	t.Run("when getting project columns", func(t *testing.T) {
		fakeClient := new(clientfakes.FakeClient)
		runConfig := config.RunConfig{
			ProjectID: 42,
			StartDate: startDate,
			EndDate:   startDate.AddDate(0, 0, 30),
		}
		object := NewColumnsRunner(runConfig, fakeClient)

		t.Run("returns correct error", func(t *testing.T) {
			fakeClient.GetProjectReturns(models.Project{
				Name: "Test Project",
				ID:   runConfig.ProjectID,
			}, nil)

			t.Run("from client", func(t *testing.T) {
				projectsColumnsErr := errors.New("project columns error")
				fakeClient.GetProjectColumnsReturns(nil, projectsColumnsErr)
				actualErr := object.Run(context.Background())

				assert.Equal(t, projectsColumnsErr, actualErr)
			})

			t.Run("if no project columns", func(t *testing.T) {
				fakeClient.GetProjectColumnsReturns(nil, nil)
				actualErr := object.Run(context.Background())

				assert.Equal(t, errEmptyProjectColumns, actualErr)
			})
		})
	})

	t.Run("happy path", func(t *testing.T) {
		fakeClient := new(clientfakes.FakeClient)
		runConfig := config.RunConfig{
			ProjectID: 42,
			StartDate: startDate,
			EndDate:   startDate.AddDate(0, 0, 30),
		}
		object := NewColumnsRunner(runConfig, fakeClient)
		projectName := "Project Name"
		fakeClient.GetProjectReturns(models.Project{Name: projectName}, nil)
		actualErr := object.Run(testCtx)
		assert.NoError(t, actualErr)

		t.Run("assigns project name to runner", func(t *testing.T) {
			assert.Equal(t, projectName, object.ProjectName)
		})
	})

	testCtxCancelFunc()
}
