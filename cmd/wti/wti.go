package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type JIRAResponse struct {
	Key    string `json:"key"`
	Fields struct {
		Description string `json:"description"`
		Summary     string `json:"summary"`
	} `json:"fields"`
}

// no flags please, also I don't need the program name
func getArgs() []string {
	var args []string
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-"){
			continue
		}
		args = append(args, arg)
	}
	return args
}

func main() {
	jiraUserId := getEnv("JIRA_LOGIN", "login")
	jiraPassword := getEnv("JIRA_PASSWORD", "password")
	jiraBaseURL := getEnv("JIRA_BASE_URL", "https://jira.jstor.org/rest/api/2/issue/")
	ticket := "CORE-5339"
	args := getArgs()

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr,"Usage: lookup <appname> <environment> --list=true")
		panic(errors.New("Usage: lookup <appname> <environment> --list=true"))
	} else {
		ticket = args[0]
	}

	if len(args) == 0 {

	}
	url := jiraBaseURL+ticket

	jiraClient, req := BuildRequest(url, jiraUserId, jiraPassword)

	jiraResponse := GetJiraResponse(jiraClient, req)

	fmt.Println(jiraResponse.Key)
	fmt.Println(jiraResponse.Fields.Summary)
	fmt.Println(jiraResponse.Fields.Description)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func BuildRequest(url string, jiraUserId string, jiraPassword string) (http.Client, *http.Request) {
	jiraClient := http.Client{
		Timeout: time.Second * 15, // Maximum of 15 secs
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	req.SetBasicAuth(jiraUserId, jiraPassword)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	return jiraClient, req
}

func GetJiraResponse(jiraClient http.Client, req *http.Request) JIRAResponse {
	jiraResponse := JIRAResponse{}
	res, getErr := jiraClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
		panic(getErr)
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
		panic(readErr)
	}

	jsonErr := json.Unmarshal(body, &jiraResponse)
	if jsonErr != nil {
		log.Fatal(jsonErr)
		panic(getErr)
	}
	return jiraResponse
}