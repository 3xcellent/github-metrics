package gui

import (
	"context"
	"fmt"
	"image/color"
	"log"
	"strconv"
	"strings"
	"time"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"git.sr.ht/~whereswaldon/materials"
	"github.com/3xcellent/github-metrics/client"
	"github.com/3xcellent/github-metrics/config"
	"github.com/3xcellent/github-metrics/metrics/runners"
	"github.com/sirupsen/logrus"
)

type (
	// C is the layout.Context
	C = layout.Context

	// D is the layout.Dimensions
	D = layout.Dimensions
)

var barOnBottom bool

const (
	settingNameColumnWidth    = .3
	settingDetailsColumnWidth = 1 - settingNameColumnWidth
	numYearsPrevious          = 10
)

// page options
const (
	MainPage = iota
	ProjectsPage
	ConnectionSettingsPage
)

// Page - an instance of an available page
type Page struct {
	layout func(layout.Context) layout.Dimensions
	materials.NavItem
	Actions  []materials.AppBarAction
	Overflow []materials.OverflowAction

	// laying each page out within a layout.List enables scrolling for the page
	// content.
	layout.List
}

var (
	// Config - instance of the config.Config with loaded environment and configuration values
	Config *config.AppConfig

	// State - instance of state for for gui
	State *MetricsState

	// initialize channel to send clipboard content requests on
	clipboardRequests = make(chan string, 1)

	// initialize modal layer to draw modal components
	modal   = materials.NewModal()
	navAnim = materials.VisibilityAnimation{
		Duration: time.Millisecond * 100,
		State:    materials.Invisible,
	}
	nav      = materials.NewNav("Navigation Drawer", "This is an example.")
	modalNav = materials.ModalNavFrom(&nav, modal)

	bar = materials.NewAppBar(modal)

	inset              = layout.UniformInset(unit.Dp(8))
	th                 = material.NewTheme(gofont.Collection())
	lightPalette       = th.Palette
	lightPaletteAccent = func() material.Palette {
		out := th.Palette
		out.ContrastBg = color.NRGBA{A: 0xff, R: 0x9e, G: 0x9d, B: 0x24}
		return out
	}()
	altPalette = func() material.Palette {
		out := th.Palette
		out.Bg = color.NRGBA{R: 0xff, G: 0xfb, B: 0xe6, A: 0xff}
		out.Fg = color.NRGBA{A: 0xff}
		out.ContrastBg = color.NRGBA{R: 0x35, G: 0x69, B: 0x59, A: 0xff}
		return out
	}()
	altPaletteAccent = func() material.Palette {
		out := th.Palette
		out.Bg = color.NRGBA{R: 0xff, G: 0xfb, B: 0xe6, A: 0xff}
		out.Fg = color.NRGBA{A: 0xff}
		out.ContrastBg = color.NRGBA{R: 0xfd, G: 0x55, B: 0x23, A: 0xff}
		out.ContrastFg = out.Fg
		return out
	}()
	currentAccent material.Palette = lightPaletteAccent

	// heartBtn, plusBtn, exampleOverflowState               widget.Clickable
	// red, green, blue                                      widget.Clickable
	// contextBtn                                            widget.Clickable
	// eliasCopyButton, chrisCopyButtonGH, chrisCopyButtonLP widget.Clickable
	// bottomBar                                             widget.Bool
	// customNavIcon                                         widget.Bool
	// alternatePalette                                      widget.Bool
	// favorited                                             bool
	// inputAlignmentEnum                                    widget.Enum

	// nameInput    materials.TextField
	// addressInput materials.TextField
	// priceInput   materials.TextField
	// tweetInput   materials.TextField
	// numberInput  materials.TextField

	// Need to verify all params above
	nonModalDrawer widget.Bool
	inputAlignment = layout.Start

	projectsEnum      widget.Enum
	yearsEnum         widget.Enum
	monthsEnum        widget.Enum
	commandsEnum      widget.Enum
	runButton         widget.Clickable
	selectedBoardName string
	selectedYear      int
	selectedMonth     int
	selectedCommand   string
	outputText        string
	outputTextField   materials.TextField
	resultText        string
	resultTextField   materials.TextField

	appPages = []Page{
		{
			NavItem: materials.NavItem{
				Name: "Main",
				Icon: HomeIcon,
				Tag:  MainPage,
			},
			layout: layoutMainPage,
		},
		{
			NavItem: materials.NavItem{
				Name: "Projects",
				Icon: ProjectsIcon,
				Tag:  ProjectsPage,
			},
			layout: LayoutProjectsPage,
		},
		{
			NavItem: materials.NavItem{
				Name: "Github Connection Settings",
				Icon: SettingsIcon,
				Tag:  ConnectionSettingsPage,
			},
			layout: LayoutConnectionSettings,
		},
	}
)

