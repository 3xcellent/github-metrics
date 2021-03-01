package gui

import (
	"fmt"
	"strconv"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/3xcellent/github-metrics/client"
	"github.com/3xcellent/github-metrics/models"
	"github.com/sirupsen/logrus"
)


type Repository struct {
	Model models.Repository

	hasLoadedIssues     bool
	isLoadingIssues     bool
	selectedIssueNumber int
}
type Repositories []*Repository

// Layout - layout of Project Details and Repositories
func (r *Repository) Layout(gtx C) D {
	if getReposFromGithubButton.Clicked() && !isLoadingProjects {
		r.hasLoadedIssues = false
		r.selectedIssueNumber = 0
	}

	if !r.hasLoadedIssues {
		if !r.isLoadingIssues {
			r.isLoadingIssues = true
			go func(client *client.MetricsClient) {
				// ghIssues, err := client.GetIssues(context.Background(), r.Model.Owner, []string{r.Model.Name}, time.Time{}, time.Now())
				// if err != nil {
				// 	panic(err)
				// }

				// r.Issues = make(Issues, 0, len(ghIssues))
				// for _, ghRepo := range ghRepos {
				// 	p.Repositories = append(p.Repositories, Repository{Model: ghRepo})
				// }
				r.hasLoadedIssues = true
				r.isLoadingIssues = false

				op.InvalidateOp{}.Add(gtx.Ops)
			}(State.Client)
		}

		return layout.Flex{
			Alignment: layout.Start,
			Axis:      layout.Horizontal,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return inset.Layout(gtx, material.Body1(th, `Repositories:`).Layout)
			}),
			layout.Rigid(func(gtx C) D {
				return material.Body2(th, "Loading Repositories...").Layout(gtx)
			}),
		)
	}

	if issuesEnum.Changed() {
		issueNum, err := strconv.Atoi(issuesEnum.Value)
		if err != nil {
			panic(err)
		}
		r.selectedIssueNumber = issueNum

		// nav.SetNavDestination(RunOptionsPage)
		// op.InvalidateOp{}.Add(gtx.Ops)

		logrus.Debugf("SelectedProjectID : %d", State.SelectedProjectID)
		logrus.Debugf("SelectedProjectName : %s", State.SelectedProjectName)
		reposEnum.Value = ""
	}

	if r.selectedIssueNumber != 0 {
		// project, err := availableProjects.GetProject(State.SelectedProjectID)
		// if err != nil {
		// 	panic(err)
		// }

		// State.RunConfig.ProjectID = State.SelectedProjectID
		// State.SelectedProjectName = project.Model.Name

		// return project.Layout(gtx)
	}
	return layout.Flex{
		Alignment: layout.Start,
		Axis:      layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Horizontal,
			}.Layout(
				gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return inset.Layout(gtx, material.Body1(th, `Issues:`).Layout)
				}),
				layout.Rigid(func(gtx C) D {
					return material.Button(th, &getProjectsFromGithubButton, "Get from Server").Layout(gtx)
				}),
			)
		}),
		layout.Rigid(func(gtx C) D {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(
				gtx,
				// p.Repositories.Options(th, &reposEnum)...,
			)
		}),
	)
}

// Options returns list of repositories as options
func (repos Repositories) Options(th *material.Theme, input *widget.Enum) []layout.FlexChild {
	options := make([]layout.FlexChild, 0, len(repos))
	for _, r := range repos {
		repo := r
		options = append(options,
			layout.Rigid(func(gtx C) D {
				return material.RadioButton(
					th,
					input,
					fmt.Sprintf("%s", repo.Model.Name),
					repo.Model.Name,
				).Layout(gtx)
			}))
	}
	return options
}
