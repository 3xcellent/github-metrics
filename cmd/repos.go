package cmd

import (
	"context"
	"fmt"

	"github.com/3xcellent/github-metrics/client"
	"github.com/3xcellent/github-metrics/config"
	"github.com/spf13/cobra"
)

var reposCommand = &cobra.Command{
	Use:   "repos",
	Short: "shows list of repos",
	Long:  "shows list of repos",
	RunE:  repos,
}

func repos(c *cobra.Command, args []string) error {
	var err error
	ctx := context.Background()

	if Config == nil {
		Config, err = config.NewDefaultConfig()
		if err != nil {
			return err
		}

	}

	client, err := client.New(ctx, Config.API)
	if err != nil {
		return err
	}

	repos, err := client.GetUserRepos(ctx, Config.Owner)
	if err != nil {
		return err
	}

	for _, r := range repos {
		fmt.Printf("name: %s", r.Name)
		fmt.Printf("id: %v", r.ID)
		fmt.Printf("url: %s", r.URL)
	}

	// repos := Config.GithubClient.GetReposFromIssuesOnColumn(ctx, boardColumns[Config.EndColumnIndex].ID)
	// c.Println(strings.Join(repos, ", "))

	return nil
}
