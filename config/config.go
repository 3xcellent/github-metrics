package config

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type MetricsClient interface {
	GetIssue(ctx context.Context, repoOwner, repoName string, issueNumber int) (*github.Issue, error)
	GetProject(ctx context.Context, boardID int64) (*github.Project, error)
	GetProjectColumns(ctx context.Context, projectID int64) []*github.ProjectColumn
	GetPullRequests(ctx context.Context, repoOwner, repoName string) ([]*github.PullRequest, error)
	GetIssuesFromColumn(ctx context.Context, repoOwner string, columnID int64, beginDate, endDate time.Time) map[string][]*github.Issue
	GetIssueEvents(ctx context.Context, repoOwner, repoName string, issueNumber int) ([]*github.IssueEvent, error)
	GetReposFromIssuesOnColumn(ctx context.Context, columnId int64) []string
}

type Configuration struct {
	GithubClient     MetricsClient
	API              APIConfig
	Boards           map[string]Board
	NoHeaders        bool
	OutputPath       string
	CreateFile       bool
	StartColumnIndex int
	StartColumn      string
	EndColumn        string
	EndColumnIndex   int
	IssueNumber      int
	RepoName         string
	Owner            string
	StartDate        time.Time
	EndDate          time.Time
	Verbose          bool
	Timezone         *time.Location
	LoginNames       []string
	GroupName        string
}

type APIConfig struct {
	Token     string
	BaseURL   string
	UploadURL string
}

type Board struct {
	StartColumn string
	EndColumn   string
	Owner       string
	BoardID     int64
}

func (c *Configuration) CreatedByGroup(name string) string {
	for _, loginName := range c.LoginNames {
		if name == loginName {
			return c.GroupName
		}
	}
	return ""
}

func (c *Configuration) OutPath() *os.File {
	if c.OutputPath == "" {
		return os.Stdout
	}

	logrus.Debugf("writing to: %s", c.OutputPath)
	output, err := os.Create(c.OutputPath)
	if err != nil {
		panic(err)
	}
	return output
}

func (c *Configuration) GetBoard(name string) (Board, error) {
	if len(c.Boards) == 0 {
		return Board{}, errors.New("no project boards configured")
	}

	board, found := c.Boards[name]
	if !found {
		return Board{}, errors.New("no project board found with that name")
	}
	return board, nil
}

func newConfigFromEnv() (*Configuration, error) {
	loadedEnvFile := true
	err := godotenv.Load()
	if err != nil {
		loadedEnvFile = false
		if !os.IsNotExist(err) {
			return nil, errors.Wrap(err, "Error loading .env file")
		}
	}
	logrus.Debugf(".env file loaded: %v", loadedEnvFile)

	viper.SetEnvPrefix("gh_metrics")
	viper.BindEnv("api.token")

	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")      // optionally look for config in the working directory
	viper.AutomaticEnv()

	if err = viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			logrus.Debug("no config file found")
		} else {
			// Config file was found but another error was produced
			return nil, errors.Wrap(err, "error loading config file")
		}
	}
	var cfg Configuration
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode default into struct")
	}
	return &cfg, nil
}

func NewStaticConfig(config []byte) (*Configuration, error) {
	var cfg Configuration

	viper.SetConfigType("yaml")
	err := viper.ReadConfig(bytes.NewBuffer(config))
	if err != nil {
		return nil, errors.Wrap(err, "error loading config")
	}
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode into struct")
	}

	err = cfg.Init()
	return &cfg, err
}

func NewDefaultConfig() (*Configuration, error) {
	cfg, err := newConfigFromEnv()
	if err != nil {
		return nil, errors.Wrap(err, "error initializing default config")
	}
	err = cfg.Init()
	return cfg, err
}

func (c *Configuration) Init() error {
	if c.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	year := viper.GetInt("year")
	month := viper.GetInt("month")

	if viper.GetBool("askForDate") || strings.ToUpper(os.Getenv("GHM_ASK_FOR_DATE")) == "TRUE" {
		reader := bufio.NewReader(os.Stdin) //create new reader, assuming bufio imported
		print("Year: ")
		yearInput, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		yearInt, err := strconv.ParseInt(yearInput[:len(yearInput)-1], 10, 32)
		if err != nil {
			return err
		}
		year = int(yearInt)

		print("Month: ")
		monthInput, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		monthInt, err := strconv.ParseInt(monthInput[:len(monthInput)-1], 10, 32)
		if err != nil {
			return err
		}
		month = int(monthInt)
	}

	c.Timezone, _ = time.LoadLocation("Local")
	c.StartDate = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, c.Timezone)
	c.EndDate = c.StartDate.AddDate(0, 1, 0)
	if c.StartDate.After(time.Now()) {
		panic("begin date cannot be in the future")
	}
	if c.EndDate.After(time.Now()) {
		now := time.Now()
		c.EndDate = time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, c.Timezone)
	}

	logrus.Debugf("Retrieving cards in range: %s - %s", c.StartDate.String(), c.EndDate.String())
	logrus.Debugf("Github Metrics Configuration:\n%#v\n", c)

	return nil
}

func (c *Configuration) SetMetricsClient(metricsClient MetricsClient) {
	c.GithubClient = metricsClient
}
