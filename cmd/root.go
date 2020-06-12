package cmd

import (
	"os"
	"time"

	"github.com/3xcellent/github-metrics/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	GithubMetricsCmd = &cobra.Command{
		Use:   "github-metrics",
		Short: "github-metrics",
		Long:  `Github Metrics gathers data from a github server and generates csv reports`,
		Args:  cobra.MinimumNArgs(1),
	}
	Config     *config.Configuration
	verbose    bool
	askForDate bool
	token      string
	year       int
	month      int
	noHeaders  bool
	outpath    string
	newFile    bool
)

func init() {

	GithubMetricsCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	GithubMetricsCmd.PersistentFlags().StringVarP(&token, "token", "t", "", "Auth token used when connecting to github server")
	GithubMetricsCmd.PersistentFlags().BoolVarP(&askForDate, "askForDate", "d", false, "command will ask for user to input year and month parameters at runtime")
	GithubMetricsCmd.PersistentFlags().IntVarP(&year, "year", "y", time.Now().Year(), "specify year")
	GithubMetricsCmd.PersistentFlags().IntVarP(&month, "month", "m", int(time.Now().Month()), "specify month")
	GithubMetricsCmd.PersistentFlags().BoolVarP(&noHeaders, "no-headers", "", false, "disable csv header row")
	GithubMetricsCmd.PersistentFlags().StringVarP(&outpath, "outpath", "o", "", "set output path")
	GithubMetricsCmd.PersistentFlags().BoolVarP(&newFile, "create-file", "c", false, "set outpath path to [board_name]_[command_name]_[year]_[month].csv)")

	viper.BindPFlag("verbose", GithubMetricsCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("api.token", GithubMetricsCmd.PersistentFlags().Lookup("token"))
	viper.BindPFlag("askForDate", GithubMetricsCmd.PersistentFlags().Lookup("askForDate"))
	viper.BindPFlag("year", GithubMetricsCmd.PersistentFlags().Lookup("year"))
	viper.BindPFlag("month", GithubMetricsCmd.PersistentFlags().Lookup("month"))
	viper.BindPFlag("noHeaders", GithubMetricsCmd.PersistentFlags().Lookup("no-headers"))
	viper.BindPFlag("outputPath", GithubMetricsCmd.PersistentFlags().Lookup("outpath"))
	viper.BindPFlag("createFile", GithubMetricsCmd.PersistentFlags().Lookup("create-file"))

	GithubMetricsCmd.AddCommand(projectCommand, issuesCmd, columnsCmd, pullRequestsCmd, reposCommand)
}

func Execute() {
	if err := GithubMetricsCmd.Execute(); err != nil {
		logrus.Debug(err)
		os.Exit(1)
	}
}