func debug(args ...interface{}) {
	for _, arg := range args {
		value, ok := arg.(string)
		if !ok {
			value = *arg.(*string)
		}
		outputText = fmt.Sprintf("%s\n%s", outputText, value)
	}
	outputTextField.SetText(outputText)
}

// Start - starts gui
func Start(ctx context.Context, apiConfig config.APIConfig) error {
	w := app.NewWindow()
	// initialize state and github client
	State = NewState(ctx)
	err := State.SetClient(apiConfig)
	if err != nil {
		panic(err)
	}

	// initialize form fields
	logrus.Info("initializing form fields")
	tokenInput.SetText(State.APIConfig.Token)
	ownerInput.SetText(State.APIConfig.Owner)
	baseURLInput.SetText(State.APIConfig.BaseURL)
	uploadURLInput.SetText(State.APIConfig.UploadURL)

	var ops op.Ops

	bar.NavigationIcon = MenuIcon

	// assign navigation tags and configure navigation bar with all appPages
	for i := range appPages {
		page := &appPages[i]
		page.List.Axis = layout.Vertical
		page.NavItem.Tag = i
		nav.AddNavItem(page.NavItem)
	}

	// configure app bar initial state
	page := appPages[nav.CurrentNavDestination().(int)]
	bar.Title = page.Name
	bar.SetActions(page.Actions, page.Overflow)

	for {
		select {
		case content := <-clipboardRequests:
			w.WriteClipboard(content)
		case e := <-w.Events():
			switch e := e.(type) {
			case system.DestroyEvent:
				return e.Err
			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)
				if nav.NavDestinationChanged() {
					page := appPages[nav.CurrentNavDestination().(int)]
					bar.Title = page.Name
					bar.SetActions(page.Actions, page.Overflow)

				}
				for _, event := range bar.Events(gtx) {
					switch event := event.(type) {
					case materials.AppBarNavigationClicked:
						if nonModalDrawer.Value {
							navAnim.ToggleVisibility(gtx.Now)
						} else {
							modalNav.Appear(gtx.Now)
							navAnim.Disappear(gtx.Now)
						}
					case materials.AppBarContextMenuDismissed:
						log.Printf("Context menu dismissed: %v", event)
					case materials.AppBarOverflowActionClicked:
						log.Printf("Overflow action selected: %v", event)
					}
				}
				// if alternatePalette.Value {
				// 	th.Palette = altPalette
				// 	currentAccent = altPaletteAccent
				// } else {
				// 	th.Palette = lightPalette
				// 	currentAccent = lightPaletteAccent
				// }

				if State.RunRequested && !State.RunStarted {
					State.RunStarted = true
					debug(fmt.Sprintf("Starting: %s - %d / %d\n", selectedCommand, selectedMonth, selectedYear))
					switch selectedCommand {
					case "columns":
						debug(fmt.Sprintf("columns: %s - %d / %d\n", selectedCommand, selectedMonth, selectedYear))
						selectedProject, err := availableProjects.GetProject(State.SelectedProjectID)
						if err != nil {
							panic(err)
						}

						runCfg := selectedProject.RunConfig()
						runCfg.StartDate = time.Date(selectedYear, time.Month(selectedMonth), 1, 0, 0, 0, 0, time.Now().Location())
						runCfg.EndDate = runCfg.StartDate.AddDate(0, 1, 0)

						logrus.Debugf("runCfg: %#v", runCfg)

						ghClient, err := client.New(ctx, State.APIConfig)
						if err != nil {
							panic(err)
						}

						colsRunner := runners.NewColumnsRunner(runCfg, ghClient)
						logrus.Debugf("colsRunner: %#v", colsRunner)

						doAfter := func(rowValues [][]string) error {
							logrus.Debugf("doAfter rowValues: %#v", rowValues)
							for _, row := range rowValues {
								resultText = fmt.Sprintf("%s\n%s", resultText, strings.Join(row, ","))
							}

							State.RunCompleted = true
							State.RunRequested = false
							State.RunStarted = false

							resultTextField.SetText(resultText)
							logrus.Debugf("resultTextField: %s", resultText)

							return nil
						}
						colsRunner.After(doAfter)
						go colsRunner.Run(ctx)
					}
				}

				// TODO try to move the below section outside of events select/case, but keep in for loop
				paint.Fill(gtx.Ops, th.Palette.Bg)

				layout.Inset{
					Top:    e.Insets.Top,
					Bottom: e.Insets.Bottom,
					Left:   e.Insets.Left,
					Right:  e.Insets.Right,
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					content := layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Max.X /= 3
								return nav.Layout(gtx, th, &navAnim)
							}),
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								page := &appPages[nav.CurrentNavDestination().(int)]
								return page.List.Layout(gtx, 1, func(gtx C, _ int) D {
									return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
										return page.layout(gtx)
									})
								})
							}),
						)
					})
					bar := layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return bar.Layout(gtx, th)
					})
					flex := layout.Flex{Axis: layout.Vertical}
					flex.Layout(gtx, bar, content)
					modal.Layout(gtx, th)
					return layout.Dimensions{Size: gtx.Constraints.Max}
				})

				e.Frame(gtx.Ops)
			}
		}
	}
}

