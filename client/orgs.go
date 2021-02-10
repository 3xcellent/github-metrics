package client

import (
	"context"

	"github.com/3xcellent/github-metrics/models"
	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
)

// GetUserOrgs - returns gihub organizations by owner, returns nil, err on error
func (m *MetricsClient) GetUserOrgs(ctx context.Context, owner string) (models.Organizations, error) {
	orgs := make(models.Organizations, 0)
	opt := &github.ListOptions{PerPage: 100}

	for {
		pageOrgs, resp, err := m.c.Organizations.List(ctx, "", opt)
		if err != nil {
			return nil, err
		}

		for _, org := range pageOrgs {
			logrus.Debugf("User Org found: \"%s\" - %5d - %s", org.GetName(), org.GetID(), org.GetURL())
			orgs = append(orgs, mapToOrganzation(org))
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return orgs, nil
}

func mapToOrganzation(org *github.Organization) models.Organization {
	return models.Organization{
		Name:     org.GetName(),
		ID:       org.GetID(),
		URL:      org.GetURL(),
		ReposURL: org.GetReposURL(),
	}
}
