package cmd

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/3xcellent/github-metrics/client"
	"github.com/3xcellent/github-metrics/config"
	"github.com/3xcellent/github-metrics/metrics"
	"github.com/3xcellent/github-metrics/models"
	"github.com/spf13/cobra"
)

var (
	pullRequestsCmd = &cobra.Command{
		Use:   "pull_requests [project]",
		Short: "output number of issues in each column for a github board to csv",
		Long:  "aggregate duration pull_request are open for the list of github repos either using --repoName=repo1,repo2 flag or name of the board to gather all repos within year and month provided (default is current year and month)",
		RunE:  pullRequests,
	}
	repoNames string
)

func init() {
	pullRequestsCmd.Flags().StringVarP(&repoNames, "repoNames", "r", "", "list of repos to generate reports for (repo1,repo2)")
}

func pullRequests(c *cobra.Command, args []string) error {
	var err error
	if Config == nil {
		Config, err = config.NewDefaultConfig()
		if err != nil {
			return err
		}
	}

	if len(args) == 0 && repoNames == "" {
		return errors.New("project name or --repoName required")
	}
	ghClient, err := client.New(c.Context(), Config.API)
	if err != nil {
		return err
	}

	runCfg, err := Config.GetRunConfig(args[0])
	if err != nil {
		return err
	}

	projectColumns, err := ghClient.GetProjectColumns(c.Context(), runCfg.ProjectID)
	if err != nil {
		return err
	}

	var repoList models.Repositories
	if len(repoNames) > 0 {
		for _, repoName := range strings.Split(repoNames, ",") {
			repoList = append(repoList, models.Repository{
				Name:  repoName,
				Owner: runCfg.Owner,
			})
		}
	} else {
		repoList, err = ghClient.GetReposFromProjectColumn(c.Context(), projectColumns[len(projectColumns)-1].ID)
		if err != nil {
			return err
		}
	}

	for _, repo := range repoList {
		repoName := strings.Trim(repo.Name, " ")
		output, err := os.Create(fmt.Sprintf("%s_pullrequests.csv", repoName))
		writer := csv.NewWriter(output)

		cols := []string{
			"repo",
			"issueNumber",
			"CreatedAt",
			"CreatedBy",
			"Group",
			"ClosedAt",
			"RequestedReviewers",
			"Days Open",
		}

		err = writer.Write(cols)
		if err != nil {
			panic(err)
		}

		prs, err := ghClient.GetPullRequests(c.Context(), Config.Owner, repo.Name)
		if err != nil {
			return err
		}
		for _, pr := range prs {
			if pr.ClosedAt.IsZero() {
				// pr is still open
				continue
			}

			_, repoName, issueNumber := client.ParseIssueURL(pr.IssueURL)
			cols := []string{
				repoName,
				fmt.Sprintf("%d", issueNumber),
				pr.CreatedAt.String(),
				pr.CreatedByUser,
				Config.CreatedByGroup(pr.CreatedByUser),
				pr.ClosedAt.String(),
				strings.Join(pr.RequestedReviewers, ","),
				metrics.FmtDaysHours(pr.ClosedAt.Sub(pr.CreatedAt)),
			}
			err := writer.Write(cols)
			if err != nil {
				panic(err)
			}
		}
		writer.Flush()
	}
	return nil
}
