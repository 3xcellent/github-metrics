package cmd

import (
	"context"
	"errors"

	"github.com/3xcellent/github-metrics/client"
	"github.com/3xcellent/github-metrics/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var projectsCommand = &cobra.Command{
	Use:   "projects",
	Short: "lists available projects",
	Long:  "lists available projects",
	RunE:  getProjects,
}

func getProjects(c *cobra.Command, args []string) error {
	var err error
	if Config == nil {
		Config, err = config.NewDefaultConfig()
		if err != nil {
			return err
		}
	}

	logrus.Debugf("getting projects for owner: %s", Config.Owner)

	ctx := context.Background()
	ghClient, err := client.New(ctx, Config.API)
	if err != nil {
		panic(err)
	}

	if len(args) > 0 && args[0] != "" {
		Config.Owner = args[0]
	}

	if Config.Owner == "" {
		panic(errors.New("must provid owner"))
	}

	projects, err := ghClient.GetProjects(ctx, Config.Owner)
	if err != nil {
		return err
	}

	for _, p := range projects {
		c.Println("Name:\t", p.Name)
		c.Println("ID:\t", p.ID)
		c.Println("URL:\t", p.URL)
		c.Println("Owner:\t", p.Owner)
		c.Println("Repo:\t", p.Repo)
	}

	return nil
}
