package gui

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"git.sr.ht/~whereswaldon/materials"
	"github.com/3xcellent/github-metrics/config"
	"github.com/sirupsen/logrus"
)

var (
	connectionSettingsDoneButton widget.Clickable
	tokenInput                   materials.TextField
	ownerInput                   materials.TextField
	baseURLInput                 materials.TextField
	uploadURLInput               materials.TextField
)

// LayoutConnectionSettings - connection settings.
func LayoutConnectionSettings(gtx C) D {
	// set initial form values

	if connectionSettingsDoneButton.Clicked() {
		logrus.Debugf("connectionSettingsDoneButton.Clicked()")

		err := State.SetClient(config.APIConfig{
			Token:     tokenInput.Text(),
			Owner:     ownerInput.Text(),
			BaseURL:   baseURLInput.Text(),
			UploadURL: uploadURLInput.Text(),
		})
		if err != nil {
			panic(err)
		}
		State.HasValidatedConnection = true

		logrus.Debugf("updated connection settings")
		if State.SelectedProjectID == 0 || State.SelectedProjectName == "" {
			nav.SetNavDestination(ProjectsPage)
		}
		nav.SetNavDestination(RunOptionsPage)
	}
	return layout.Flex{
		Alignment: layout.Middle,
		Axis:      layout.Vertical,
	}.Layout(
		gtx,
		layout.Rigid(func(gtx C) D {
			tokenInput.Alignment = inputAlignment
			return tokenInput.Layout(gtx, th, "Token")
		}),
		layout.Rigid(func(gtx C) D {
			return inset.Layout(gtx, material.Body2(th, "Github Personal Access Token").Layout)
		}),
		layout.Rigid(func(gtx C) D {
			ownerInput.Alignment = inputAlignment
			return ownerInput.Layout(gtx, th, "Owner")
		}),
		layout.Rigid(func(gtx C) D {
			return inset.Layout(gtx, material.Body2(th, "Owner (username)").Layout)
		}),
		layout.Rigid(func(gtx C) D {
			baseURLInput.Alignment = inputAlignment
			return baseURLInput.Layout(gtx, th, "Base URL")
		}),
		layout.Rigid(func(gtx C) D {
			return inset.Layout(gtx, material.Body2(th, "Github Base URL (example: https://github.com)").Layout)
		}),
		layout.Rigid(func(gtx C) D {
			uploadURLInput.Alignment = inputAlignment
			return uploadURLInput.Layout(gtx, th, "Upload URL")
		}),
		layout.Rigid(func(gtx C) D {
			return inset.Layout(gtx, material.Body2(th, "Github Upload URL (only needed if different from base URL)").Layout)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(
				gtx,
				layout.Rigid(func(gtx C) D {
					return material.Button(th, &connectionSettingsDoneButton, "Done").Layout(gtx)
				}),
			)
		}),
	)
}
