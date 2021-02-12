package cmd

import (
	"context"
	"os"
	"time"

	"github.com/3xcellent/github-metrics/client"
	"github.com/3xcellent/github-metrics/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// MetricsCommand - the root cobra.Command
	MetricsCommand = &cobra.Command{
		Use:   "github-metrics",
		Short: "github-metrics",
		Long:  `Github Metrics gathers data from a github server and generates csv reports`,
		Args:  cobra.MinimumNArgs(1),
	}
	verbose     bool
	askForDate  bool
	token       string
	year        int
	month       int
	issueNumber int
	noHeaders   bool
	outpath     string
	repoName    string
	newFile     bool

	// Config - instance of the config for CLI
	Config *config.AppConfig
)

func init() {

	MetricsCommand.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	MetricsCommand.PersistentFlags().StringVarP(&token, "token", "t", "", "Auth token used when connecting to github server")
	MetricsCommand.PersistentFlags().BoolVarP(&askForDate, "askForDate", "d", false, "command will ask for user to input year and month parameters at runtime")
	MetricsCommand.PersistentFlags().IntVarP(&year, "year", "y", time.Now().Year(), "specify year")
	MetricsCommand.PersistentFlags().IntVarP(&month, "month", "m", int(time.Now().Month()), "specify month")
	MetricsCommand.PersistentFlags().BoolVarP(&noHeaders, "no-headers", "", false, "disable csv header row")
	MetricsCommand.PersistentFlags().StringVarP(&outpath, "outpath", "o", "", "set output path")
	MetricsCommand.PersistentFlags().StringVarP(&repoName, "repoName", "r", "", "repoName (use with repoName)")
	MetricsCommand.PersistentFlags().IntVarP(&issueNumber, "issueNumber", "i", 0, "issueNumber (use with issueNumber)")
	MetricsCommand.PersistentFlags().BoolVarP(&newFile, "create-file", "c", false, "set outpath path to [board_name]_[command_name]_[year]_[month].csv)")

	viper.BindPFlag("verbose", MetricsCommand.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("api.token", MetricsCommand.PersistentFlags().Lookup("token"))
	viper.BindPFlag("askForDate", MetricsCommand.PersistentFlags().Lookup("askForDate"))
	viper.BindPFlag("year", MetricsCommand.PersistentFlags().Lookup("year"))
	viper.BindPFlag("month", MetricsCommand.PersistentFlags().Lookup("month"))
	viper.BindPFlag("noHeaders", MetricsCommand.PersistentFlags().Lookup("no-headers"))
	viper.BindPFlag("outputPath", MetricsCommand.PersistentFlags().Lookup("outpath"))
	viper.BindPFlag("createFile", MetricsCommand.PersistentFlags().Lookup("create-file"))
	viper.BindPFlag("repoName", MetricsCommand.PersistentFlags().Lookup("repoName"))
	viper.BindPFlag("issueNumber", MetricsCommand.PersistentFlags().Lookup("issueNumber"))

	MetricsCommand.AddCommand(
		guiCmd,
		orgsCommand,
		projectCommand,
		projectsCommand,
		issuesCmd,
		columnsCmd,
		pullRequestsCmd,
		reposCommand,
	)
}

// Execute runs the command
func Execute() {
	if err := MetricsCommand.Execute(); err != nil {
		logrus.Debug(err)
		os.Exit(1)
	}
}

// SetupCLI - will return context
func SetupCLI(ctx context.Context, runCfgName string) (*client.MetricsClient, config.RunConfig, error) {
	cfg, err := config.NewDefaultConfig()
	if err != nil {
		return nil, config.RunConfig{}, err
	}

	client, err := client.New(ctx, cfg.API)
	if err != nil {
		return nil, config.RunConfig{}, err
	}

	runCfg, err := cfg.GetRunConfig(runCfgName)
	if err != nil {
		return nil, config.RunConfig{}, err
	}
	return client, runCfg, nil
}
