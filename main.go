package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func runJqlQuery(config map[string]interface{}, jql string) int {
	username := config["jiraUsername"].(string)
	password := config["jiraPassword"].(string)
	host := config["jiraHost"].(string)

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

func main() {
	rawConfig, configErr := ioutil.ReadFile("./config.json")
	checkError(configErr)
	var jsonConfig interface{}
	configErr = json.Unmarshal(rawConfig, &jsonConfig)
	checkError(configErr)
	config := jsonConfig.(map[string]interface{})
	queries := config["queries"].([]interface{})
	for _, queryObject := range queries {
		q := queryObject.(map[string]interface{})
		jql := q["jql"].(string)
		measurement := q["name"]
		count := runJqlQuery(config, jql)
		fmt.Printf("%v: %d issues found\n", measurement, count)
	}

}
