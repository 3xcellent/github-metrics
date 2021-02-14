package gui

import (
	"time"

	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/3xcellent/github-metrics/metrics/runners"
)

// local pages vars

var (
	isRunning bool
)

// LayoutResults - results of the columns metric
func LayoutResults(gtx C) D {
	// set initial form values
	if State.RunRequested && !State.RunStarted {
		State.RunStarted = true
		selectedProject, err := availableProjects.GetProject(State.SelectedProjectID)
		if err != nil {
			panic(err)
		}
		State.RunConfig.ProjectID = selectedProject.ID
		State.RunConfig.Owner = selectedProject.Owner
		State.RunConfig.StartDate = time.Date(selectedYear, time.Month(selectedMonth), 1, 0, 0, 0, 0, time.Now().Location())
		State.RunConfig.EndDate = State.RunConfig.StartDate.AddDate(0, 1, 0)

		runner, err := runners.New(State.RunConfig, State.Client)
		if err != nil {
			panic(err)
		}

		doAfter := func(rowValues [][]string) error {
			State.RunValues = rowValues
			State.RunCompleted = true
			State.RunStarted = false
			State.RunRequested = false
			return nil
		}
		runner.After(doAfter)
		go runner.Run(State.Context)
	}

	if State.RunRequested {
		return inset.Layout(gtx, material.Body2(th, "working...").Layout)
	}
	return layoutResultValues(gtx, State.RunValues)
	// return layout.Flex{
	// 	Alignment: layout.Middle,
	// 	Axis:      layout.Vertical,
	// }.Layout(
	// 	gtx,
	// 	layout.Rigid(func(gtx C) D {
	// 		if State.RunRequested {
	// 			return inset.Layout(gtx, material.Body2(th, "working...").Layout)
	// 		}
	// 		return layoutResultValues(gtx, State.RunValues)
	// 	}),
	// )
}

func layoutResultValues(gtx C, values [][]string) D {
	return layout.Flex{
		Alignment: layout.Start,
		Axis:      layout.Horizontal,
	}.Layout(gtx,
		resultCols(gtx, values)...,
	)
}
func resultCols(gtx C, values [][]string) []layout.FlexChild {
	childRows := make([]layout.FlexChild, 0, len(values))
	if len(values) == 0 {
		return childRows
	}
	for colIdx := 0; colIdx < len(values[0]); colIdx++ {
		idx := colIdx
		childRows = append(childRows, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Alignment: layout.Start,
				Axis:      layout.Vertical,
			}.Layout(gtx,
				resultCol(gtx, values, idx)...,
			)
		}))
	}
	return childRows
}

func resultCol(gtx C, values [][]string, colIdx int) []layout.FlexChild {
	items := make([]layout.FlexChild, 0, len(values))
	for rowIdx := 0; rowIdx < len(values); rowIdx++ {
		i := values[rowIdx][colIdx]
		items = append(items, layout.Rigid(func(gtx C) D {
			return inset.Layout(gtx, material.Body2(th, i).Layout)
		}))
	}
	return items
}
