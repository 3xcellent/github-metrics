package models

// Organization - model for github.Organization
type Organization struct {
	Name     string
	ID       int64
	URL      string
	ReposURL string
}

// Organizations - slice of Organization
type Organizations []Organization
