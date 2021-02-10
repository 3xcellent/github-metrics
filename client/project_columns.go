package client

import (
	"context"

	"github.com/3xcellent/github-metrics/models"
	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
)

// GetProjectColumns - uses the projectID to retrieve []*github.ProjectColumn and map to models.ProjectColumns
func (m *MetricsClient) GetProjectColumns(ctx context.Context, projectID int64) (models.ProjectColumns, error) {
	projectColumns := make(models.ProjectColumns, 0)
	opt := &github.ListOptions{PerPage: 100}
	logrus.Debugf("getting columns for project: %d", projectID)
	for {
		columnsForPage, resp, err := m.c.Projects.ListProjectColumns(ctx, projectID, opt)
		if err != nil {
			return nil, err
		}
		for _, col := range columnsForPage {
			logrus.Debugf("\tProjectColumn found: %q - %5d", col.GetName(), col.GetID())
			projectColumns = append(projectColumns, mapToProjectColumn(col))
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	logrus.Infof("\t\t%d columns found", len(projectColumns))
	return projectColumns, nil
}

func mapToProjectColumn(ghProjectColumn *github.ProjectColumn) models.ProjectColumn {
	return models.ProjectColumn{
		Name: ghProjectColumn.GetName(),
		ID:   ghProjectColumn.GetID(),
	}
}
