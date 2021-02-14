package gui

import (
	"fmt"
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
		debug(fmt.Sprintf("Starting: %s - %d / %d\n", selectedCommand, selectedMonth, selectedYear))
		switch selectedCommand {
		case "columns":
			debug(fmt.Sprintf("\tcolumns: %s - %d / %d\n", selectedCommand, selectedMonth, selectedYear))
			selectedProject, err := availableProjects.GetProject(State.SelectedProjectID)
			if err != nil {
				panic(err)
			}

			State.RunConfig.ProjectID = selectedProject.ID
			State.RunConfig.Owner = selectedProject.Owner
			State.RunConfig.StartDate = time.Date(selectedYear, time.Month(selectedMonth), 1, 0, 0, 0, 0, time.Now().Location())
			State.RunConfig.EndDate = State.RunConfig.StartDate.AddDate(0, 1, 0)

			colsRunner := runners.NewColumnsRunner(State.RunConfig, State.Client)
			logrus.Debugf("colsRunner: %#v", colsRunner)

			doAfter := func(rowValues [][]string) error {
				State.RunValues = rowValues
				State.RunCompleted = true
				State.RunStarted = false
				State.RunRequested = false
				return nil
			}
			colsRunner.After(doAfter)
			go colsRunner.Run(State.Context)
		}
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

func layoutResultValues(gtx C, rows [][]string) D {
	return layout.Flex{
		Alignment: layout.Start,
		Axis:      layout.Vertical,
	}.Layout(gtx,
		resultRows(gtx, rows)...,
	)
}

func resultRows(gtx C, rows [][]string) []layout.FlexChild {
	childRows := make([]layout.FlexChild, 0, len(rows))
	for _, row := range rows {
		r := row
		childRows = append(childRows, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Alignment: layout.Start,
				Axis:      layout.Horizontal,
			}.Layout(gtx,
				resultRow(gtx, r)...,
			)
		}))
	}
	return childRows
}

func resultRow(gtx C, row []string) []layout.FlexChild {
	items := make([]layout.FlexChild, 0, len(row))
	for _, item := range row {
		i := item
		items = append(items, layout.Rigid(func(gtx C) D {
			return inset.Layout(gtx, material.Body2(th, i).Layout)
		}))
	}
	return items
}
