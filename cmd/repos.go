package cmd

import (
	"context"
	"strings"

	"github.com/3xcellent/github-metrics/client"
	"github.com/3xcellent/github-metrics/config"
	"github.com/3xcellent/github-metrics/metrics"
	"github.com/spf13/cobra"
)

var reposCommand = &cobra.Command{
	Use:   "repos [project]",
	Short: "shows list of repos for a specific project",
	Long:  "shows list of repos for a specific project",
	RunE:  repos,
	Args:  cobra.MinimumNArgs(1),
}

func repos(c *cobra.Command, args []string) error {
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

	board, err := Config.GetBoard(args[0])
	if err != nil {
		return err
	}

	boardColumns := make(metrics.BoardColumns, 0)
	for _, col := range Config.GithubClient.GetProjectColumns(ctx, board.BoardID) {
		colName := col.GetName()
		boardColumns = append(boardColumns, &metrics.BoardColumn{Name: colName, ID: col.GetID()})
	}

	repos := Config.GithubClient.GetReposFromIssuesOnColumn(ctx, boardColumns[Config.EndColumnIndex].ID)
	c.Println(strings.Join(repos, ", "))

	return nil
}
