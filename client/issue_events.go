package client

import (
	"context"
	"strings"

	"github.com/3xcellent/github-metrics/models"
	"github.com/google/go-github/v32/github"
)

// GetIssueEvents - uses owner and repo and issue number to retrieve []*github.IssueEvent and map to models.IssueEvents
func (m *MetricsClient) GetIssueEvents(ctx context.Context, repoOwner, repoName string, issueNumber int) (models.IssueEvents, error) {
	issueEvents := make(models.IssueEvents, 0)
	opt := &github.ListOptions{PerPage: 100}

	for {
		events, resp, err := m.c.Issues.ListIssueEvents(ctx, repoOwner, repoName, issueNumber, opt)
		if err != nil {
			return nil, err
		}
		issueEvents = append(issueEvents, MapToIssueEvents(events)...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return issueEvents, nil
}

// MapToIssueEvents - maps []*github.IssueEvent to models.IssueEvents
func MapToIssueEvents(issuesEvents []*github.IssueEvent) models.IssueEvents {
	events := make(models.IssueEvents, 0)

	for _, e := range issuesEvents {
		events = append(events, MapToIssueEvent(e))
	}
	return events
}

// MapToIssueEvent - maps *github.IssueEvent to models.IssueEvent
func MapToIssueEvent(e *github.IssueEvent) models.IssueEvent {
	return models.IssueEvent{
		Event:              e.GetEvent(), // TODO: remove since using Type throughout
		ProjectID:          e.GetProjectCard().GetProjectID(),
		Type:               models.IssueEventType(strings.ToUpper(e.GetEvent())),
		ColumnName:         e.GetProjectCard().GetColumnName(),
		PreviousColumnName: e.GetProjectCard().GetPreviousColumnName(),
		Label:              e.GetLabel().GetName(),
		Assignee:           e.GetAssignee().GetLogin(),
		Note:               e.GetProjectCard().GetNote(),
		LoginName:          e.GetActor().GetLogin(),
		CreatedAt:          e.GetCreatedAt().Local(),
	}
}
