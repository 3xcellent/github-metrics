package gui

import (
	"gioui.org/layout"
	"gioui.org/widget/material"
)

// local pages vars

var (
	isRunning          bool
)

// LayoutResults - results of the columns metric
func LayoutResults(gtx C) D {
	// set initial form values

	return layout.Flex{
		Alignment: layout.Middle,
		Axis:      layout.Vertical,
	}.Layout(
		gtx,
		layout.Rigid(func(gtx C) D {
			return inset.Layout(gtx, material.H1(th, "Results").Layout)
		}),
		layout.Rigid(func(gtx C) D {
			if isRunning {
				return inset.Layout(gtx, material.Body2(th, "running...").Layout)
			}
			return inset.Layout(gtx, material.Body2(th, outputText).Layout)
		}),
	)
}
