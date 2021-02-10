package cmd

import (
	"context"

	"github.com/3xcellent/github-metrics/client"
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

	var err error
	if Config == nil {
		Config, err = config.NewDefaultConfig()
		if err != nil {
			return err
		}
	}

	runCfg, err := Config.GetRunConfig(args[0])
	if err != nil {
		return err
	}

	ctx := context.Background()
	ghClient, err := client.New(ctx, Config.API)
	if err != nil {
		panic(err)
	}

	project, err := ghClient.GetProject(ctx, runCfg.ProjectID)
	if err != nil {
		return err
	}

	c.Println("Name:\t", project.Name)
	c.Println("ID:\t", project.ID)
	c.Println("URL:\t", project.URL)
	c.Println("Owner:\t", project.OwnerURL)
	c.Println("Body:\t", project.Body)

	return nil
}
