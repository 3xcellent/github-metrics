package config

import (
	"bufio"
	"bytes"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// AppConfig - the Config used by the CLI and GUI apps
type AppConfig struct {
	API         APIConfig
	RunConfigs  RunConfigs
	NoHeaders   bool
	OutputPath  string
	CreateFile  bool
	StartColumn string
	EndColumn   string
	// StartColumnIndex int
	// EndColumnIndex   int
	IssueNumber int
	RepoName    string
	Owner       string
	StartDate   time.Time
	EndDate     time.Time
	Verbose     bool
	Timezone    *time.Location
	LoginNames  []string
	GroupName   string
}

func (c *AppConfig) CreatedByGroup(name string) string {
	for _, loginName := range c.LoginNames {
		if name == loginName {
			return c.GroupName
		}
	}
	return ""
}

func newConfigFromEnv() (*AppConfig, error) {
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
	var cfg AppConfig
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode default into struct")
	}
	return &cfg, nil
}

// NewStaticConfig - returns a config defined with the provided yaml ([]byte)
func NewStaticConfig(config []byte) (*AppConfig, error) {
	var cfg AppConfig

	viper.SetConfigType("yaml")
	err := viper.ReadConfig(bytes.NewBuffer(config))
	if err != nil {
		return nil, errors.Wrap(err, "error loading config")
	}
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode into struct")
	}

	err = cfg.init()
	return &cfg, err
}

// NewDefaultConfig - returns a config from current environment settings
func NewDefaultConfig() (*AppConfig, error) {
	cfg, err := newConfigFromEnv()
	if err != nil {
		return nil, errors.Wrap(err, "error initializing default config")
	}
	err = cfg.init()
	return cfg, err
}

func (c *AppConfig) init() error {
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

// GetRunConfig - finds RunConfig by its name, or err if not found
func (c *AppConfig) GetRunConfig(name string) (RunConfig, error) {
	if len(c.RunConfigs) == 0 {
		return RunConfig{}, errors.New("no project run configs configured")
	}

	for _, runConfig := range c.RunConfigs {
		if runConfig.Name == name {
			rc := runConfig
			if rc.Owner == "" {
				rc.Owner = c.Owner
			}
			if rc.StartColumn == "" {
				rc.StartColumn = c.StartColumn
			}
			if rc.EndColumn == "" {
				rc.EndColumn = c.EndColumn
			}

			rc.RepoName = c.RepoName
			rc.IssueNumber = c.IssueNumber
			rc.CreateFile = c.CreateFile
			rc.NoHeaders = c.NoHeaders
			rc.StartDate = c.StartDate
			rc.EndDate = c.EndDate

			return rc, nil
		}
	}

	return RunConfig{}, errors.New("no run configs found with that name")

}
