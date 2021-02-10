package gui

import (
	"gioui.org/widget"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

// icons
var (
	MenuIcon *widget.Icon = func() *widget.Icon {
		icon, _ := widget.NewIcon(icons.NavigationMenu)
		return icon
	}()

	HomeIcon *widget.Icon = func() *widget.Icon {
		icon, _ := widget.NewIcon(icons.ActionHome)
		return icon
	}()

	ProjectsIcon *widget.Icon = func() *widget.Icon {
		icon, _ := widget.NewIcon(icons.ActionDashboard)
		return icon
	}()

	SettingsIcon *widget.Icon = func() *widget.Icon {
		icon, _ := widget.NewIcon(icons.ActionSettings)
		return icon
	}()

	OtherIcon *widget.Icon = func() *widget.Icon {
		icon, _ := widget.NewIcon(icons.ActionHelp)
		return icon
	}()

	HeartIcon *widget.Icon = func() *widget.Icon {
		icon, _ := widget.NewIcon(icons.ActionFavorite)
		return icon
	}()

	PlusIcon *widget.Icon = func() *widget.Icon {
		icon, _ := widget.NewIcon(icons.ContentAdd)
		return icon
	}()

	EditIcon *widget.Icon = func() *widget.Icon {
		icon, _ := widget.NewIcon(icons.ContentCreate)
		return icon
	}()
)
