package models

import "time"

// Issue - model for github issue
type Issue struct {
	Owner     string
	RepoName  string
	Title     string
	Number    int
	Labels    []string
	CreatedAt time.Time
	Events    IssueEvents
}

// Issues - slice of Issue
type Issues []Issue
