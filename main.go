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
	host := config["jiraHost"].(string)
	username := config["jiraUsername"].(string)
	password := config["jiraPassword"].(string)

	client := &http.Client{}
	params := url.Values{}
	params.Add("jql", jql)
	params.Add("fields", "key")
	req, err := http.NewRequest("GET", host+"/rest/api/latest/search?"+params.Encode(), nil)
	req.SetBasicAuth(username, password)
	resp, err := client.Do(req)
	checkError(err)
	defer resp.Body.Close()
	rawBody, err := ioutil.ReadAll(resp.Body)
	checkError(err)
	var jsonResult interface{}
	err = json.Unmarshal(rawBody, &jsonResult)
	checkError(err)
	m := jsonResult.(map[string]interface{})
	return int(m["total"].(float64))
}

func createInfluxClient(config map[string]interface{}) client.Client {
	host := config["influxHost"].(string)
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

func addPoint(batchPoints client.BatchPoints, rawTags map[string]interface{}, count int) {
	tags := map[string]string{}
	for key, value := range rawTags {
		tags[key] = value.(string)
	}
	fields := map[string]interface{}{"count": count}
	pt, err := client.NewPoint("issue_count", tags, fields, time.Now())
	checkError(err)
	fmt.Printf("%v: %d issues\n", tags, count)
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
	for _, queryObject := range queries {
		q := queryObject.(map[string]interface{})
		// run jira query
		jql := q["jql"].(string)
		count := runJqlQuery(config, jql)
		// create influx point and save for later
		addPoint(batchPoints, q["tags"].(map[string]interface{}), count)
	}
	fmt.Println("Writing data to InfluxDB")
	// write the points
	influxClient.Write(batchPoints)

}
