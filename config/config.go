package config

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"sort"
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
	GetIssues(ctx context.Context, repoOwner string, reposNames []string, beginDate, endDate time.Time) []*github.Issue
	GetIssueEvents(ctx context.Context, repoOwner, repoName string, issueNumber int) ([]*github.IssueEvent, error)
	GetRepos(ctx context.Context, columnId int64) []string
}

type Config struct {
	GithubClient     MetricsClient
	API              APIConfig
	Boards           map[string]BoardConfig
	NoHeaders        bool
	OutputPath       string
	CreateFile       bool
	StartColumnIndex int
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

type BoardConfig struct {
	Name        string
	StartColumn string
	StartDate   time.Time
	EndColumn   string
	EndDate     time.Time
	Owner       string
	BoardID     int64
}

func (c *Config) CreatedByGroup(name string) string {
	for _, loginName := range c.LoginNames {
		if name == loginName {
			return c.GroupName
		}
	}
	return ""
}

func (c *Config) OutPath() *os.File {
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

func (c *Config) GetBoardConfig(name string) (BoardConfig, error) {
	if len(c.Boards) == 0 {
		return BoardConfig{}, errors.New("no project boards configured")
	}

	logrus.Infof("Config.Owner: %s", c.Owner)
	board, found := c.Boards[name]
	if !found {
		return BoardConfig{}, errors.New("no project board found with that name")
	}

	board.Name = name
	board.StartDate = c.StartDate
	board.EndDate = c.StartDate.AddDate(0, 1, 0)
	if board.Owner == "" {
		board.Owner = c.Owner
	}

	now := time.Now()
	if board.EndDate.After(now) {
		board.EndDate = time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, c.Timezone)
	}

	return board, nil
}

func (c *Config) GetSortedBoards() []BoardConfig {
	sortedBoardNames := make([]string, 0, len(c.Boards))
	for k, _ := range c.Boards {
		sortedBoardNames = append(sortedBoardNames, k)
	}
	sort.Strings(sortedBoardNames)
	boards := make([]BoardConfig, 0, len(c.Boards))
	for _, k := range sortedBoardNames {
		boards = append(boards, c.Boards[k])
	}
	return boards
}

func newConfigFromEnv() (*Config, error) {
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
	var cfg Config
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode default into struct")
	}
	return &cfg, nil
}

func NewStaticConfig(config []byte) (*Config, error) {
	var cfg Config

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

func NewDefaultConfig() (*Config, error) {
	cfg, err := newConfigFromEnv()
	if err != nil {
		return nil, errors.Wrap(err, "error initializing default config")
	}
	err = cfg.Init()
	return cfg, err
}

func (c *Config) Init() error {
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
	logrus.Debugf("Github Metrics Config:\n%#v\n", c)

	return nil
}

func (c *Config) SetMetricsClient(metricsClient MetricsClient) {
	c.GithubClient = metricsClient
}

func (c *Config) SetIndexes(projectColumns []*github.ProjectColumn, startColumn, endColumn string) {
	for i, col := range projectColumns {
		colName := col.GetName()
		if colName == startColumn {
			logrus.Debugf("StartColumnIndex: %d", i)
			c.StartColumnIndex = i
		}

		if colName == endColumn {
			logrus.Debugf("EndColumnIndex: %d", i)
			c.EndColumnIndex = i
		}
	}
}
