package gui

import (
	"context"
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

var (
	getProjectFromGithubButton widget.Clickable
	selectProjectButton        widget.Clickable

	isLoadingProjects bool
	hasLoadedProjects bool

	availableProjects     models.Projects
	availableProjectsEnum widget.Enum
)

// Projects holds the list of available projects provides a Layout
type Projects struct {
	Model models.Project
}

// func (projects *Projects) Layout(gtx C) D {

// }

func projectOptions(th *material.Theme, input *widget.Enum, projects models.Projects) []layout.FlexChild {
	options := make([]layout.FlexChild, 0, len(projects))

	for _, p := range projects {
		project := p
		options = append(options,
			layout.Rigid(func(gtx C) D {
				return material.RadioButton(
					th,
					input,
					fmt.Sprintf("%d", project.ID),
					project.Name,
				).Layout(gtx)
			}))
	}
	return options
}

// LayoutProjectsPage - layout of available projects
func LayoutProjectsPage(gtx C) D {
	if getProjectFromGithubButton.Clicked() && !isLoadingProjects {
		hasLoadedProjects = false
	}

	if !hasLoadedProjects {
		if !isLoadingProjects {
			isLoadingProjects = true
			go func(client *client.MetricsClient) {
				logrus.Info("getting projects...")
				ghProjects, err := client.GetProjects(context.Background(), State.APIConfig.Owner)
				if err != nil {
					panic(err)
				}
				availableProjects = ghProjects
				hasLoadedProjects = true
				isLoadingProjects = false
			}(State.Client)
		}
		op.InvalidateOp{}.Add(gtx.Ops)

		return layout.Flex{
			Alignment: layout.Start,
			Axis:      layout.Horizontal,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return inset.Layout(gtx, material.Body1(th, `Project Options:`).Layout)
			}),
			layout.Rigid(func(gtx C) D {
				if projectsEnum.Changed() {
					selectedBoardName = projectsEnum.Value
					op.InvalidateOp{}.Add(gtx.Ops)
				}
				return inset.Layout(
					gtx,
					func(gtx C) D {
						return layout.Flex{
							Axis: layout.Vertical,
						}.Layout(
							gtx,
							layout.Rigid(func(gtx C) D {
								return material.Body2(th, "Loading Projects...").Layout(gtx)
							}),
						)
					},
				)
			}),
		)
	}

	if projectsEnum.Changed() {
		id, err := strconv.Atoi(projectsEnum.Value)
		if err != nil {
			panic(err)
		}
		State.SelectedProjectID = int64(id)

		project, err := availableProjects.GetProject(State.SelectedProjectID)
		if err != nil {
			panic(err)
		}
		State.RunConfig.ProjectID = State.SelectedProjectID
		State.SelectedProjectName = project.Name
		nav.SetNavDestination(RunOptionsPage)
		op.InvalidateOp{}.Add(gtx.Ops)
	}

	logrus.Debugf("SelectedProjectID : %d", State.SelectedProjectID)
	logrus.Debugf("SelectedProjectName : %s", State.SelectedProjectName)
	return layout.Flex{
		Alignment: layout.Start,
		Axis:      layout.Horizontal,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(
				gtx,
				layout.Rigid(func(gtx C) D {
					return material.Button(th, &getProjectFromGithubButton, "Get Projects From Github").Layout(gtx)
				}),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return inset.Layout(gtx, material.Body1(th, `Project Options:`).Layout)
		}),
		layout.Rigid(func(gtx C) D {
			projectsEnum.Value = ""
			return inset.Layout(
				gtx,
				func(gtx C) D {
					return layout.Flex{
						Axis: layout.Vertical,
					}.Layout(
						gtx,
						layout.Rigid(func(gtx C) D {
							return material.Body2(th, "Project:").Layout(gtx)
						}),
						layout.Rigid(func(gtx C) D {
							return layout.Flex{
								Axis: layout.Vertical,
							}.Layout(
								gtx,
								projectOptions(th, &projectsEnum, availableProjects)...,
							)
						}),
					)
				},
			)
		}),
	)
}
