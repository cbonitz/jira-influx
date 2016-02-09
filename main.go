package main

import (
	"encoding/json"
	"fmt"
	"github.com/influxdb/influxdb/client/v2"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func runJqlQuery(config map[string]interface{}, jql string) int {
	host := config["jiraUrl"].(string)
	username := config["jiraUsername"].(string)
	password := config["jiraPassword"].(string)

	// Create the authenticated  HTTP request
	client := &http.Client{}
	params := url.Values{}
	params.Add("jql", jql)
	// we actually only care about the total atm.
	// specifying one field reduces the amount of useless data in the response
	params.Add("fields", "key")
	req, err := http.NewRequest("GET", host+"/rest/api/latest/search?"+params.Encode(), nil)
	req.SetBasicAuth(username, password)
	resp, err := client.Do(req)
	checkError(err)

	// Read and parse JSON body
	defer resp.Body.Close()
	rawBody, err := ioutil.ReadAll(resp.Body)
	checkError(err)
	var jsonResult interface{}
	err = json.Unmarshal(rawBody, &jsonResult)
	checkError(err)
	m := jsonResult.(map[string]interface{})

	// extract the interesting data
	return int(m["total"].(float64))
}

func createInfluxClient(config map[string]interface{}) client.Client {
	host := config["influxUrl"].(string)
	username := ""
	password := ""
	if config["influxPassword"] != nil {
		username = config["influxUsername"].(string)
		password = config["influxPassword"].(string)
	}
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     host,
		Username: username,
		Password: password,
	})
	checkError(err)
	return c
}

func createBatchPoints(config map[string]interface{}, c client.Client) client.BatchPoints {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  config["influxDB"].(string),
		Precision: "s",
	})
	checkError(err)
	return bp
}

// addPoint adds a point with tags to a BatchPoints object for sending them later
func addPoint(batchPoints client.BatchPoints, timeOfQuery time.Time, rawTags map[string]interface{}, count int, durationMilliseconds int64) {
	// put tags from config into the right type of map
	tags := map[string]string{}
	for key, value := range rawTags {
		tags[key] = value.(string)
	}
	fields := map[string]interface{}{
		"count":                count,
		"durationMilliseconds": durationMilliseconds}
	// for now, the measurement name is fixed
	pt, err := client.NewPoint("issue_count", tags, fields, timeOfQuery)
	checkError(err)
	fmt.Printf("Prepared for sending: %v: %d issues (in %d ms)\n", tags, count, durationMilliseconds)
	batchPoints.AddPoint(pt)
}

func main() {
	// Read json config
	rawConfig, configErr := ioutil.ReadFile("./config.json")
	checkError(configErr)
	var jsonConfig interface{}
	configErr = json.Unmarshal(rawConfig, &jsonConfig)
	checkError(configErr)
	config := jsonConfig.(map[string]interface{})
	queries := config["queries"].([]interface{})
	// create influx client and batch points (only one send operation at the end)
	influxClient := createInfluxClient(config)
	batchPoints := createBatchPoints(config, influxClient)
	durationBetweenJiraQueries := time.Duration(100) * time.Millisecond
	if val, defined := config["jiraPauseMilliseconds"]; defined {
		durationBetweenJiraQueries = time.Duration(int(val.(float64))) * time.Millisecond
	}
	for _, queryObject := range queries {
		q := queryObject.(map[string]interface{})
		// run jira query
		jql := q["jql"].(string)

		timeBeforeQuery := time.Now()
		count := runJqlQuery(config, jql)
		timeAfterQuery := time.Now()

		queryDurationMilliseconds := timeAfterQuery.Sub(timeBeforeQuery).Nanoseconds() / 1000000
		// create influx point and save for later
		addPoint(batchPoints, timeBeforeQuery, q["tags"].(map[string]interface{}), count, queryDurationMilliseconds)
		time.Sleep(durationBetweenJiraQueries)
	}
	fmt.Println("Writing data to InfluxDB")
	// write the points
	influxErr := influxClient.Write(batchPoints)
	checkError(influxErr)
}
