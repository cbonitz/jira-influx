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

func main() {
	rawConfig, configErr := ioutil.ReadFile("./config.json")
	checkError(configErr)
	var jsonConfig interface{}
	configErr = json.Unmarshal(rawConfig, &jsonConfig)
	checkError(configErr)
	config := jsonConfig.(map[string]interface{})
	host := config["host"].(string)
	username := config["username"].(string)
	password := config["password"].(string)
	queries := config["queries"].([]interface{})
	for _, queryObject := range queries {
		client := &http.Client{}
		params := url.Values{}
		q := queryObject.(map[string]interface{})
		params.Add("jql", q["jql"].(string))
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
		fmt.Printf("%v: %d issues found\n", q["name"], int(m["total"].(float64)))
	}

}
