package models

import (
	"fmt"

	"github.com/3xcellent/github-metrics/config"
)

// Project - model used for processing metrics
type Project struct {
	Name     string
	ID       int64
	Owner    string
	OwnerURL string
	Body     string
	Repo     string
	URL      string
}

// RunConfig - returns a new config.RunConfig with Name, ProjectID, and Owner
func (p *Project) RunConfig() config.RunConfig {
	return config.RunConfig{
		Name:      p.Name,
		ProjectID: p.ID,
		Owner:     p.Owner,
	}
}

// Projects - group pf Projects
type Projects []Project

// GetProject - returns project found by id or error
func (p Projects) GetProject(id int64) (Project, error) {
	for _, proj := range p {
		if proj.ID == id {
			return proj, nil
		}
	}
	return Project{}, fmt.Errorf("no project found with id %d", id)
}
