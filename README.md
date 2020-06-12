# github-metrics
a cli client for generating metrics reports from github project boards and issues.

```
Github Metrics gathers data from a github server and generates csv reports

Usage:
  github-metrics [command]

Available Commands:
  columns       output number of issues in each column for a github board to csv
  help          Help about any command
  issues        gathers metrics from issues on a board and outputs as csv
  project       shows information about a specific project
  pull_requests output number of issues in each column for a github board to csv
  repos         shows list of repos for a specific project

Flags:
  -d, --askForDate       command will ask for user to input year and month parameters at runtime
  -c, --create-file      set outpath path to [board_name]_[command_name]_[year]_[month].csv)
  -h, --help             help for github-metrics
  -m, --month int        specify month (default 6)
      --no-headers       disable csv header row
  -o, --outpath string   set output path
  -t, --token string     Auth token used when connecting to github server
  -v, --verbose          verbose output
  -y, --year int         specify year (default 2020)

Use "github-metrics [command] --help" for more information about a command.

```
#To Use
1. create a `config.yaml` file in the folder you run the binary that looks like:
    ```yaml
   ---
     API:
       BaseURL: https://my.github.server/api/v3
     Owner: 3xcellent
     IncludeHeaders: true
     Boards:
       - MyBoard:
           boardID: 1234
           startColumn: Develop
       - MyOtherBoard:
           boardID: 4321
           startColumn: Develop
           endColumn: Ready For Demo
     GroupName: An 3xcellent Team
     LoginNames:
       - 3xcellent
       - t3h.dud3
       - bill.gates
   ```    
1. Then run:
    ```bash
    github-metrics issues MyBoard \
    --token $MY_GITHUB_ACCESS_TOKEN \
    --year 2020 \
    --month 1 \
    ```
    which outputs something like:
    ```csv
    Card #,Team,Type,Description,Develop,Code Review,Deploying to Nonprod,Deploying to Prod,Done Placeholder,Development Days,Feature?,Blocked?,Blocked Days
    1234,github-metrics,Bug,the README is weak; needs details,01/01/20,01/02/20,01/03/20,01/05/20,01/08/20,6.8,false,false,0
    ```

# Generating a Github Access Token 
The token can be provided in two different ways
1. As a command line option `-t` or `--token`
1. As an environment variable $GH_METRICS_API_TOKEN; supports *envFile* (see below)

To create your access token
1. Go to `https://my.github.server/settings/tokens` and `Generate new token`
1. Provide the following scopes:
      - repo:status
      - repo_deployment
      - public_repo
      - repo:invite
      - read:org
      - read:public_key
      - read:repo_hook
      - read:user
      - read:discussion

# envFile
Create a `.env` file in folder you run from that contains:

    GH_METRICS_API_TOKEN=<your token>
    
It is recommended to protect the API access token from any source code managment systems (git, svn) 
if you do want to commit the `config.yaml`
