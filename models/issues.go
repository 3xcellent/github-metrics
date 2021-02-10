package models

import "time"

// Issue - model for github issue
type Issue struct {
	Owner     string
	RepoName  string
	Title     string
	Number    int
	Events    IssueEvents
	Labels    []string
	CreatedAt time.Time

	// Type        string
	// IsFeature   bool
	// BlockedTime time.Duration
	// DevTime     time.Duration
}

// Issues - slice of Issue
type Issues []Issue