func layoutMainPage(gtx C) D {
	if !State.HasValidatedConnection {
		nav.SetNavDestination(ConnectionSettingsPage)
	}

	if State.SelectedProjectID == 0 {
		nav.SetNavDestination(ProjectsPage)
	}

	if runButton.Clicked() {
		State.RunRequested = true
		debug(fmt.Sprintf("runButton.Clicked() - State.RunRequested:%t", State.RunRequested))
		// State.Config.StartDate = time.Date(selectedYear, time.Month(selectedMonth), 1, 0, 0, 0, 0, time.Now().Location())
		op.InvalidateOp{}.Add(gtx.Ops)
	}
	if State.RunCompleted {
		switch selectedCommand {
		default:
			return layoutResults(gtx)
		}
	}
	if State.RunRequested {
		switch selectedCommand {
		default:
			return layoutProgress(gtx)
		}
	}

	return layoutMainRunOptions(gtx)
}

func layoutProgress(gtx C) D {
	return layout.Flex{
		Alignment: layout.Start,
		Axis:      layout.Horizontal,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			if !State.RunCompleted {
				outputTextField.SetText(outputText)
			}
			return outputTextField.Layout(gtx, th, "running")
		}),
	)
}

func layoutResults(gtx C) D {
	return layout.Flex{
		Alignment: layout.Start,
		Axis:      layout.Horizontal,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return resultTextField.Layout(gtx, th, "CSV Output")
		}),
	)
}

