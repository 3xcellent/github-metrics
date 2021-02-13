package gui

import (
	"fmt"
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
				logrus.Debugf("doAfter rowValues: %#v", rowValues)
				for _, row := range rowValues {
					resultText = fmt.Sprintf("%s\n%s", resultText, strings.Join(row, ","))
				}

				State.RunCompleted = true
				State.RunRequested = false
				State.RunStarted = false

				outputText = resultText
				return nil
			}
			colsRunner.After(doAfter)
			go colsRunner.Run(State.Context)
		}
	}

	// if State.RunCompleted {
	// 	switch selectedCommand {
	// 	default:
	// 		return layoutResults(gtx)
	// 	}
	// }

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
			return inset.Layout(gtx, material.Body2(th, resultText).Layout)
		}),
	)
}
