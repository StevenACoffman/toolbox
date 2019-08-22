package main

import (
	"encoding/json"
	"errors"
	"flag"
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

// no args please
func getFlags() []string {
	var args []string
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-"){
			args = append(args,arg)
		}
	}
	return args
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

func getEnvOrDie(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	panic("You must set your environment variable " + key)
}

func main() {
	//jiraUserId := getEnv("JIRA_LOGIN", "login")
	//jiraPassword := getEnv("JIRA_PASSWORD", "password")
	//jiraBaseURL := getEnvOrDie("JIRA_BASE_URL", "https://jira.jstor.org")
	//jiraAPIURI := getEnv("JIRA_API_URI", "/rest/api/2/issue/")
	jiraUserId := getEnvOrDie("JIRA_LOGIN")
	jiraPassword := getEnvOrDie("JIRA_PASSWORD")
	jiraBaseURL := getEnvOrDie("JIRA_BASE_URL")
	jiraAPIURI := getEnv("JIRA_API_URI", "/rest/api/2/issue/")

	ticket := ""
	args := getArgs()


	if len(args) == 0 {
		fmt.Fprintln(os.Stderr,"Usage: wti <ticket> --resolves=true")
		panic(errors.New("Usage: wti <ticket> --resolves=true"))
	} else {
		ticket = args[0]
	}

	resolvesFlag := flag.Bool("resolves", false, "insert resolves link")

	flag.CommandLine.Parse(getFlags())

	url := jiraBaseURL+jiraAPIURI+ticket

	jiraClient, req := BuildRequest(url, jiraUserId, jiraPassword)

	jiraResponse := GetJiraResponse(jiraClient, req)

	fmt.Printf("%s - %s\n\n", jiraResponse.Key, jiraResponse.Fields.Summary)
	if *resolvesFlag {
		fmt.Printf("Resolves [%s|%s%s%s]\n\n", jiraResponse.Key, jiraBaseURL,"/browse/", jiraResponse.Key)
	}
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