func layoutMainRunOptions(gtx C) D {
	return layout.Flex{
		Alignment: layout.Start,
		Axis:      layout.Horizontal,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return inset.Layout(gtx, material.Body1(th, fmt.Sprintf("Project : %s", State.SelectedProjectName)).Layout)
		}),
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
				selectedCommand = commandsEnum.Value
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
								layout.Rigid(func(gtx C) D {
									return material.RadioButton(
										th,
										&commandsEnum,
										"issues",
										"Issues",
									).Layout(gtx)
								}),
								layout.Rigid(func(gtx C) D {
									return material.RadioButton(
										th,
										&commandsEnum,
										"columns",
										"Columns",
									).Layout(gtx)
								}),
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
					return material.Button(th, &runButton, "Run").Layout(gtx)
				}),
			)
		}),
	)
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

// func LayoutAppBarPage(gtx C) D {
// 	return layout.Flex{
// 		Alignment: layout.Middle,
// 		Axis:      layout.Vertical,
// 	}.Layout(gtx,
// 		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
// 			return inset.Layout(gtx, material.Body1(th, `The app bar widget provides a consistent interface element for triggering navigation and page-specific actions.

// The controls below allow you to see the various features available in our App Bar implementation.`).Layout)
// 		}),
// 		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
// 			return layout.Flex{Alignment: layout.Baseline}.Layout(gtx,
// 				layout.Flexed(settingNameColumnWidth, func(gtx C) D {
// 					return inset.Layout(gtx, material.Body1(th, "Contextual App Bar").Layout)
// 				}),
// 				layout.Flexed(settingDetailsColumnWidth, func(gtx C) D {
// 					if contextBtn.Clicked() {
// 						bar.SetContextualActions(
// 							[]materials.AppBarAction{
// 								materials.SimpleIconAction(th, &red, HeartIcon,
// 									materials.OverflowAction{
// 										Name: "House",
// 										Tag:  &red,
// 									},
// 								),
// 							},
// 							[]materials.OverflowAction{
// 								{
// 									Name: "foo",
// 									Tag:  &blue,
// 								},
// 								{
// 									Name: "bar",
// 									Tag:  &green,
// 								},
// 							},
// 						)
// 						bar.ToggleContextual(gtx.Now, "Contextual Title")
// 					}
// 					return material.Button(th, &contextBtn, "Trigger").Layout(gtx)
// 				}),
// 			)
// 		}),
// 		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
// 			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
// 				layout.Flexed(settingNameColumnWidth, func(gtx C) D {
// 					return inset.Layout(gtx, material.Body1(th, "Bottom App Bar").Layout)
// 				}),
// 				layout.Flexed(settingDetailsColumnWidth, func(gtx C) D {
// 					if bottomBar.Changed() {
// 						if bottomBar.Value {
// 							nav.Anchor = materials.Bottom
// 						} else {
// 							nav.Anchor = materials.Top
// 						}
// 					}

// 					return inset.Layout(gtx, material.Switch(th, &bottomBar).Layout)
// 				}),
// 			)
// 		}),
// 		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
// 			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
// 				layout.Flexed(settingNameColumnWidth, func(gtx C) D {
// 					return inset.Layout(gtx, material.Body1(th, "Custom Navigation Icon").Layout)
// 				}),
// 				layout.Flexed(settingDetailsColumnWidth, func(gtx C) D {
// 					if customNavIcon.Changed() {
// 						if customNavIcon.Value {
// 							bar.NavigationIcon = HomeIcon
// 						} else {
// 							bar.NavigationIcon = MenuIcon
// 						}
// 					}
// 					return inset.Layout(gtx, material.Switch(th, &customNavIcon).Layout)
// 				}),
// 			)
// 		}),
// 		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
// 			return layout.Flex{Alignment: layout.Baseline}.Layout(gtx,
// 				layout.Flexed(settingNameColumnWidth, func(gtx C) D {
// 					return inset.Layout(gtx, material.Body1(th, "Animated Resize").Layout)
// 				}),
// 				layout.Flexed(settingDetailsColumnWidth, func(gtx C) D {
// 					return inset.Layout(gtx, material.Body2(th, "Resize the width of your screen to see app bar actions collapse into or emerge from the overflow menu (as size permits).").Layout)
// 				}),
// 			)
// 		}),
// 		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
// 			return layout.Flex{Alignment: layout.Baseline}.Layout(gtx,
// 				layout.Flexed(settingNameColumnWidth, func(gtx C) D {
// 					return inset.Layout(gtx, material.Body1(th, "Custom Action Buttons").Layout)
// 				}),
// 				layout.Flexed(settingDetailsColumnWidth, func(gtx C) D {
// 					if heartBtn.Clicked() {
// 						favorited = !favorited
// 					}
// 					return inset.Layout(gtx, material.Body2(th, "Click the heart action to see custom button behavior.").Layout)
// 				}),
// 			)
// 		}),
// 	)
// }

