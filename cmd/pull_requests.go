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

	// gather repo list
	var repoList []string
	if len(repoNames) > 0 {
		repoList = strings.Split(repoNames, ",")
	} else {
		if Config.Boards == nil {
			return errors.New("no project boards configured")
		}
		board, found := Config.Boards[args[0]]
		if !found {
			return errors.New("no project board found with that name")
		}
		boardColumns := make(metrics.BoardColumns, 0)
		for _, col := range ghClient.GetProjectColumns(c.Context(), board.BoardID) {
			colName := col.GetName()

			boardColumns = append(boardColumns, &metrics.BoardColumn{Name: colName, ID: col.GetID()})
		}
		repoList = ghClient.GetReposFromIssuesOnColumn(c.Context(), boardColumns[Config.EndColumnIndex].ID)
	}

	for _, repo := range repoList {
		repo = strings.Trim(repo, " ")
		output, err := os.Create(fmt.Sprintf("%s_pullrequests.csv", repo))
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

		prs, err := ghClient.GetPullRequests(c.Context(), Config.Owner, repo)
		if err != nil {
			return err
		}
		for _, pr := range prs {
			if pr.GetClosedAt().IsZero() {
				// pr is still open
				continue
			}
			reviewerName := ""
			if len(pr.RequestedReviewers) > 0 {
				reviewers := make([]string, 0, len(pr.RequestedReviewers))
				for _, reviewer := range pr.RequestedReviewers {
					reviewers = append(reviewers, reviewer.GetLogin())
				}
				reviewerName = strings.Join(reviewers, "|")
			}

			_, repoName, issueNumber := client.ParseContentURL(pr.GetIssueURL())
			cols := []string{
				repoName,
				fmt.Sprintf("%d", issueNumber),
				pr.GetCreatedAt().String(),
				pr.GetUser().GetLogin(),
				Config.CreatedByGroup(pr.GetUser().GetLogin()),
				pr.GetClosedAt().String(),
				reviewerName,
				metrics.FmtDaysHours(pr.GetClosedAt().Sub(pr.GetCreatedAt())),
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
