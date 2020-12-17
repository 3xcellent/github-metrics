package gui

import (
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"git.sr.ht/~whereswaldon/materials"

	"github.com/3xcellent/github-metrics/config"
)

type (
	// D - i really do not like
	D = layout.Dimensions

	// C -
	C = layout.Context
)

var (
	baseURLInput   = materials.TextField{}
	uploadURLInput = materials.TextField{}
	tokenInput     = materials.TextField{}

	inputAlignment     layout.Alignment
	inputAlignmentEnum widget.Enum

	inset = layout.UniformInset(unit.Dp(8))
	th    = material.NewTheme(gofont.Collection())
)

// InitializeAPIConfig - sets the initial values of the api config
func InitializeAPIConfig(cfg config.APIConfig) {
	baseURLInput.SetText(cfg.BaseURL)
	uploadURLInput.SetText(cfg.UploadURL)
	tokenInput.SetText(cfg.Token)
}

// LayoutAPIConfigPage - layout api config options
func LayoutAPIConfigPage(gtx C, cfg *config.APIConfig) D {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(
		gtx,
		textField(gtx, th, &baseURLInput, &cfg.BaseURL, "Base URL"),
		textField(gtx, th, &uploadURLInput, &cfg.UploadURL, "Upload URL"),
		textField(gtx, th, &tokenInput, &cfg.Token, "Token"),
	)
}

// Config returns the config layout
func Config(gtx layout.Context, th *material.Theme, metricsCfg *config.Configuration) layout.Dimensions {
	return LayoutAPIConfigPage(gtx, &metricsCfg.API)
}

func textField(gtx layout.Context, th *material.Theme, f *materials.TextField, valuePtr *string, label string) layout.FlexChild {
	return layout.Rigid(func(gtx C) D {
		if len(f.Events()) > 0 {
			for _, e := range f.Events() {
				if e, ok := e.(widget.SubmitEvent); ok {
					*valuePtr = e.Text
				}
			}
		}
		f.Alignment = inputAlignment
		return f.Layout(gtx, th, label)
	})
}