// func LayoutNavDrawerPage(gtx C) D {
// 	return layout.Flex{
// 		Alignment: layout.Middle,
// 		Axis:      layout.Vertical,
// 	}.Layout(gtx,
// 		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
// 			return inset.Layout(gtx, material.Body1(th, `The nav drawer widget provides a consistent interface element for navigation.

// The controls below allow you to see the various features available in our Navigation Drawer implementation.`).Layout)
// 		}),
// 		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
// 			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
// 				layout.Flexed(settingNameColumnWidth, func(gtx C) D {
// 					return inset.Layout(gtx, material.Body1(th, "Use non-modal drawer").Layout)
// 				}),
// 				layout.Flexed(settingDetailsColumnWidth, func(gtx C) D {
// 					if nonModalDrawer.Changed() {
// 						if nonModalDrawer.Value {
// 							navAnim.Appear(gtx.Now)
// 						} else {
// 							navAnim.Disappear(gtx.Now)
// 						}
// 					}
// 					return inset.Layout(gtx, material.Switch(th, &nonModalDrawer).Layout)
// 				}),
// 			)
// 		}),
// 		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
// 			return layout.Flex{Alignment: layout.Baseline}.Layout(gtx,
// 				layout.Flexed(settingNameColumnWidth, func(gtx C) D {
// 					return inset.Layout(gtx, material.Body1(th, "Drag to Close").Layout)
// 				}),
// 				layout.Flexed(settingDetailsColumnWidth, func(gtx C) D {
// 					return inset.Layout(gtx, material.Body2(th, "You can close the modal nav drawer by dragging it to the left.").Layout)
// 				}),
// 			)
// 		}),
// 		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
// 			return layout.Flex{Alignment: layout.Baseline}.Layout(gtx,
// 				layout.Flexed(settingNameColumnWidth, func(gtx C) D {
// 					return inset.Layout(gtx, material.Body1(th, "Touch Scrim to Close").Layout)
// 				}),
// 				layout.Flexed(settingDetailsColumnWidth, func(gtx C) D {
// 					return inset.Layout(gtx, material.Body2(th, "You can close the modal nav drawer touching anywhere in the translucent scrim to the right.").Layout)
// 				}),
// 			)
// 		}),
// 		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
// 			return layout.Flex{Alignment: layout.Baseline}.Layout(gtx,
// 				layout.Flexed(settingNameColumnWidth, func(gtx C) D {
// 					return inset.Layout(gtx, material.Body1(th, "Bottom content anchoring").Layout)
// 				}),
// 				layout.Flexed(settingDetailsColumnWidth, func(gtx C) D {
// 					return inset.Layout(gtx, material.Body2(th, "If you toggle support for the bottom app bar in the App Bar settings, nav drawer content will anchor to the bottom of the drawer area instead of the top.").Layout)
// 				}),
// 			)
// 		}),
// 	)
// }

// func LayoutAboutPage(gtx C) D {
// 	th := *th
// 	th.Palette = currentAccent
// 	return layout.Flex{
// 		Alignment: layout.Middle,
// 		Axis:      layout.Vertical,
// 	}.Layout(gtx,
// 		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
// 			return inset.Layout(gtx, material.Body1(&th, `This library implements material design components from https://material.io using https://gioui.org.

