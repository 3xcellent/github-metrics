package gui

import (
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/sirupsen/logrus"
)

var (
	selectProjectButton   widget.Clickable
	availableProjectsEnum widget.Enum
)

func LayoutProjectOptions(gtx C) D {
	if availableProjectsEnum.Changed() {
		logrus.Infof("availableProjectsEnum.Changed()")
		op.InvalidateOp{}.Add(gtx.Ops)
	}
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
							return material.Body2(th, "Project:").Layout(gtx)
						}),
						layout.Rigid(func(gtx C) D {
							return layout.Flex{
								Axis: layout.Vertical,
							}.Layout(
								gtx,
								projectBoardOptions(th, &projectsEnum, Config.GetSortedBoards())...,
							)
						}),
					)
				},
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {

			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(
				gtx,
				layout.Rigid(func(gtx C) D {
					return material.Button(th, &selectProjectButton, "Select").Layout(gtx)
				}),
			)
		}),
	)
}
