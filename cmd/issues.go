package cmd

import (
	"encoding/csv"
	"os"

	"github.com/3xcellent/github-metrics/metrics/runners"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var issuesCmd = &cobra.Command{
	Use:   "issues [board_name]",
	Short: "gathers metrics from issues on a board and outputs as csv",
	Long:  "gathers issues from a github repoName board, calculates column and blocked durations, and outputs as comma separated values (.csv)",
	RunE:  issues,
	Args:  cobra.MinimumNArgs(1),
}

func issues(c *cobra.Command, args []string) error {
	ctx := c.Context()

	client, runCfg, err := SetupCLI(ctx, args[0])
	if err != nil {
		return err
	}

	runCfg.MetricName = "issues"

	runner, err := runners.New(runCfg, client)
	if err != nil {
		return err
	}

	err = runner.Run(ctx)
	if err != nil {
		return err
	}

	outpath := runner.RunName()
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

	for _, rowValues := range runner.Values() {
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
