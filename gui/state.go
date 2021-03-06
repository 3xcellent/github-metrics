package gui

import (
	"context"

	"github.com/3xcellent/github-metrics/client"
	"github.com/3xcellent/github-metrics/config"
	"github.com/3xcellent/github-metrics/models"
	"github.com/sirupsen/logrus"
)

// MetricsState object mainstans state for gui
type MetricsState struct {
	APIConfig config.APIConfig
	Client    *client.MetricsClient

	HasUpdatedAPIConfig    bool
	HasValidatedConnection bool

	RunConfig    config.RunConfig
	RunRequested bool
	RunStarted   bool
	RunCompleted bool
	RunValues    [][]string

	Result Result

	SelectedProjectID   int64
	SelectedProjectName string
}

type Result struct {
	MetricName string
	Owner      string
	ProjectID  int64
	Repos      Repositories
}

type Repository struct {
	m models.Repository
}
type Repositories []Repository

// NewState - returns new state object
func NewState(ctx context.Context) *MetricsState {
	return &MetricsState{}
}

// SetClient - applies settings from the API Config and sets HasUpdateAPIConfig to true
func (s *MetricsState) SetClient(cfg config.APIConfig) error {
	c := cfg
	s.APIConfig.Token = c.Token
	s.APIConfig.Owner = c.Owner
	s.APIConfig.BaseURL = c.BaseURL
	s.APIConfig.UploadURL = c.UploadURL
	s.HasUpdatedAPIConfig = true
	s.HasValidatedConnection = false

	var err error
	ctx := context.Background()
	s.Client, err = client.New(ctx, s.APIConfig)
	if err != nil {
		return err
	}
	logrus.Info("initialized client")
	s.HasUpdatedAPIConfig = false
	s.HasValidatedConnection = true

	return nil
}
