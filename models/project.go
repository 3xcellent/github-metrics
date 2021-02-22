package models

import (
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
