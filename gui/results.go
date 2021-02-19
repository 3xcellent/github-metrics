package gui

import (
	"context"
	"time"

	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/3xcellent/github-metrics/metrics/runners"
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
		go runner.Run(context.Background())
	}

	if State.RunRequested {
		return inset.Layout(gtx, material.Body2(th, "working...").Layout)
	}
	switch State.RunConfig.MetricName {
	case "issues":
		return layoutResultValues(gtx, State.RunValues)
	default:
		return defaultValues(State.RunValues).Layout(gtx)
	}
}

type defaultValues [][]string

func (vals defaultValues) Layout(gtx C) D {
	return layout.Flex{
		Alignment: layout.Start,
		Axis:      layout.Horizontal,
	}.Layout(gtx,
		vals.details(gtx)...,
	)
}
func (vals defaultValues) details(gtx C) []layout.FlexChild {
	numRows := len(vals)
	childRows := make([]layout.FlexChild, 0, numRows)
	if numRows == 0 {
		return childRows
	}
	numCols := len(vals[0])
	for colIdx := 0; colIdx < numCols; colIdx++ {
		idx := colIdx
		childRows = append(childRows, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Alignment: layout.Start,
				Axis:      layout.Vertical,
			}.Layout(gtx,
				vals.columnDetails(gtx, idx)...,
			)
		}))
	}
	return childRows
}

func (vals defaultValues) columnDetails(gtx C, colIdx int) []layout.FlexChild {
	numRows := len(vals)

	items := make([]layout.FlexChild, 0, numRows)
	for rowIdx := 0; rowIdx < numRows; rowIdx++ {
		i := vals[rowIdx][colIdx]
		items = append(items, layout.Rigid(func(gtx C) D {
			return inset.Layout(gtx, material.Body2(th, i).Layout)
		}))
	}
	return items
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
