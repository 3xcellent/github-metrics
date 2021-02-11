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

// Names - returns the names from the slice of Repositories
func (repos Repositories) Names() []string {
	repoList := make([]string, 0, len(repos))
	for _, r := range repos {
		repoList = append(repoList, r.Name)
	}
	return repoList
}
