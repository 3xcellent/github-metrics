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
	cfg, err := config.NewDefaultConfig()
	if err != nil {
		return err
	}
	cfg.GithubClient, err = client.New(ctx, cfg.API)
	if err != nil {
		return err
	}

	if cfg.Boards == nil {
		return errors.New("no project boards configured")
	}
	board, found := cfg.Boards[args[0]]
	if !found {
		return errors.New("no project board found with that name")
	}

	project, err := cfg.GithubClient.GetProject(ctx, board.BoardID)
	if err != nil {
		return err
	}

	if cfg.CreateFile {
		cfg.OutputPath = fmt.Sprintf("%s_%s_%d-%02d.csv",
			strings.Replace(project.GetName(), " ", "_", -1),
			c.Name(),
			cfg.StartDate.Year(),
			cfg.StartDate.Month(),
		)
	}

	boardColumns := make(metrics.BoardColumns, 0)
	for i, col := range cfg.GithubClient.GetProjectColumns(ctx, board.BoardID) {
		colName := col.GetName()
		if colName == board.StartColumn {
			logrus.Debugf("StartColumnIndex: %d", i)
			cfg.StartColumnIndex = i
		}

		if colName == board.EndColumn {
			logrus.Debugf("EndColumnIndex: %d", i)
			cfg.EndColumnIndex = i
		}
		boardColumns = append(boardColumns, &metrics.BoardColumn{Name: colName, ID: col.GetID()})
	}

	if cfg.EndColumnIndex == 0 {
		cfg.EndColumn = boardColumns[0].Name
	}

	if cfg.EndColumnIndex == 0 {
		cfg.EndColumnIndex = len(boardColumns) - 1
		cfg.EndColumn = boardColumns[cfg.EndColumnIndex].Name
	}

	logrus.Debugf("columns for %s: %s", project.GetName(), strings.Join(boardColumns.ColumnNames()[cfg.StartColumnIndex:cfg.EndColumnIndex], ","))

	issues := metrics.Issues{}
	if cfg.IssueNumber != 0 && cfg.RepoName != "" {
		ghIssue, err := cfg.GithubClient.GetIssue(ctx, cfg.Owner, cfg.RepoName, cfg.IssueNumber)
		if err != nil {
			panic(err)
		}

		newColumnDates := make(metrics.BoardColumns, len(boardColumns))
		copy(newColumnDates, boardColumns)
		issue := &metrics.Issue{
			//Issue: ghIssue,
			Owner:            cfg.Owner,
			RepoName:         cfg.RepoName,
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

		events, err := cfg.GithubClient.GetIssueEvents(ctx, cfg.Owner, issue.RepoName, issue.Number)
		if err != nil {
			return err
		}

		issue.ProcessIssueEvents(events, cfg.StartColumnIndex)

		issues = metrics.Issues{issue}
	} else {
		repos := cfg.GithubClient.GetRepos(ctx, boardColumns[cfg.EndColumnIndex].ID)
		for _, repoIssue := range cfg.GithubClient.GetIssues(ctx,
			cfg.Owner,
			repos,
			cfg.StartDate,
			cfg.EndDate,
		) {
			issue := &metrics.Issue{
				//Issue: issue,
				Owner:            cfg.Owner,
				RepoName:         repoIssue.GetRepository().GetName(),
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

			events, err := cfg.GithubClient.GetIssueEvents(ctx, cfg.Owner, issue.RepoName, issue.Number)
			if err != nil {
				return err
			}

			issue.ProcessIssueEvents(events, cfg.StartColumnIndex)

			issues = append(issues, issue)
		}

	}

	var writer *csv.Writer
	if cfg.OutputPath == "" {
		writer = csv.NewWriter(c.OutOrStdout())
	} else {
		writer = csv.NewWriter(cfg.OutPath())
	}
	defer writer.Flush()

	for _, rowColumns := range issues.CSVRowColumns(cfg.StartColumnIndex, cfg.EndColumnIndex, board.BoardID, !cfg.NoHeaders, cfg.StartDate, cfg.EndDate) {
		if err := writer.Write(rowColumns); err != nil {
			return err
		}
	}

	c.Println()
	if cfg.OutputPath != "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		c.Printf("Wrote to: file://%s/%s\n", wd, cfg.OutputPath)
	}
	return nil
}
