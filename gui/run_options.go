package gui

import (
	"fmt"
	"strconv"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/3xcellent/github-metrics/config"
	"github.com/3xcellent/github-metrics/metrics"
	"github.com/sirupsen/logrus"
)

func layoutRunOptionsPage(gtx C) D {
	if !State.HasValidatedConnection {
		nav.SetNavDestination(ConnectionSettingsPage)
	}

	if State.SelectedProjectID == 0 {
		nav.SetNavDestination(ProjectsPage)
	}

	if runButton.Clicked() {
		State.RunRequested = true
		logrus.Debugf("runButton.Clicked() - State.RunRequested:%t", State.RunRequested)
		nav.SetNavDestination(ResultsPage)
		op.InvalidateOp{}.Add(gtx.Ops)
	}

	return layoutRunOptions(gtx)
}

func layoutRunOptions(gtx C) D {
	return layout.Flex{
		Alignment: layout.Start,
		Axis:      layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return inset.Layout(gtx, material.Body1(th, fmt.Sprintf("Project : %s", State.SelectedProjectName)).Layout)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Alignment: layout.Start,
				Axis:      layout.Horizontal,
			}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					if yearsEnum.Changed() {
						intVal, err := strconv.Atoi(yearsEnum.Value)
						if err != nil {
							panic(err)
						}
						selectedYear = intVal

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
									return material.Body2(th, "Year:").Layout(gtx)
								}),
								layout.Rigid(func(gtx C) D {
									return layout.Flex{
										Axis: layout.Vertical,
									}.Layout(
										gtx,
										yearOptions(th, &yearsEnum, time.Now().Year())...,
									)
								}),
							)
						},
					)
				}),
				layout.Rigid(func(gtx C) D {
					if monthsEnum.Changed() {
						intVal, err := strconv.Atoi(monthsEnum.Value)
						if err != nil {
							panic(err)
						}
						selectedMonth = intVal

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
									return material.Body2(th, "Month:").Layout(gtx)
								}),
								layout.Rigid(func(gtx C) D {
									return layout.Flex{
										Axis: layout.Vertical,
									}.Layout(
										gtx,
										monthOptions(th, &monthsEnum)...,
									)
								}),
							)
						},
					)
				}),
				layout.Rigid(func(gtx C) D {
					if commandsEnum.Changed() {
						State.RunConfig.MetricName = commandsEnum.Value
						logrus.Debugf("Metric selected: %s", State.RunConfig.MetricName)
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
									return material.Body2(th, "Metric:").Layout(gtx)
								}),
								layout.Rigid(func(gtx C) D {
									return layout.Flex{
										Axis: layout.Vertical,
									}.Layout(
										gtx,
										metricOptions(th, &commandsEnum)...,
									)
								}),
							)
						},
					)
				}),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {

			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(
				gtx,
				layout.Rigid(func(gtx C) D {
					return material.Button(th, &runButton, "Run").Layout(gtx)
				}),
			)
		}),
	)
}

func metricOptions(th *material.Theme, input *widget.Enum) []layout.FlexChild {
	options := make([]layout.FlexChild, 0, len(metrics.AvailableMetrics))

	for _, metric := range metrics.AvailableMetrics {
		m := metric
		options = append(options,
			layout.Rigid(func(gtx C) D {
				return material.RadioButton(
					th,
					input,
					m.Name,
					m.Name,
				).Layout(gtx)
			}),
			// layout.Rigid(func(gtx C) D {
			// 	return material.Body2(th, m.Description).Layout(gtx)
			// }),
		)
	}
	return options
}

func monthOptions(th *material.Theme, input *widget.Enum) []layout.FlexChild {
	options := make([]layout.FlexChild, 0, numYearsPrevious)

	for i := 0; i < 12; i++ {
		mon := time.Month(i + 1)
		options = append(options,
			layout.Rigid(func(gtx C) D {
				return material.RadioButton(
					th,
					input,
					fmt.Sprintf("%d", mon),
					fmt.Sprintf("%02d - %s", mon, mon.String()),
				).Layout(gtx)
			}))
	}
	return options
}

func yearOptions(th *material.Theme, input *widget.Enum, thisYear int) []layout.FlexChild {
	options := make([]layout.FlexChild, 0, numYearsPrevious)

	for i := 0; i < numYearsPrevious; i++ {
		year := thisYear - i
		options = append(options,
			layout.Rigid(func(gtx C) D {
				return material.RadioButton(
					th,
					input,
					fmt.Sprintf("%d", year),
					fmt.Sprintf("%d", year),
				).Layout(gtx)
			}))
	}
	return options
}

func projectBoardOptions(th *material.Theme, input *widget.Enum, boards []config.RunConfig) []layout.FlexChild {
	options := make([]layout.FlexChild, 0, len(boards))

	for _, b := range boards {
		board := b
		options = append(options,
			layout.Rigid(func(gtx C) D {
				return material.RadioButton(
					th,
					input,
					board.Name,
					board.Name,
				).Layout(gtx)
			}))
	}
	return options
}