// Materials (this library) would not be possible without the incredible work of Elias Naur and the Gio community. Materials is maintained by Chris Waldon.

// If you like this library and work like it, please consider sponsoring Elias and/or Chris!`).Layout)
// 		}),
// 		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
// 			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
// 				layout.Flexed(settingDetailsColumnWidth, func(gtx C) D {
// 					return inset.Layout(gtx, material.Body1(&th, "Try another theme:").Layout)
// 				}),
// 				layout.Flexed(settingNameColumnWidth, func(gtx C) D {
// 					return inset.Layout(gtx, material.Switch(&th, &alternatePalette).Layout)
// 				}),
// 			)
// 		}),
// 		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
// 			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
// 				layout.Flexed(settingDetailsColumnWidth, func(gtx C) D {
// 					return inset.Layout(gtx, material.Body1(&th, "Elias Naur can be sponsored on GitHub at "+sponsorEliasURL).Layout)
// 				}),
// 				layout.Flexed(settingNameColumnWidth, func(gtx C) D {
// 					if eliasCopyButton.Clicked() {
// 						clipboardRequests <- sponsorEliasURL
// 					}
// 					return inset.Layout(gtx, material.Button(&th, &eliasCopyButton, "Copy Sponsorship URL").Layout)
// 				}),
// 			)
// 		}),
// 		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
// 			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
// 				layout.Flexed(settingDetailsColumnWidth, func(gtx C) D {
// 					return inset.Layout(gtx, material.Body1(&th, "Chris Waldon can be sponsored on GitHub at "+sponsorChrisURLGitHub+" and on Liberapay at "+sponsorChrisURLLiberapay).Layout)
// 				}),
// 				layout.Flexed(settingNameColumnWidth, func(gtx C) D {
// 					if chrisCopyButtonGH.Clicked() {
// 						clipboardRequests <- sponsorChrisURLGitHub
// 					}
// 					if chrisCopyButtonLP.Clicked() {
// 						clipboardRequests <- sponsorChrisURLLiberapay
// 					}
// 					return inset.Layout(gtx, func(gtx C) D {
// 						return layout.Flex{}.Layout(gtx,
// 							layout.Flexed(.5, material.Button(&th, &chrisCopyButtonGH, "Copy GitHub URL").Layout),
// 							layout.Flexed(.5, material.Button(&th, &chrisCopyButtonLP, "Copy Liberapay URL").Layout),
// 						)
// 					})
// 				}),
// 			)
// 		}),
// 	)
// }

