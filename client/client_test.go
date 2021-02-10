package client

import (
	"context"
	"testing"

	"github.com/3xcellent/github-metrics/config"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		config     config.APIConfig
		want       *MetricsClient
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:       "returns error when baseURL is not set",
			config:     config.APIConfig{Token: "token"},
			wantErr:    true,
			wantErrMsg: "github baseURL must be set",
		},
		{
			name:       "returns error when Token is not set",
			config:     config.APIConfig{BaseURL: "https://api.github.com"},
			wantErr:    true,
			wantErrMsg: "github access token not set",
		},
		{
			name:       "no errors",
			config:     config.APIConfig{BaseURL: "https://api.github.com"},
			wantErr:    true,
			wantErrMsg: "github access token not set",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(context.Background(), tt.config)
			if err != nil {
				if tt.wantErr {
					assert.EqualError(t, err, tt.wantErrMsg)
					return
				}
				t.Errorf("New() got unexpected error: %v", err)
			}
		})
	}
}
