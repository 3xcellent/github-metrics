package client

import (
	"context"
	"regexp"

	"github.com/3xcellent/github-metrics/models"
	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
)

// GetProject - retrieves project from github and maps to models.Project, returns empty project, err on error
func (m *MetricsClient) GetProject(ctx context.Context, projectID int64) (models.Project, error) {
	logrus.Debugf("\tgetting project for id: %d", projectID)
	project, _, err := m.c.Projects.GetProject(ctx, projectID)
	if err != nil {
		return models.Project{}, err
	}
	logrus.Debugf("\tfound project: %q", project.GetName())
	return mapToProject(project), nil
}

// GetProjects - returns gihub projects by owner, returns nil, err on error
func (m *MetricsClient) GetProjects(ctx context.Context, owner string) (models.Projects, error) {
	projects, err := m.getUserProjects(ctx, owner)
	if err != nil {
		return nil, err
	}

	userProjects, err := m.getUserProjects(ctx, owner)
	if err != nil {
		return nil, err
	}

	projects = append(projects, userProjects...)

	repos, err := m.GetUserRepos(ctx, owner)
	if err != nil {
		return nil, err
	}
	for _, r := range repos {
		projectOwner, projectRepo := parseRepoURL(r.URL)
		logrus.Debugf("repo: %s/%s (%s)", projectOwner, projectRepo, r.URL)
		repoProjects, err := m.getRepoProjects(ctx, projectOwner, projectRepo)
		if err != nil {
			return nil, err
		}
		projects = append(projects, repoProjects...)
	}
	return projects, nil
}

func (m *MetricsClient) getOrgProjects(ctx context.Context, owner string) ([]*github.Project, error) {
	logrus.Debugf("getting org projects for: %s", owner)

	projects := make([]*github.Project, 0)
	opt := github.ListOptions{
		PerPage: 100,
	}

	for {
		projects, resp, err := m.c.Organizations.ListProjects(ctx, owner, &github.ProjectListOptions{ListOptions: opt})
		if err != nil {
			return nil, err
		}
		for _, p := range projects {
			logrus.Debugf("Organization Project (page %d): found \"%s\" - %5d", opt.Page, p.GetName(), p.GetID())
			projects = append(projects, p)
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return projects, nil
}

func (m *MetricsClient) getUserProjects(ctx context.Context, owner string) (models.Projects, error) {
	logrus.Debugf("getting projects for user: %s", owner)

	projects := make(models.Projects, 0)
	opt := github.ListOptions{PerPage: 100}

	for {
		projects, resp, err := m.c.Users.ListProjects(ctx, owner, &github.ProjectListOptions{ListOptions: opt})
		if err != nil {
			return nil, err
		}
		for _, p := range projects {
			logrus.Debugf("\tfound \"%s\" - %5d", p.GetName(), p.GetID())
			projects = append(projects, p)
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return projects, nil
}

func (m *MetricsClient) getRepoProjects(ctx context.Context, owner, repo string) (models.Projects, error) {
	logrus.Debugf("getting repo projects for: %s/%s", owner, repo)

	projects := make(models.Projects, 0)
	opt := github.ListOptions{PerPage: 100}

	for {

		pageProjects, resp, err := m.c.Repositories.ListProjects(ctx,
			owner,
			repo,
			&github.ProjectListOptions{ListOptions: opt})
		if err != nil {
			logrus.Warnf("err accessing projects for \"%s\"-\"%s\": %s", owner, repo, err.Error())
		}
		if resp.StatusCode != 200 {
			break
		}
		for _, p := range pageProjects {
			logrus.Debugf("\tfound: \"%s\" - %5d", p.GetName(), p.GetID())
			projects = append(projects, mapRepoProject(p, owner, repo))
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return projects, nil
}

func mapRepoProject(ghProject *github.Project, projectOwner, projectRepo string) models.Project {
	p := ghProject

	return models.Project{
		Name:     p.GetName(),
		ID:       p.GetID(),
		Owner:    projectOwner,
		OwnerURL: p.GetOwnerURL(),
		Body:     p.GetBody(),
		URL:      p.GetURL(),
		Repo:     projectRepo,
	}
}

func mapRepoProjects(ghProjects []*github.Project, projectOwner, projectRepo string) models.Projects {
	modelProjects := make(models.Projects, 0, len(ghProjects))
	for _, g := range ghProjects {
		p := g
		modelProjects = append(modelProjects, mapRepoProject(p, projectOwner, projectRepo))
	}
	return modelProjects
}

func mapToProject(project *github.Project) models.Project {
	project.GetOwnerURL()
	return models.Project{
		Name: project.GetName(),
		ID:   project.GetID(),
		URL:  project.GetURL(),
	}
}

// Project URLs look like: https://api.github.com/repos/3xcellent/github-metrics/
// captures {$1:"3xcellent",$2:"github-metrics"}
const regexParseRepoURL = `^.*\/repos\/(\S+)\/(\S+)(\/|\z)`

// ParseRepoURL -- returns owner, and repo from a Project URL
func parseRepoURL(url string) (owner string, repo string) {
	re, err := regexp.Compile(regexParseRepoURL)
	if err != nil {
		panic("unable to parse regex")
		//fmt.Printf("could not compile regex: %v", err)
	}
	matches := re.FindStringSubmatch(url)
	if len(matches) < 2 {
		logrus.Errorf("Error parsing url: %s", url)
	}
	return matches[1], matches[2]
}
