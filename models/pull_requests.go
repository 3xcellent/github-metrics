package models

import "time"

// PullRequest - model for github pullrequest
type PullRequest struct {
	Owner              string
	RepoName           string
	ID                 int64
	URL                string
	CreatedAt          time.Time
	ClosedAt           time.Time
	CreatedByUser      string
	IssueURL           string
	RequestedReviewers []string
}

// PullRequests - slice of []github.PullRequest
type PullRequests []PullRequest
