package main

import (
	"log"
	"os"

	"gioui.org/app"
	"github.com/3xcellent/github-metrics/config"
	"github.com/3xcellent/github-metrics/gui"
)

var (
	//TODO add correct inputs
	cfg *config.Config
)

func main() {

	go func() {
		w := app.NewWindow()
		if err := gui.Loop(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}
