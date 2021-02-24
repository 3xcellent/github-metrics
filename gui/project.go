package gui

import (
	"context"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/3xcellent/github-metrics/client"
	"github.com/3xcellent/github-metrics/models"
	"github.com/sirupsen/logrus"
)

var (
	getReposFromGithubButton widget.Clickable
	reposEnum                widget.Enum
)

// Project provides a layout for a models.Project
type Project struct {
	Model        models.Project
	Repositories Repositories

	isLoadingRepos   bool
	hasLoadedRepos   bool
	selectedRepoName string
}

// Layout - layout of Project Details and Repositories
func (p *Project) Layout(gtx C) D {
	if getReposFromGithubButton.Clicked() && !isLoadingProjects {
		p.hasLoadedRepos = false
		p.selectedRepoName = ""
	}

	if !p.hasLoadedRepos {
		if !p.isLoadingRepos {
			p.isLoadingRepos = true
			go func(client *client.MetricsClient) {
				projectColumns, err := client.GetProjectColumns(context.Background(), State.SelectedProjectID)
				if err != nil {
					panic(err)
				}

				ghRepos, err := client.GetReposFromProjectColumn(context.Background(), projectColumns[len(projectColumns)-1].ID)
				if err != nil {
					panic(err)
				}

				p.Repositories = make(Repositories, 0, len(ghRepos))
				for _, ghRepo := range ghRepos {
					p.Repositories = append(p.Repositories, &Repository{Model: ghRepo})
				}
				p.hasLoadedRepos = true
				p.isLoadingRepos = false

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

	if reposEnum.Changed() {
		p.selectedRepoName = reposEnum.Value

		// nav.SetNavDestination(RunOptionsPage)
		// op.InvalidateOp{}.Add(gtx.Ops)

		logrus.Debugf("SelectedProjectID : %d", State.SelectedProjectID)
		logrus.Debugf("SelectedProjectName : %s", State.SelectedProjectName)
		reposEnum.Value = ""
	}

	if p.selectedRepoName != "" {
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
					return inset.Layout(gtx, material.Body1(th, `Project Options:`).Layout)
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
				p.Repositories.Options(th, &reposEnum)...,
			)
		}),
	)
}
