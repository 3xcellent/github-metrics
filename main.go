package main

import (
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget/material"
	"github.com/3xcellent/github-metrics/config"
	"github.com/3xcellent/github-metrics/gui"
)

func main() {
	metricsCfg, err := config.NewDefaultConfig()
	if err != nil {
		panic(err)
	}
	gui.InitializeAPIConfig(metricsCfg.API)

	go func() {
		w := app.NewWindow()
		if err := loop(w, metricsCfg); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func loop(w *app.Window, metricsCfg *config.Configuration) error {
	th := material.NewTheme(gofont.Collection())
	var ops op.Ops

	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			gui.Config(gtx, th, metricsCfg)
			e.Frame(gtx.Ops)
		}
	}
}
