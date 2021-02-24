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
	"github.com/sirupsen/logrus"
)

var (
	getProjectsFromGithubButton widget.Clickable

	isLoadingProjects bool
	hasLoadedProjects bool

	availableProjects Projects
	projectsEnum      widget.Enum
)

// Projects holds the list of available projects provides a Layout
type Projects []*Project

// Options returns the list of available projects as []FlexChild
func (projects Projects) Options(th *material.Theme, input *widget.Enum) []layout.FlexChild {
	options := make([]layout.FlexChild, 0, len(projects))
	for _, p := range projects {
		project := p
		options = append(options,
			layout.Rigid(func(gtx C) D {
				return material.RadioButton(
					th,
					input,
					fmt.Sprintf("%d", project.Model.ID),
					project.Model.Name,
				).Layout(gtx)
			}))
	}
	return options
}

// GetProject - returns project found by id or error
func (projects Projects) GetProject(id int64) (*Project, error) {
	for _, proj := range projects {
		if proj.Model.ID == id {
			return proj, nil
		}
	}
	return &Project{}, fmt.Errorf("no project found with id %d", id)
}

// LayoutProjectsPage - layout of available projects
func LayoutProjectsPage(gtx C) D {
	if getProjectsFromGithubButton.Clicked() && !isLoadingProjects {
		hasLoadedProjects = false
	}

	if !hasLoadedProjects {
		if !isLoadingProjects {
			isLoadingProjects = true
			go func(client *client.MetricsClient) {
				ghProjects, err := client.GetProjects(context.Background(), State.APIConfig.Owner)
				if err != nil {
					panic(err)
				}

				availableProjects = make(Projects, 0, len(ghProjects))
				for _, ghProject := range ghProjects {
					availableProjects = append(availableProjects, &Project{Model: ghProject})
				}
				hasLoadedProjects = true
				isLoadingProjects = false
			}(State.Client)
		}
		op.InvalidateOp{}.Add(gtx.Ops)

		return layout.Flex{
			Alignment: layout.Start,
			Axis:      layout.Horizontal,
		}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return material.Body2(th, "Loading Projects...").Layout(gtx)
			}),
		)
	}

	if projectsEnum.Changed() {
		id, err := strconv.Atoi(projectsEnum.Value)
		if err != nil {
			panic(err)
		}
		State.SelectedProjectID = int64(id)

		// nav.SetNavDestination(RunOptionsPage)
		// op.InvalidateOp{}.Add(gtx.Ops)

		logrus.Debugf("SelectedProjectID : %d", State.SelectedProjectID)
		logrus.Debugf("SelectedProjectName : %s", State.SelectedProjectName)
		projectsEnum.Value = ""
	}

	if State.SelectedProjectID != 0 {
		project, err := availableProjects.GetProject(State.SelectedProjectID)
		if err != nil {
			panic(err)
		}
		State.RunConfig.ProjectID = State.SelectedProjectID
		State.SelectedProjectName = project.Model.Name

		return project.Layout(gtx)
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
				availableProjects.Options(th, &projectsEnum)...,
			)
		}),
	)
}
