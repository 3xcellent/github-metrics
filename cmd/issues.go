package cmd

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/3xcellent/github-metrics/client"
	"github.com/3xcellent/github-metrics/config"
	"github.com/3xcellent/github-metrics/metrics"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var issuesCmd = &cobra.Command{
	Use:   "issues [board_name]",
	Short: "gathers metrics from issues on a board and outputs as csv",
	Long:  "gathers issues from a github repoName board, calculates column and blocked durations, and outputs as comma separated values (.csv)",
	RunE:  issues,
	Args:  cobra.MinimumNArgs(1),
}

func issues(c *cobra.Command, args []string) error {
	var err error
	ctx := context.Background()
	if Config == nil {
		Config, err = config.NewDefaultConfig()
		if err != nil {
			return err
		}
		Config.GithubClient, err = client.New(ctx, Config.API)
		if err != nil {
			return err
		}
	}

	if Config.Boards == nil {
		return errors.New("no project boards configured")
	}
	board, found := Config.Boards[args[0]]
	if !found {
		return errors.New("no project board found with that name")
	}

	project, err := Config.GithubClient.GetProject(ctx, board.BoardID)
	if err != nil {
		return err
	}

	if Config.CreateFile {
		Config.OutputPath = fmt.Sprintf("%s_%s_%d-%02d.csv",
			strings.Replace(project.GetName(), " ", "_", -1),
			c.Name(),
			Config.StartDate.Year(),
			Config.StartDate.Month(),
		)
	}

	boardColumns := make(metrics.BoardColumns, 0)
	for i, col := range Config.GithubClient.GetProjectColumns(ctx, board.BoardID) {
		colName := col.GetName()
		if colName == board.StartColumn {
			logrus.Debugf("StartColumnIndex: %d", i)
			Config.StartColumnIndex = i
		}

		if colName == board.EndColumn {
			logrus.Debugf("EndColumnIndex: %d", i)
			Config.EndColumnIndex = i
		}
		boardColumns = append(boardColumns, &metrics.BoardColumn{Name: colName, ID: col.GetID()})
	}

	if Config.EndColumnIndex == 0 {
		Config.EndColumn = boardColumns[0].Name
	}

	if Config.EndColumnIndex == 0 {
		Config.EndColumnIndex = len(boardColumns) - 1
		Config.EndColumn = boardColumns[Config.EndColumnIndex].Name
	}

	logrus.Debugf("columns for %s: %s", project.GetName(), strings.Join(boardColumns.ColumnNames()[Config.StartColumnIndex:Config.EndColumnIndex], ","))

	issues := metrics.Issues{}
	if Config.IssueNumber != 0 && Config.RepoName != "" {
		ghIssue, err := Config.GithubClient.GetIssue(ctx, Config.Owner, Config.RepoName, Config.IssueNumber)
		if err != nil {
			panic(err)
		}

		newColumnDates := make(metrics.BoardColumns, len(boardColumns))
		copy(newColumnDates, boardColumns)
		issue := &metrics.Issue{
			//Issue: ghIssue,
			Owner:            Config.Owner,
			RepoName:         Config.RepoName,
			Number:           ghIssue.GetNumber(),
			Type:             "",
			Title:            "",
			IsFeature:        false,
			TotalTimeBlocked: 0,
			ProjectID:        0,
		}

		issue.ProcessLabels(ghIssue.Labels)

		issue.ColumnDates = make(metrics.BoardColumns, len(boardColumns))
		for i, cd := range boardColumns {
			issue.ColumnDates[i] = &metrics.BoardColumn{Name: cd.Name}
		}

		events, err := Config.GithubClient.GetIssueEvents(ctx, Config.Owner, issue.RepoName, issue.Number)
		if err != nil {
			return err
		}

		issue.ProcessIssueEvents(events, Config.StartColumnIndex)

		issues = metrics.Issues{issue}
	} else {
		for repo, repoIssues := range Config.GithubClient.GetIssuesFromColumn(ctx,
			Config.Owner,
			boardColumns[Config.EndColumnIndex].ID,
			Config.StartDate,
			Config.EndDate,
		) {
			for _, repoIssue := range repoIssues {
				issue := &metrics.Issue{
					//Issue: issue,
					Owner:            Config.Owner,
					RepoName:         repo,
					Number:           repoIssue.GetNumber(),
					Title:            repoIssue.GetTitle(),
					IsFeature:        false,
					TotalTimeBlocked: 0,
					ProjectID:        0,
					Type:             "Enhancement",
				}

				issue.ProcessLabels(repoIssue.Labels)

				issue.ColumnDates = make(metrics.BoardColumns, len(boardColumns))
				for i, cd := range boardColumns {
					issue.ColumnDates[i] = &metrics.BoardColumn{Name: cd.Name}
				}

				events, err := Config.GithubClient.GetIssueEvents(ctx, Config.Owner, issue.RepoName, issue.Number)
				if err != nil {
					return err
				}

				issue.ProcessIssueEvents(events, Config.StartColumnIndex)

				issues = append(issues, issue)
			}

		}
	}

	var writer *csv.Writer
	if Config.OutputPath == "" {
		writer = csv.NewWriter(c.OutOrStdout())
	} else {
		writer = csv.NewWriter(Config.OutPath())
	}
	defer writer.Flush()

	for _, rowColumns := range issues.CSVRowColumns(Config.StartColumnIndex, Config.EndColumnIndex, board.BoardID, !Config.NoHeaders, Config.StartDate, Config.EndDate) {
		if err := writer.Write(rowColumns); err != nil {
			return err
		}
	}

	c.Println()
	if Config.OutputPath != "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		c.Printf("Wrote to: file://%s/%s\n", wd, Config.OutputPath)
	}
	return nil
}
