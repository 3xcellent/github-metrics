package cmd_test

import (
	"testing"
	"time"

	"github.com/3xcellent/github-metrics/cmd"
	"github.com/3xcellent/github-metrics/config"
	"github.com/3xcellent/github-metrics/config/configfakes"
	"github.com/3xcellent/github-metrics/metrics"
	"github.com/google/go-github/v32/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_columns(t *testing.T) {
	projectName := "projectName"
	columnName1 := "columnName1"
	columnName2 := "columnName2"
	columnName3 := "columnName3"

	ID := int64(1234)
	URL := "URL"
	OwnerURL := "OwnerURL"
	body := "body"
	project := github.Project{
		Name:     &projectName,
		ID:       &ID,
		URL:      &URL,
		OwnerURL: &OwnerURL,
		Body:     &body,
	}
	projectColumn1 := github.ProjectColumn{
		Name: &columnName1,
	}
	projectColumn2 := github.ProjectColumn{
		Name: &columnName2,
	}
	projectColumn3 := github.ProjectColumn{
		Name: &columnName3,
	}
	projectColumns := []*github.ProjectColumn{
		&projectColumn1,
		&projectColumn2,
		&projectColumn3,
	}
	issueNumber := 4200
	issueTitle := "issueTitle"
	repoIssuesMap := map[string][]*github.Issue{
		"repoName": {{Number: &issueNumber, Title: &issueTitle}},
	}
	issueEvent1CreatedAt := time.Date(2020, 1, 1, 0, 0, 0, 0, time.Now().Location())
	issueEvent1Type := string(metrics.AddedToProject)
	issueEvent2Type := string(metrics.MovedColumns)
	issueEvent3Type := string(metrics.MovedColumns)
	issueEvent2CreatedAt := issueEvent1CreatedAt.Add(time.Hour * 48)
	issueEvent3CreatedAt := issueEvent2CreatedAt.Add(time.Hour * 48)
	issueEvent1projectCard := github.ProjectCard{
		ProjectID:  &ID,
		ColumnName: &columnName1,
	}
	issueEvent2projectCard := github.ProjectCard{
		ProjectID:  &ID,
		ColumnName: &columnName2,
	}
	issueEvent3projectCard := github.ProjectCard{
		ProjectID:  &ID,
		ColumnName: &columnName3,
	}
	issueEvents := []*github.IssueEvent{
		{
			ID:              nil,
			URL:             nil,
			Actor:           nil,
			Event:           &issueEvent1Type,
			CreatedAt:       &issueEvent1CreatedAt,
			Issue:           nil,
			Assignee:        nil,
			Assigner:        nil,
			CommitID:        nil,
			Milestone:       nil,
			Label:           nil,
			Rename:          nil,
			LockReason:      nil,
			ProjectCard:     &issueEvent1projectCard,
			DismissedReview: nil,
		},
		{
			ID:              nil,
			URL:             nil,
			Actor:           nil,
			Event:           &issueEvent2Type,
			CreatedAt:       &issueEvent2CreatedAt,
			Issue:           nil,
			Assignee:        nil,
			Assigner:        nil,
			CommitID:        nil,
			Milestone:       nil,
			Label:           nil,
			Rename:          nil,
			LockReason:      nil,
			ProjectCard:     &issueEvent2projectCard,
			DismissedReview: nil,
		},
		{
			ID:              nil,
			URL:             nil,
			Actor:           nil,
			Event:           &issueEvent3Type,
			CreatedAt:       &issueEvent3CreatedAt,
			Issue:           nil,
			Assignee:        nil,
			Assigner:        nil,
			CommitID:        nil,
			Milestone:       nil,
			Label:           nil,
			Rename:          nil,
			LockReason:      nil,
			ProjectCard:     &issueEvent3projectCard,
			DismissedReview: nil,
		},
	}
	var configYAML = []byte(`---
API:
  BaseURL: https://enterprise.github.com/api/v3
  Token: token
Boards:
  - board1:
      boardID: 1234
      startColumn: Develop
IncludeHeaders: true
GroupName: An 3xcellent Team
Owner: 3xcellent
LoginNames:
  - 3xcellent
`)

	var columnsUsage = `
Usage:
  github-metrics columns [board_name] [flags]

Flags:
  -h, --help   help for columns

Global Flags:
  -d, --askForDate       command will ask for user to input year and month parameters at runtime
  -c, --create-file      set outpath path to [board_name]_[command_name]_[year]_[month].csv)
  -m, --month int        specify month (default 6)
      --no-headers       disable csv header row
  -o, --outpath string   set output path
  -t, --token string     Auth token used when connecting to github server
  -v, --verbose          verbose output
  -y, --year int         specify year (default 2020)

`

	tests := []struct {
		name       string
		args       []string
		config     []byte
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "returns error when no project name provided",
			config:     configYAML,
			args:       []string{"columns"},
			wantErr:    true,
			wantOutput: "Error: requires at least 1 arg(s), only received 0" + columnsUsage,
		},
		{
			name:       "returns error when no boards configured",
			config:     noBoardsYAML,
			args:       []string{"columns", "board1"},
			wantErr:    true,
			wantOutput: "Error: no project boards configured" + columnsUsage,
		},
		{
			name:    "happy path",
			config:  configYAML,
			args:    []string{"columns", "board1"},
			wantErr: false,
			wantOutput: `
Day,columnName1,columnName2,columnName3
2020-01-01,1,0,0
2020-01-02,1,0,0
2020-01-03,0,1,0
2020-01-04,0,1,0
2020-01-05,0,0,1
2020-01-06,0,0,0
2020-01-07,0,0,0
2020-01-08,0,0,0
2020-01-09,0,0,0
2020-01-10,0,0,0
2020-01-11,0,0,0
2020-01-12,0,0,0
2020-01-13,0,0,0
2020-01-14,0,0,0
2020-01-15,0,0,0
2020-01-16,0,0,0
2020-01-17,0,0,0
2020-01-18,0,0,0
2020-01-19,0,0,0
2020-01-20,0,0,0
2020-01-21,0,0,0
2020-01-22,0,0,0
2020-01-23,0,0,0
2020-01-24,0,0,0
2020-01-25,0,0,0
2020-01-26,0,0,0
2020-01-27,0,0,0
2020-01-28,0,0,0
2020-01-29,0,0,0
2020-01-30,0,0,0
2020-01-31,0,0,0
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			cmd.Config, err = config.NewStaticConfig(tt.config)
			require.NoError(t, err)

			fakeClient := new(configfakes.FakeMetricsClient)
			fakeClient.GetProjectReturns(&project, nil)
			fakeClient.GetProjectColumnsReturns(projectColumns)
			fakeClient.GetIssuesFromColumnReturns(repoIssuesMap)
			fakeClient.GetIssueEventsReturns(issueEvents, nil)
			cmd.Config.SetMetricsClient(fakeClient)
			cmd.Config.StartDate = time.Date(2020, 1, 1, 0, 0, 0, 0, time.Now().Location())
			cmd.Config.EndDate = cmd.Config.StartDate.AddDate(0, 1, 0)
			cmd.Config.EndColumnIndex = len(projectColumns) - 1
			cmd.Config.EndColumn = *projectColumns[cmd.Config.EndColumnIndex].Name

			output, err := executeCommand(cmd.GithubMetricsCmd, tt.args...)
			if tt.wantErr {
				assert.Error(t, err)
			}
			assert.Equal(t, tt.wantOutput, output)
		})
	}
}
