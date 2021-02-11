package client

import (
	"context"
	"time"

	"github.com/3xcellent/github-metrics/models"
	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
)

// GetIssues - uses owner, list of repo names, begindate and enddates to retrieve []*github.Issue and map to models.Issues
func (m *MetricsClient) GetIssues(ctx context.Context, repoOwner string, repos []string, beginDate, endDate time.Time) (models.Issues, error) {
	projectIssues := make(models.Issues, 0)
	for _, repo := range repos {
		repoIssues := make(models.Issues, 0)
		logrus.Debugf("getting issues for repo: %s", repo)
		opt := github.ListOptions{PerPage: 100}
		for {
			issuesForPage, resp, err := m.c.Issues.ListByRepo(ctx, repoOwner, repo, &github.IssueListByRepoOptions{
				//Milestone: "",
				State: all,
				//Assignee:  "",
				//Creator:   "",
				//Mentioned:   "",
				//Labels:      nil,
				Since:       beginDate,
				ListOptions: opt,
			})
			if err != nil {
				return nil, err
			}
			if resp != nil && resp.StatusCode == 404 {
				logrus.Warnf("URL Not Found: %s", resp.Request.URL.String())
				continue
			}
			logrus.Debugf("retrieved %d issue for page %d", len(issuesForPage), opt.Page)
			for _, issue := range issuesForPage {
				issue := mapToIssue(issue)
				issue.RepoName = repo
				issue.Owner = repoOwner
				if issue.CreatedAt.After(endDate) {
					logrus.Debugf("skipping.... issue.CreatedAt: %s - After(%s)", issue.CreatedAt.String(), endDate.String())
					continue
				}
				logrus.Debugf("\tadding issue: %s/%d - %s", issue.RepoName, issue.Number, issue.Title)
				repoIssues = append(repoIssues, issue)
			}

			if resp.NextPage == 0 {
				break
			}
			opt.Page = resp.NextPage
		}
		logrus.Debugf("repo %s has %d issues", repo, len(repoIssues))
		projectIssues = append(projectIssues, repoIssues...)
	}
	return projectIssues, nil
}

// GetIssue - uses owner, repo and issue number to retrieve *gitthub.Issue and map to models.Issue.
// Returns empty issue, err when not found.
func (m *MetricsClient) GetIssue(ctx context.Context, repoOwner, repoName string, issueNumber int) (models.Issue, error) {
	ghIssue, _, err := m.c.Issues.Get(ctx, repoOwner, repoName, issueNumber)
	issue := mapToIssue(ghIssue)
	issue.Owner = repoOwner
	issue.RepoName = repoName
	issue.Number = issueNumber
	return issue, err
}

func mapToIssue(ghIssue *github.Issue) models.Issue {
	labels := make([]string, 0)
	for _, l := range ghIssue.Labels {
		labels = append(labels, l.GetName())
	}
	return models.Issue{
		Title:     ghIssue.GetTitle(),
		Number:    ghIssue.GetNumber(),
		CreatedAt: ghIssue.GetCreatedAt(),
		Labels:    labels,
	}
}
