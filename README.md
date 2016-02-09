# jira-influx: Write JIRA issue metrics to InfluxDB
This tool counts issues for different JQL Queries and writes the results and their query duration to InfluxDB using measurement `issue_count` and custom tags.
In addition to showing trends about issue counts, the measurements can be a valuable source of information about a JIRA instances' performance.

## Installation
Configure your `GOPATH` and run `go get github.com/cbonitz/jira-influx`

## Usage
Create a configuration file, then just run `jira-influx` in the same directory.
This will create one measurement for each query.
Use your favorite job scheduler to do this at the interval of your choice.

## Configuration
Create a `config.json` file with the same structure as `config.json.sample`.
Most configuration parameters are self explanatory, please see the sample file for details and the following hints:
* JIRA: `jiraUrl` - the JIRA base URL, `jiraUsername`, `jiraPassword`, and `jiraPauseMilliseconds` an (optional) pause between JIRA queries to avoid creating too much load.
* InfluxDB: `influxUrl` - the Influx base URL, `InfluxDB` - the database to use, `influxUsername` (optional), `influxPassword` (optional)
* Queries: Each query has `jql`, a JQL query, and `tags`, the tags used for InfluxDB