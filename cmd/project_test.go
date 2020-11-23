package cmd_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/3xcellent/github-metrics/cmd"
	"github.com/3xcellent/github-metrics/config"
	"github.com/3xcellent/github-metrics/config/configfakes"
	"github.com/google/go-github/v32/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_project(t *testing.T) {
	name := "name"
	ID := int64(42)
	URL := "URL"
	OwnerURL := "OwnerURL"
	body := "body"
	project := &github.Project{
		Name:     &name,
		ID:       &ID,
		URL:      &URL,
		OwnerURL: &OwnerURL,
		Body:     &body,
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

	var usageMsg = fmt.Sprintf(`
Usage:
  github-metrics project [name] [flags]

Flags:
  -h, --help   help for project

Global Flags:
  -d, --askForDate        command will ask for user to input year and month parameters at runtime
  -c, --create-file       set outpath path to [board_name]_[command_name]_[year]_[month].csv)
  -i, --issueNumber int   issueNumber (use with issueNumber)
  -m, --month int         specify month (default %d)
      --no-headers        disable csv header row
  -o, --outpath string    set output path
  -r, --repoName string   repoName (use with repoName)
  -t, --token string      Auth token used when connecting to github server
  -v, --verbose           verbose output
  -y, --year int          specify year (default %d)

`, time.Now().Month(), time.Now().Year())

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
			args:       []string{"project"},
			wantErr:    true,
			wantOutput: "Error: requires at least 1 arg(s), only received 0" + usageMsg,
		},
		{
			name:       "returns error when no boards configured",
			config:     noBoardsYAML,
			args:       []string{"project", "board1"},
			wantErr:    true,
			wantOutput: "Error: no project boards configured" + usageMsg,
		},
		{
			name:    "happy path",
			config:  configYAML,
			args:    []string{"project", "board1"},
			wantErr: false,
			wantOutput: "Name:\t " + name + "\n" +
				"ID:\t " + fmt.Sprintf("%d", ID) + "\n" +
				"URL:\t " + URL + "\n" +
				"Owner:\t " + OwnerURL + "\n" +
				"Body:\t " + body + "\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			cmd.Config, err = config.NewStaticConfig(tt.config)
			require.NoError(t, err)

			fakeClient := new(configfakes.FakeMetricsClient)
			fakeClient.GetProjectReturns(project, nil)
			cmd.Config.SetMetricsClient(fakeClient)

			output, err := executeCommand(cmd.GithubMetricsCmd, tt.args...)
			if tt.wantErr {
				assert.Error(t, err)
			}
			assert.Equal(t, tt.wantOutput, output)
		})
	}
}
