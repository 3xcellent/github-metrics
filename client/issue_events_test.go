package client

import (
	"context"
	"reflect"
	"testing"

	"github.com/3xcellent/github-metrics/models"
	"github.com/google/go-github/v32/github"
)

func TestMetricsClient_GetIssueEvents(t *testing.T) {
	type fields struct {
		c *github.Client
	}
	type args struct {
		ctx         context.Context
		repoOwner   string
		repoName    string
		issueNumber int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    models.IssueEvents
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MetricsClient{
				c: tt.fields.c,
			}
			got, err := m.GetIssueEvents(tt.args.ctx, tt.args.repoOwner, tt.args.repoName, tt.args.issueNumber)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricsClient.GetIssueEvents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MetricsClient.GetIssueEvents() = %v, want %v", got, tt.want)
			}
		})
	}
}

// events := []*github.IssueEvent{
// 	{Event: &movedColumnsEvent, CreatedAt: &date1, ProjectCard: &github.ProjectCard{ColumnName: &col1}},
// 	{Event: &movedColumnsEvent, CreatedAt: &date2, ProjectCard: &github.ProjectCard{ColumnName: &col2}},
// 	{Event: &movedColumnsEvent, CreatedAt: &date3, ProjectCard: &github.ProjectCard{ColumnName: &col3}},
// 	{Event: &movedColumnsEvent, CreatedAt: &date4, ProjectCard: &github.ProjectCard{ColumnName: &col4}},
// }
