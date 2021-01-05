package client

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/3xcellent/github-metrics/config"
	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

const StateArchived = "archived"
const StateAll = "all"

type MetricsClient struct {
	c *github.Client
}

func New(ctx context.Context, config config.APIConfig) (*MetricsClient, error) {
	baseURL := config.BaseURL
	if baseURL == "" {
		return nil, errors.New("github baseURL must be set")
	}

	uploadURL := config.UploadURL
	if len(uploadURL) == 0 {
		uploadURL = baseURL
	}

	token := config.Token
	if token == "" {
		return nil, errors.New("github access token not set")
	}

	client, err := github.NewEnterpriseClient(
		baseURL,
		uploadURL,
		oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})))
	if err != nil {
		return nil, err
	}

	return &MetricsClient{
		c: client,
	}, nil
}

func (m *MetricsClient) GetIssue(ctx context.Context, repoOwner, repoName string, issueNumber int) (*github.Issue, error) {
	issue, _, err := m.c.Issues.Get(ctx, repoOwner, repoName, issueNumber)
	if err != nil {
		return nil, err
	}
	return issue, nil
}

func (m *MetricsClient) GetProject(ctx context.Context, boardID int64) (*github.Project, error) {
	project, _, err := m.c.Projects.GetProject(ctx, boardID)
	return project, err
}

func (m *MetricsClient) GetProjectColumns(ctx context.Context, projectID int64) []*github.ProjectColumn {
	projectColumns := make([]*github.ProjectColumn, 0)
	nextPage := 1
	for nextPage != 0 {
		columnsForPage, resp, err := m.c.Projects.ListProjectColumns(ctx,
			projectID,
			&github.ListOptions{
				PerPage: 100,
				Page:    nextPage,
			})
		if err != nil {
			panic(err)
		}
		for _, col := range columnsForPage {
			logrus.Debugf("ProjectColumns: found \"%s\" - %5d", col.GetName(), col.GetID())
			projectColumns = append(projectColumns, col)
		}
		nextPage = resp.NextPage
	}

	return projectColumns
}

func (m *MetricsClient) GetPullRequests(ctx context.Context, repoOwner, repoName string) ([]*github.PullRequest, error) {
	repoPullRequests := make([]*github.PullRequest, 0)
	nextPage := 1
	for nextPage != 0 {
		prs, resp, err := m.c.PullRequests.List(ctx, repoOwner, repoName, &github.PullRequestListOptions{
			State: "all",
			//Head:        "",
			//Base:        "",
			//Sort:        "",
			//Direction:   "",
			ListOptions: github.ListOptions{
				Page:    nextPage,
				PerPage: 1000,
			},
		})
		if err != nil {
			return nil, err
		}
		logrus.Infof("%s/%d - %d", repoName, nextPage, len(prs))
		repoPullRequests = append(repoPullRequests, prs...)
		nextPage = resp.NextPage
	}
	return repoPullRequests, nil
}

func (m *MetricsClient) GetIssues(ctx context.Context, repoOwner string, repos []string, beginDate, endDate time.Time) []*github.Issue {
	issues := make([]*github.Issue, 0)

	for _, repo := range repos {
		logrus.Infof("getting issues for repo: %s", repo)
		nextPage := 1
		for nextPage != 0 {
			issuesForPage, resp, err := m.c.Issues.ListByRepo(ctx, repoOwner, repo, &github.IssueListByRepoOptions{
				//Milestone: "",
				State: StateAll,
				//Assignee:  "",
				//Creator:   "",
				//Mentioned:   "",
				//Labels:      nil,
				Since: beginDate,
				ListOptions: github.ListOptions{
					PerPage: 100,
					Page:    nextPage,
				},
			})
			if resp != nil && resp.StatusCode == 404 {
				logrus.Warnf("URL Not Found: %s", resp.Request.URL.String())
				nextPage = 0
				continue
			}
			if err != nil {
				panic(err)
			}

			for _, issue := range issuesForPage {
				repoName := repo
				issue.Repository = &github.Repository{Name: &(repoName)} // patch repo name on returned object to have later since they don't populate
				if issue.GetCreatedAt().After(endDate) {
					continue
				}
				logrus.Warnf("\tadding issue: %s/%d - %s", issue.GetRepository().GetName(), issue.GetNumber(), issue.GetTitle())
				issues = append(issues, issue)
			}

			nextPage = resp.NextPage
		}
	}
	return issues
}

func (m *MetricsClient) GetIssueEvents(ctx context.Context, repoOwner, repoName string, issueNumber int) ([]*github.IssueEvent, error) {
	issueEvents := make([]*github.IssueEvent, 0)
	nextPage := 1
	for nextPage != 0 {
		events, resp, err := m.c.Issues.ListIssueEvents(
			ctx,
			repoOwner,
			repoName,
			issueNumber,
			&github.ListOptions{
				Page: nextPage,
			},
		)
		if err != nil {
			return nil, err
		}
		issueEvents = append(issueEvents, events...)
		nextPage = resp.NextPage
	}
	return issueEvents, nil
}

// GetRepos - returns slice of repo names gathered from the issues found in the columnID provided.
func (m *MetricsClient) GetRepos(ctx context.Context, columnId int64) []string {
	repos := make([]string, 0)
	repoMap := map[string]struct{}{}
	state := StateAll
	nextPage := 1
	for nextPage != 0 {
		cards, resp, err := m.c.Projects.ListProjectCards(ctx, columnId, &github.ProjectCardListOptions{
			ArchivedState: &state,
			ListOptions: github.ListOptions{
				PerPage: 100,
				Page:    nextPage,
			},
		})
		if err != nil {
			panic(err)
		}

		for _, c := range cards {
			if c.GetContentURL() == "" {
				continue
			}

			_, repo, _ := ParseContentURL(c.GetContentURL())

			if _, found := repoMap[repo]; !found {
				repoMap[repo] = struct{}{}
				repos = append(repos, repo)
			}
		}
		nextPage = resp.NextPage
	}

	logrus.Debugf("\tusing repos found on project board: %s", strings.Join(repos, ","))
	return repos
}

// Issue Content URLs look like: https://github.platforms.engineering/api/v3/repos/GraphRoots/performance-master/issues/1014
// captures {$1:"Graphroots",$2:"performance-master",$3:"1014"}
const regexParseContentURL = `^.*\/repos\/(\S+)\/(\S+)\/issues\/(\d+)`

func ParseContentURL(url string) (owner string, repo string, issueNumber int) {
	re, err := regexp.Compile(regexParseContentURL)
	if err != nil {
		panic("unable to parse regex")
		//fmt.Printf("could not compile regex: %v", err)
	}
	matches := re.FindStringSubmatch(url)
	if len(matches) < 1 {
		logrus.Errorf("Error parsing url: %s", url)
	}
	number, err := strconv.Atoi(matches[3])
	if err != nil {
		panic(err)
	}
	return matches[1], matches[2], number
}
