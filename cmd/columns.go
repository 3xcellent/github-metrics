package cmd

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/3xcellent/github-metrics/client"
	"github.com/3xcellent/github-metrics/config"
	"github.com/3xcellent/github-metrics/metrics"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const fmtDateKey = "2006-01-02"

var columnsCmd = &cobra.Command{
	Use:   "columns [board_name]",
	Short: "output number of issues in each column for a github board to csv",
	Long:  "aggregate column totals for a github repoName board within year and month provided (default is current year and month)",
	RunE:  columns,
	Args:  cobra.MinimumNArgs(1),
}

func columns(c *cobra.Command, args []string) error {
	ctx := context.Background()
	var err error
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

	board, err := Config.GetBoard(args[0])
	if err != nil {
		return err
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
			Config.StartDate.Month())
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

	logrus.Debugf("columns for %s: %s",
		project.GetName(),
		strings.Join(boardColumns.ColumnNames()[Config.StartColumnIndex:Config.EndColumnIndex], ","),
	)

	var issues metrics.Issues
	if Config.IssueNumber != 0 && Config.RepoName != "" {
		ghIssue, err := Config.GithubClient.GetIssue(ctx, Config.Owner, Config.RepoName, Config.IssueNumber)
		if err != nil {
			panic(err)
		}
		issue := &metrics.Issue{
			Owner:    Config.Owner,
			RepoName: Config.RepoName,
			Number:   ghIssue.GetNumber(),
			Title:    ghIssue.GetTitle(),
		}

		issue.ProcessLabels(ghIssue.Labels)
		issues = metrics.Issues{issue}
	} else {
		for repo, ghIssues := range Config.GithubClient.GetIssuesFromColumn(ctx,
			Config.Owner,
			boardColumns[Config.EndColumnIndex].ID,
			Config.StartDate,
			Config.EndDate,
		) {
			for _, ghIssue := range ghIssues {
				issue := &metrics.Issue{
					Owner:    Config.Owner,
					RepoName: repo,
					Number:   ghIssue.GetNumber(),
					Title:    ghIssue.GetTitle(),
				}

				issue.ProcessLabels(ghIssue.Labels)

				issues = append(issues, issue)
			}
		}
	}

	columns := metrics.NewColumns(board.BoardID, Config.StartDate, Config.EndDate, Config.EndColumn)
	for _, issue := range issues {
		events, err := Config.GithubClient.GetIssueEvents(ctx, Config.Owner, issue.RepoName, issue.Number)
		if err != nil {
			return err
		}
		columns.ProcessEvents(events)
	}

	var writer *csv.Writer
	if Config.OutputPath == "" {
		writer = csv.NewWriter(c.OutOrStdout())
	} else {
		writer = csv.NewWriter(Config.OutPath())
	}
	defer writer.Flush()

	if !Config.NoHeaders {
		logrus.Debugf("option: headers")
		headers := []string{"Day"}
		for i := Config.StartColumnIndex; i <= Config.EndColumnIndex; i++ {
			headers = append(headers, boardColumns[i].Name)
		}
		if err := writer.Write(headers); err != nil {
			logrus.Fatalf("error writing data to file")
		}
	}
	logrus.Debug(columns.Dump())

	for currentDate := Config.StartDate; currentDate.Before(Config.EndDate); currentDate = currentDate.AddDate(0, 0, 1) {
		logrus.Debugf("currentDate: %s", currentDate.String())
		dateRow := []string{currentDate.Format(fmtDateKey)}
		for i := Config.StartColumnIndex; i <= Config.EndColumnIndex; i++ {
			appendVal := "0"
			val, found := columns.DateColumn(currentDate, boardColumns[i].Name)
			if found {
				appendVal = strconv.Itoa(val)
			}
			dateRow = append(dateRow, appendVal)
		}
		if err := writer.Write(dateRow); err != nil {
			logrus.Fatalf("error writing data to file")
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
