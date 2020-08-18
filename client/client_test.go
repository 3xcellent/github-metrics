package client

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/3xcellent/github-metrics/config"
	"github.com/google/go-github/v32/github"
	"github.com/stretchr/testify/assert"
)

func Test_metricsClient_GetIssuesFromColumn(t *testing.T) {
	type fields struct {
		c *github.Client
	}
	type args struct {
		ctx       context.Context
		repoOwner string
		columnID  int64
		beginDate time.Time
		endDate   time.Time
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string][]*github.Issue
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &metricsClient{
				c: tt.fields.c,
			}
			if got := m.GetIssuesFromColumn(tt.args.ctx, tt.args.repoOwner, tt.args.columnID, tt.args.beginDate, tt.args.endDate); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetIssuesFromColumn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		config     config.APIConfig
		want       *metricsClient
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
				} else {
					t.Errorf("New() got unexpected error: %v", err)
				}
			}
		})
	}
}
