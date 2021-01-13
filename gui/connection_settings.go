package gui

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/sirupsen/logrus"
)

var (
	connectionSettionsDoneButton widget.Clickable
)

const (
	MainPage = iota
	ConnectionSettingsPage
)

func LayoutConnectionSettings(gtx C) D {
	if connectionSettionsDoneButton.Clicked() {
		logrus.Infof("connectionSettionsDoneButton.Clicked()")
		nav.SetNavDestination(MainPage)
	}
	return layout.Flex{
		Alignment: layout.Middle,
		Axis:      layout.Vertical,
	}.Layout(
		gtx,
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
		layout.Rigid(func(gtx C) D {
			tokenInput.Alignment = inputAlignment
			return tokenInput.Layout(gtx, th, "Token")
		}),
		layout.Rigid(func(gtx C) D {
			return inset.Layout(gtx, material.Body2(th, "Github Personal Access Token").Layout)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(
				gtx,
				layout.Rigid(func(gtx C) D {
					return material.Button(th, &connectionSettionsDoneButton, "Done").Layout(gtx)
				}),
			)
		}),
	)
}
