package gui

import (
	"context"
	"image/color"
	"log"
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
	"github.com/3xcellent/github-metrics/config"
	"github.com/3xcellent/github-metrics/models"
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
	RunOptionsPage = iota
	ProjectsPage
	ResultsPage
	ConnectionSettingsPage
)

var ProjectsView = new(Projects)

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
	nav      = materials.NewNav("Gethub Metrics", "generate project metrics")
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
	outputText        string
	outputTextField   materials.TextField
	resultText        string
	resultTextField   materials.TextField

	appPages = []Page{
		{
			NavItem: materials.NavItem{
				Name: "Run Options",
				Icon: RunIcon,
				Tag:  RunOptionsPage,
			},
			layout: layoutRunOptionsPage,
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
				Name: "Results",
				Icon: ProjectsIcon,
				Tag:  ResultsPage,
			},
			layout: LayoutResults,
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

// Start - starts gui
func Start(ctx context.Context, cfg *config.AppConfig, args []string) error {
	w := app.NewWindow()
	// initialize state and set github client
	State = NewState(ctx)

	err := State.SetClient(cfg.API)
	if err != nil {
		return err
	}

	if len(args) > 0 {
		runConfig, err := cfg.GetRunConfig(args[0])
		if err != nil {
			return err
		}
		State.SelectedProjectID = runConfig.ProjectID
		project, err := State.Client.GetProject(ctx, State.SelectedProjectID)
		if err != nil {
			return err
		}
		hasLoadedProjects = true
		project.Owner = runConfig.Owner
		availableProjects = models.Projects{project}
		State.SelectedProjectName = project.Name
		State.RunConfig = runConfig
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

				paint.Fill(gtx.Ops, th.Palette.Bg)
				layoutApp(gtx, e)
				e.Frame(gtx.Ops)
			}
		}
	}
}

func layoutApp(gtx C, e system.FrameEvent) D {
	return layout.Inset{
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
}
