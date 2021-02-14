package gui

import (
	"strings"
	"time"

	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/3xcellent/github-metrics/metrics/runners"
	"github.com/sirupsen/logrus"
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

	return layout.Flex{
		Alignment: layout.Middle,
		Axis:      layout.Vertical,
	}.Layout(
		gtx,
		layout.Rigid(func(gtx C) D {
			return inset.Layout(gtx, material.H1(th, "Results").Layout)
		}),
		layout.Rigid(func(gtx C) D {
			if State.RunRequested {
				return inset.Layout(gtx, material.Body2(th, "working...").Layout)
			}
			return layoutResultValues(gtx, State.RunValues)
		}),
	)
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
	logrus.Debugf("getting values for colIdx: %d", colIdx)
	items := make([]layout.FlexChild, 0, len(values))
	for rowIdx := 0; rowIdx < len(values); rowIdx++ {
		rowVals := values[rowIdx]
		logrus.Debugf("rowVals: %s", strings.Join(rowVals, ","))
		i := values[rowIdx][colIdx]
		items = append(items, layout.Rigid(func(gtx C) D {
			return inset.Layout(gtx, material.Body2(th, i).Layout)
		}))
	}
	return items
}

// func layoutResultValues(gtx C, rows [][]string) D {
// 	return layout.Flex{
// 		Alignment: layout.Start,
// 		Axis:      layout.Vertical,
// 	}.Layout(gtx,
// 		resultRows(gtx, rows)...,
// 	)
// }

// func resultRows(gtx C, rows [][]string) []layout.FlexChild {
// 	childRows := make([]layout.FlexChild, 0, len(rows))
// 	for _, row := range rows {
// 		r := row
// 		childRows = append(childRows, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
// 			return layout.Flex{
// 				Alignment: layout.Start,
// 				Axis:      layout.Horizontal,
// 			}.Layout(gtx,
// 				resultRow(gtx, r)...,
// 			)
// 		}))
// 	}
// 	return childRows
// }

// func resultRow(gtx C, row []string) []layout.FlexChild {
// 	items := make([]layout.FlexChild, 0, len(row))
// 	for _, item := range row {
// 		i := item
// 		items = append(items, layout.Rigid(func(gtx C) D {
// 			return inset.Layout(gtx, material.Body2(th, i).Layout)
// 		}))
// 	}
// 	return items
// }
