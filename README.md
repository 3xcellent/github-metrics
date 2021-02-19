# github-metrics

a cli client for generating metrics reports from github project boards and issues.

Run without arguments for help

```bash
    github-metrics
```

#To Use

1. create a `config.yaml` file in the folder you run the binary that looks like:
   ```yaml
   ---
   API:
   Owner: 3xcellent
   RunConfigs:
   - name: github-metrics
       owner: 3xcellent
       projectID: 10966824
       startColumn: In progress
   IncludeHeaders: true
   GroupName: Github
   Owner: 3xcellent
   LoginNames:
   - 3xcellent
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
1. As an environment variable $GH*METRICS_API_TOKEN; supports \_envFile* (see below)

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
