package runners

import (
	"math/rand"
	"testing"
	"time"

	"github.com/3xcellent/github-metrics/client/clientfakes"
	"github.com/3xcellent/github-metrics/config"
	"github.com/stretchr/testify/assert"
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

	t.Run("creating a new ColumnsRunner", func(t *testing.T) {
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
