package cmd

import (
	"context"
	"log"
	"os"

	"gioui.org/app"
	"github.com/3xcellent/github-metrics/config"
	"github.com/3xcellent/github-metrics/gui"
	"github.com/spf13/cobra"
)

var guiCmd = &cobra.Command{
	Use:   "gui",
	Short: "starts the gui",
	Long:  "starts the gui interface; currently in alpha",
	RunE:  startGUI,
}

func startGUI(c *cobra.Command, args []string) error {

	ctx := context.Background()

	var err error
	if Config == nil {
		Config, err = config.NewDefaultConfig()
		if err != nil {
			panic(err)
		}
	}
	go func() {
		if err := gui.Start(ctx, Config.API); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()

	app.Main()

	return nil
}
