package client

import (
	"context"

	"github.com/3xcellent/github-metrics/models"
	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
)

// GetPullRequests - uses owner and reponame to retrieve []*github.PullRequest and map to models.PullRequests
func (m *MetricsClient) GetPullRequests(ctx context.Context, owner, repoName string) (models.PullRequests, error) {
	repoPullRequests := make(models.PullRequests, 0)
	opt := github.ListOptions{PerPage: 100}
	for {
		prs, resp, err := m.c.PullRequests.List(ctx, owner, repoName, &github.PullRequestListOptions{
			State: "all",
			//Head:        "",
			//Base:        "",
			//Sort:        "",
			//Direction:   "",
			ListOptions: github.ListOptions{
				Page:    opt.Page,
				PerPage: 1000,
			},
		})
		if err != nil {
			return nil, err
		}
		logrus.Debugf("%s/%d - %d", repoName, opt.Page, len(prs))
		pagedPRs := mapToPullRequests(prs, owner, repoName)
		repoPullRequests = append(repoPullRequests, pagedPRs...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return repoPullRequests, nil
}

func mapToPullRequest(pr *github.PullRequest, owner, repoName string) models.PullRequest {
	reviewers := make([]string, 0)
	for _, reviewer := range pr.RequestedReviewers {
		reviewers = append(reviewers, reviewer.GetLogin())
	}
	return models.PullRequest{
		ID:                 pr.GetID(),
		Owner:              owner,
		RepoName:           repoName,
		CreatedAt:          pr.GetCreatedAt(),
		ClosedAt:           pr.GetClosedAt(),
		CreatedByUser:      pr.GetUser().GetLogin(),
		IssueURL:           pr.GetIssueURL(),
		URL:                pr.GetURL(),
		RequestedReviewers: reviewers,
	}
}

func mapToPullRequests(prs []*github.PullRequest, owner, repoName string) models.PullRequests {
	pullRequests := make(models.PullRequests, 0)
	for _, pr := range prs {
		pullRequests = append(pullRequests, mapToPullRequest(pr, owner, repoName))
	}
	return pullRequests
}
