package cmd

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/3xcellent/github-metrics/client"
	"github.com/3xcellent/github-metrics/config"
	"github.com/3xcellent/github-metrics/metrics"
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
	ctx := context.Background()
	var err error
	cfg, err := config.NewDefaultConfig()
	if err != nil {
		return err
	}
	cfg.GithubClient, err = client.New(ctx, cfg.API)
	if err != nil {
		return err
	}

	boardCfg, err := cfg.GetBoardConfig(args[0])
	if err != nil {
		return err
	}

	project, err := cfg.GithubClient.GetProject(ctx, boardCfg.BoardID)
	if err != nil {
		return err
	}

	if cfg.CreateFile {
		cfg.OutputPath = fmt.Sprintf("%s_%s_%d-%02d.csv",
			strings.Replace(project.GetName(), " ", "_", -1),
			c.Name(),
			cfg.StartDate.Year(),
			cfg.StartDate.Month())
	}

	projectColumns := cfg.GithubClient.GetProjectColumns(ctx, boardCfg.BoardID)
	cols := metrics.NewColumnsMetric(boardCfg, projectColumns, logrus.Debug)
	cols.GatherAndProcessIssues(ctx, cfg.GithubClient)

	var writer *csv.Writer
	if cfg.OutputPath == "" {
		writer = csv.NewWriter(c.OutOrStdout())
	} else {
		writer = csv.NewWriter(cfg.OutPath())
	}
	defer writer.Flush()

	if !cfg.NoHeaders {
		logrus.Debugf("option: headers")
		headers := []string{"Day"}
		for i := cfg.StartColumnIndex; i <= cfg.EndColumnIndex; i++ {
			headers = append(headers, projectColumns[i].GetName())
		}
		if err := writer.Write(headers); err != nil {
			logrus.Fatalf("error writing data to file")
		}
	}
	logrus.Debug(cols.Dump())

	//for currentDate := cfg.StartDate; currentDate.Before(cfg.EndDate); currentDate = currentDate.AddDate(0, 0, 1) {
	//	logrus.Debugf("currentDate: %s", currentDate.String())
	//	dateRow := []string{currentDate.Format(metrics.DateKeyFmt)}
	//	for i := cfg.StartColumnIndex; i <= cfg.EndColumnIndex; i++ {
	//		appendVal := "0"
	//		val, found := cols.DateColumn(currentDate, cols.ColumnNames[i])
	//		if found {
	//			appendVal = strconv.Itoa(val)
	//		}
	//		dateRow = append(dateRow, appendVal)
	//	}
	//	if err := writer.Write(dateRow); err != nil {
	//		logrus.Fatalf("error writing data to file")
	//	}
	//}
	for _, rowValues := range cols.RowValues() {
		if err := writer.Write(rowValues); err != nil {
			return err
		}
	}

	c.Println()
	if cfg.OutputPath != "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		c.Printf("Wrote to: file://%s/%s\n", wd, cfg.OutputPath)
	}

	return nil
}
