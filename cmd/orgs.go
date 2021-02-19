package cmd

import (
	"context"

	"github.com/3xcellent/github-metrics/client"
	"github.com/3xcellent/github-metrics/config"
	"github.com/spf13/cobra"
)

var orgsCommand = &cobra.Command{
	Use:   "orgs",
	Short: "lists available orgs",
	Long:  "lists available orgs",
	RunE:  getOrgs,
}

func getOrgs(c *cobra.Command, args []string) error {
	var err error
	if Config == nil {
		Config, err = config.NewDefaultConfig()
		if err != nil {
			return err
		}
	}

	ctx := context.Background()
	ghClient, err := client.New(ctx, Config.API)
	if err != nil {
		panic(err)
	}

	orgs, err := ghClient.GetUserOrgs(ctx, Config.Owner)
	if err != nil {
		return err
	}

	for _, org := range orgs {
		c.Printf("Name:\t%s\n", org.Name)
		c.Println("ID:\t", org.ID)
		c.Println("URL:\t", org.URL)
		c.Println("repos url:\t", org.ReposURL)
	}

	return nil
}
