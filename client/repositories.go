package client

import (
	"context"
	"strings"

	"github.com/3xcellent/github-metrics/models"
	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
)

// GetReposFromProjectColumn - returns slice of repo names gathered from the issues found in the columnID provided.
func (m *MetricsClient) GetReposFromProjectColumn(ctx context.Context, colID int64) (models.Repositories, error) {
	repos := make(models.Repositories, 0)
	repoMap := map[string]models.Repository{}
	state := all
	opt := github.ListOptions{PerPage: 100}
	logrus.Debugf("getting repos for project: %d", colID)
	for {
		cards, resp, err := m.c.Projects.ListProjectCards(ctx, colID, &github.ProjectCardListOptions{
			ArchivedState: &state,
			ListOptions:   opt,
		})
		if err != nil {
			return nil, err
		}

		for _, c := range cards {
			if c.GetContentURL() == "" {
				continue
			}

			owner, repoName, _ := ParseIssueURL(c.GetContentURL())

			if _, found := repoMap[repoName]; !found {
				repo := models.Repository{
					Name:  repoName,
					Owner: owner,
				}
				repoMap[repoName] = repo
				logrus.Debugf("\tadding repo %s", repoName)
				repos = append(repos, repo)
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	logrus.Infof("repos found: %s", strings.Join(repos.Names(), ","))
	return repos, nil
}

// GetUserRepos - returns gihub repositories by owner, returns nil, err on error
func (m *MetricsClient) GetUserRepos(ctx context.Context, owner string) (models.Repositories, error) {
	repos := make(models.Repositories, 0)
	opt := github.ListOptions{PerPage: 100}
	for {
		pageRepos, resp, err := m.c.Repositories.List(ctx, "", &github.RepositoryListOptions{ListOptions: opt})

		if err != nil {
			return nil, err
		}
		for _, r := range pageRepos {
			logrus.Debugf("User Repo found: \"%s\" - %5d - %s", r.GetName(), r.GetID(), r.GetURL())
			repos = append(repos, mapToRepository(r))
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return repos, nil
}

func mapToRepository(repo *github.Repository) models.Repository {
	return models.Repository{
		Owner: repo.GetOwner().GetName(),
		Name:  repo.GetName(),
		ID:    repo.GetID(),
		URL:   repo.GetURL(),
	}
}