// func LayoutTextFieldPage(gtx C) D {
// 	return layout.Flex{
// 		Axis: layout.Vertical,
// 	}.Layout(
// 		gtx,
// 		layout.Rigid(func(gtx C) D {
// 			nameInput.Alignment = inputAlignment
// 			return nameInput.Layout(gtx, th, "Name")
// 		}),
// 		layout.Rigid(func(gtx C) D {
// 			return inset.Layout(gtx, material.Body2(th, "Responds to hover events.").Layout)
// 		}),
// 		layout.Rigid(func(gtx C) D {
// 			addressInput.Alignment = inputAlignment
// 			return addressInput.Layout(gtx, th, "Address")
// 		}),
// 		layout.Rigid(func(gtx C) D {
// 			return inset.Layout(gtx, material.Body2(th, "Label animates properly when you click to select the text field.").Layout)
// 		}),
// 		layout.Rigid(func(gtx C) D {
// 			priceInput.Prefix = func(gtx C) D {
// 				th := *th
// 				th.Palette.Fg = color.NRGBA{R: 100, G: 100, B: 100, A: 255}
// 				return material.Label(&th, th.TextSize, "$").Layout(gtx)
// 			}
// 			priceInput.Suffix = func(gtx C) D {
// 				th := *th
// 				th.Palette.Fg = color.NRGBA{R: 100, G: 100, B: 100, A: 255}
// 				return material.Label(&th, th.TextSize, ".00").Layout(gtx)
// 			}
// 			priceInput.SingleLine = true
// 			priceInput.Alignment = inputAlignment
// 			return priceInput.Layout(gtx, th, "Price")
// 		}),
// 		layout.Rigid(func(gtx C) D {
// 			return inset.Layout(gtx, material.Body2(th, "Can have prefix and suffix elements.").Layout)
// 		}),
// 		layout.Rigid(func(gtx C) D {
// 			if err := func() string {
// 				for _, r := range numberInput.Text() {
// 					if !unicode.IsDigit(r) {
// 						return "Must contain only digits"
// 					}
// 				}
// 				return ""
// 			}(); err != "" {
// 				numberInput.SetError(err)
// 			} else {
// 				numberInput.ClearError()
// 			}
// 			numberInput.SingleLine = true
// 			numberInput.Alignment = inputAlignment
// 			return numberInput.Layout(gtx, th, "Number")
// 		}),
// 		layout.Rigid(func(gtx C) D {
// 			return inset.Layout(gtx, material.Body2(th, "Can be validated.").Layout)
// 		}),
// 		layout.Rigid(func(gtx C) D {
// 			if tweetInput.TextTooLong() {
// 				tweetInput.SetError("Too many characters")
// 			} else {
// 				tweetInput.ClearError()
// 			}
// 			tweetInput.CharLimit = 128
// 			tweetInput.Helper = "Tweets have a limited character count"
// 			tweetInput.Alignment = inputAlignment
// 			return tweetInput.Layout(gtx, th, "Tweet")
// 		}),
// 		layout.Rigid(func(gtx C) D {
// 			return inset.Layout(gtx, material.Body2(th, "Can have a character counter and help text.").Layout)
// 		}),
// 		layout.Rigid(func(gtx C) D {
// 			if inputAlignmentEnum.Changed() {
// 				switch inputAlignmentEnum.Value {
// 				case layout.Start.String():
// 					inputAlignment = layout.Start
// 				case layout.Middle.String():
// 					inputAlignment = layout.Middle
// 				case layout.End.String():
// 					inputAlignment = layout.End
// 				default:
// 					inputAlignment = layout.Start
// 				}
// 				op.InvalidateOp{}.Add(gtx.Ops)
// 			}
// 			return inset.Layout(
// 				gtx,
// 				func(gtx C) D {
// 					return layout.Flex{
// 						Axis: layout.Vertical,
// 					}.Layout(
// 						gtx,
// 						layout.Rigid(func(gtx C) D {
// 							return material.Body2(th, "Text Alignment").Layout(gtx)
// 						}),
// 						layout.Rigid(func(gtx C) D {
// 							return layout.Flex{
// 								Axis: layout.Vertical,
// 							}.Layout(
// 								gtx,
// 								layout.Rigid(func(gtx C) D {
// 									return material.RadioButton(
// 										th,
// 										&inputAlignmentEnum,
// 										layout.Start.String(),
// 										"Start",
// 									).Layout(gtx)
// 								}),
// 								layout.Rigid(func(gtx C) D {
// 									return material.RadioButton(
// 										th,
// 										&inputAlignmentEnum,
// 										layout.Middle.String(),
// 										"Middle",
// 									).Layout(gtx)
// 								}),
// 								layout.Rigid(func(gtx C) D {
// 									return material.RadioButton(
// 										th,
// 										&inputAlignmentEnum,
// 										layout.End.String(),
// 										"End",
// 									).Layout(gtx)
// 								}),
// 							)
// 						}),
// 					)
// 				},
// 			)
// 		}),
// 		layout.Rigid(func(gtx C) D {
// 			return inset.Layout(gtx, material.Body2(th, "This text field implementation was contributed by Jack Mordaunt. Thanks Jack!").Layout)
// 		}),
// 	)
// }
