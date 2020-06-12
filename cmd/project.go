package cmd

import (
	"context"

	"github.com/3xcellent/github-metrics/config"
	"github.com/spf13/cobra"
)

var projectCommand = &cobra.Command{
	Use:   "project [name]",
	Short: "shows information about a specific project",
	Long:  "shows information about a specific project",
	RunE:  getProject,
	Args:  cobra.MinimumNArgs(1),
}

func getProject(c *cobra.Command, args []string) error {
	ctx := context.Background()

	var err error
	if Config == nil {
		Config, err = config.NewDefaultConfig()
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

	c.Println("Name:\t", project.GetName())
	c.Println("ID:\t", project.GetID())
	c.Println("URL:\t", project.GetURL())
	c.Println("Owner:\t", project.GetOwnerURL())
	c.Println("Body:\t", project.GetBody())

	return nil
}
