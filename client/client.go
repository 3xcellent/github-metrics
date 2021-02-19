package client

import (
	"context"
	"errors"
	"regexp"
	"strconv"

	"github.com/3xcellent/github-metrics/config"
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

// errors
const (
	ErrAccessTokenNotSet = "github access token not set"
)

// New will return a MetricsClient using the token provided in config.APIConfig.  For enterprise
// servers, provide the BaseURL (UploadURL defaults to default for BaseURL when blank)
func New(ctx context.Context, config config.APIConfig) (*MetricsClient, error) {
	token := config.Token
	if token == "" {
		return nil, errors.New(ErrAccessTokenNotSet)
	}

	authenticatedClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}))
	if config.BaseURL == "" {
		logrus.Debugf("creating new client with token: %s", token)
		return &MetricsClient{c: github.NewClient(authenticatedClient)}, nil
	}

	if config.UploadURL == "" {
		config.UploadURL = config.BaseURL
	}

	logrus.Debug("using enterprise client")
	logrus.Debugf("\tToken: %s", token)
	logrus.Debugf("\tBaseURL: %s", config.BaseURL)
	logrus.Debugf("\tUploadURL: %s", config.UploadURL)
	client, err := github.NewEnterpriseClient(config.BaseURL, config.UploadURL, authenticatedClient)
	if err != nil {
		return nil, err
	}

	return &MetricsClient{c: client}, nil
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
