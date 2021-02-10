package models

// Repository - model used for github Repositories
type Repository struct {
	Owner string
	Name  string
	ID    int64
	URL   string
}

// Repositories - slice of Repository
type Repositories []Repository
