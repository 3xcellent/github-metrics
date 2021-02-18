package gui

import (
	"fmt"

	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/3xcellent/github-metrics/models"
)

// Issue wraps models.Issue for the gui layout
type Issue struct {
	m      models.Issue
	events Events
}

// Event wraps models.IssueEvent for the gui layout
type Event struct {
	m models.IssueEvent
}

// Events allows the gui to layout a slicee of events
type Events []Event

// Layout returns the layout.Dimensions for an Issue
func (i *Issue) Layout(gtx C) D {
	return layout.Flex{
		Alignment: layout.Start,
		Axis:      layout.Horizontal,
	}.Layout(gtx,
		i.details(gtx),
		i.events.details(gtx),
	)
}

func (i *Issue) details(gtx C) layout.FlexChild {
	return layout.Rigid(func(gtx C) D {
		separator := layout.Rigid(func(gtx C) D {
			return inset.Layout(gtx, material.Body2(th, "/").Layout)
		})
		return layout.Flex{
			Alignment: layout.Start,
			Axis:      layout.Horizontal,
		}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return inset.Layout(gtx, material.Body2(th, i.m.Owner).Layout)
			}),
			separator,
			layout.Rigid(func(gtx C) D {
				return inset.Layout(gtx, material.Body2(th, i.m.RepoName).Layout)
			}),
			separator,
			layout.Rigid(func(gtx C) D {
				return inset.Layout(gtx, material.Body2(th, fmt.Sprintf("%d", i.m.Number)).Layout)
			}),
		)
	})
}

func (events Events) details(gtx C) layout.FlexChild {
	eventChildren := make([]layout.FlexChild, 0)
	for _, event := range events {
		e := event
		eventChildren = append(eventChildren, e.details(gtx))
	}
	return layout.Rigid(func(gtx C) D {
		return layout.Flex{
			Alignment: layout.Start,
			Axis:      layout.Vertical,
		}.Layout(gtx,
			eventChildren...,
		)
	})
}

// Layout returns Dimensions with Event details
func (e *Event) details(gtx C) layout.FlexChild {
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Alignment: layout.Start,
			Axis:      layout.Vertical,
		}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return inset.Layout(gtx, material.Body2(th, e.m.CreatedAt.String()).Layout)
			}),
			layout.Rigid(func(gtx C) D {
				return inset.Layout(gtx, material.Body2(th, e.m.Event).Layout)
			}),
		)
	})
}

// func layoutResultValues(gtx C, values [][]string) D {
// 	return layout.Flex{
// 		Alignment: layout.Start,
// 		Axis:      layout.Horizontal,
// 	}.Layout(gtx,
// 		resultCols(gtx, values)...,
// 	)
// }
// func resultCols(gtx C, values [][]string) []layout.FlexChild {
// 	childRows := make([]layout.FlexChild, 0, len(values))
// 	if len(values) == 0 {
// 		return childRows
// 	}
// 	for colIdx := 0; colIdx < len(values[0]); colIdx++ {
// 		idx := colIdx
// 		childRows = append(childRows, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
// 			return layout.Flex{
// 				Alignment: layout.Start,
// 				Axis:      layout.Vertical,
// 			}.Layout(gtx,
// 				resultCol(gtx, values, idx)...,
// 			)
// 		}))
// 	}
// 	return childRows
// }

// func resultCol(gtx C, values [][]string, colIdx int) []layout.FlexChild {
// 	items := make([]layout.FlexChild, 0, len(values))
// 	for rowIdx := 0; rowIdx < len(values); rowIdx++ {
// 		i := values[rowIdx][colIdx]
// 		items = append(items, layout.Rigid(func(gtx C) D {
// 			return inset.Layout(gtx, material.Body2(th, i).Layout)
// 		}))
// 	}
// 	return items
// }
