package client

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/3xcellent/github-metrics/config"
	"github.com/3xcellent/github-metrics/models"
	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

const archived = "archived"
const all = "all"

// MetricsClient provides access to user datea through an authenticated github.Client connection
type MetricsClient struct {
	c *github.Client
}

// Client - wrapper for github api
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Client
type Client interface {
	GetIssue(ctx context.Context, repoOwner, repoName string, issueNumber int) (models.Issue, error)
	GetProject(ctx context.Context, projectID int64) (models.Project, error)
	GetProjects(ctx context.Context, owner string) (models.Projects, error)
	GetProjectColumns(ctx context.Context, projectID int64) (models.ProjectColumns, error)
	GetPullRequests(ctx context.Context, repoOwner, repoName string) (models.PullRequests, error)
	GetIssues(ctx context.Context, repoOwner string, reposNames []string, beginDate, endDate time.Time) models.Issues
	GetIssueEvents(ctx context.Context, repoOwner, repoName string, issueNumber int) (models.IssueEvents, error)
	GetRepos(ctx context.Context, columnID int64) ([]string, error)
}

func authenticateHTTPClient(ctx context.Context, token string) *http.Client {
	return oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}))
}

// New - returns a MetricsClient using the correct github.Client based on settings provided in config.APIConfig
func New(ctx context.Context, config config.APIConfig) (*MetricsClient, error) {
	if config.BaseURL != "" {
		logrus.Debugf("config.BaseURL is set, using enterprise client: %s", config.BaseURL)
		return enterpriseClient(ctx, config)
	}
	return defaultClient(ctx, config)
}
func defaultClient(ctx context.Context, config config.APIConfig) (*MetricsClient, error) {
	token := config.Token
	if token == "" {
		return nil, errors.New("github access token not set")
	}
	logrus.Debugf("creating new client with token: %s", token)

	client := github.NewClient(authenticateHTTPClient(ctx, token))

	return &MetricsClient{
		c: client,
	}, nil
}

func enterpriseClient(ctx context.Context, config config.APIConfig) (*MetricsClient, error) {
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
		authenticateHTTPClient(ctx, token),
	)
	if err != nil {
		return nil, err
	}

	return &MetricsClient{
		c: client,
	}, nil
}

// Issue URLs look like: https://api.github.com/repos/3xcellent/github-metrics/issues/2
// captures {$1:"3xcellent",$2:"github-metrics",$3:"2"}
const regexParseIssueURL = `^.*\/repos\/(\S+)\/(\S+)\/issues\/(\d+)`

// ParseIssueURL -- returns owner, repo, and issuenumber from an Issue URL
func ParseIssueURL(url string) (owner string, repo string, issueNumber int) {
	re, err := regexp.Compile(regexParseIssueURL)
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
