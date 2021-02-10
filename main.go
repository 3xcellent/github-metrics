package main

import (
	"log"
	"os"

	"gioui.org/app"
	"github.com/3xcellent/github-metrics/cmd"
	"github.com/3xcellent/github-metrics/config"
	"github.com/3xcellent/github-metrics/gui"
	"github.com/sirupsen/logrus"
)

var (
	//TODO add correct inputs
	cfg *config.AppConfig
)

func main() {
	if len(os.Args) > 1 && os.Args[1][0:1] != "-" {
		cmd.Execute()
	} else {
		if len(os.Args) > 1 && os.Args[1][0:1] != "-" {
			if os.Args[1] == "-v" {
				logrus.SetLevel(logrus.DebugLevel)
			}
		}
		go func() {
			w := app.NewWindow()
			if err := gui.Loop(w); err != nil {
				log.Fatal(err)
			}
			os.Exit(0)
		}()
		app.Main()
	}
}
