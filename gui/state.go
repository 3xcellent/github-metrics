package gui

import (
	"context"

	"github.com/3xcellent/github-metrics/client"
	"github.com/3xcellent/github-metrics/config"
	"github.com/sirupsen/logrus"
)

// MetricsState object mainstans state for gui
type MetricsState struct {
	Context   context.Context
	APIConfig config.APIConfig
	Client    *client.MetricsClient

	HasUpdatedAPIConfig    bool
	HasValidatedConnection bool

	RunRequested bool
	RunStarted   bool
	RunCompleted bool

	SelectedProjectID   int64
	SelectedProjectName string
}

// NewState - returns new state object
func NewState(ctx context.Context) *MetricsState {
	return &MetricsState{Context: ctx}
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
	s.Client, err = client.New(s.Context, s.APIConfig)
	if err != nil {
		return err
	}
	logrus.Info("initialized client")
	s.HasUpdatedAPIConfig = false
	s.HasValidatedConnection = true

	return nil
}
