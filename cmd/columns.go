package cmd

import (
	"encoding/csv"
	"os"

	"github.com/3xcellent/github-metrics/metrics/runners"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var columnsCmd = &cobra.Command{
	Use:   "columns [board_name]",
	Short: "output number of issues in each column for a github board to csv",
	Long:  "aggregate column totals for a github repoName board within year and month provided (default is current year and month)",
	RunE:  columns,
	Args:  cobra.MinimumNArgs(1),
}

func columns(c *cobra.Command, args []string) error {
	ctx := c.Context()

	client, runCfg, err := SetupCLI(ctx, args[0])
	if err != nil {
		return err
	}

	colsRunner := runners.NewColumnsRunner(runCfg, client)
	colsRunner.LogFunc = logrus.Debug
	err = colsRunner.Run(ctx)
	if err != nil {
		return err
	}
	outpath := colsRunner.RunName()
	var writer *csv.Writer
	if runCfg.CreateFile {
		logrus.Debugf("writing to: %s", outpath)
		output, err := os.Create(outpath)
		if err != nil {
			panic(err)
		}
		writer = csv.NewWriter(output)
	} else {
		writer = csv.NewWriter(c.OutOrStdout())
	}
	defer writer.Flush()

	for _, rowValues := range colsRunner.Values() {
		if err := writer.Write(rowValues); err != nil {
			return err
		}
	}

	c.Println()
	if runCfg.CreateFile {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		c.Printf("Wrote to: file://%s/%s\n", wd, outpath)
	}

	return nil
}
