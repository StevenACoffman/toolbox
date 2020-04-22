///usr/bin/env go run "$0" "$@" ; exit "$?"
package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	logger "log"
	"net/http"
	"os"
)

func main() {
	log := logger.New(os.Stderr, "", 0)
	//replace with what you actually have
	jiraUserID := "steve@khanacademy.org"
	jiraPassword := os.ExpandEnv("${JIRA_API_TOKEN}")
	jiraBaseURL := "https://khanacademy.atlassian.net"
	jiraAPIURI := "/rest/api/2/issue/"

	ticket := ""
	//I don't need the program name
	args := os.Args[1:]

	if len(args) == 0 {
		log.Println("Usage: wti <ticket>")
		panic(errors.New("Usage: wti <ticket>"))
	} else {
		ticket = args[0]
	}

	response, _ := getJiraTicket(jiraBaseURL, jiraAPIURI, ticket, jiraUserID, jiraPassword)

	log.Printf("%s - %s\n\n", response.Key, response.Fields.Summary)
	log.Println(response.Fields.Description)
}

// JIRAFields contains the JIRA fields for the issue
type JIRAFields struct {
	Description string `json:"description"`
	Summary     string `json:"summary"`
}

// JIRAResponse holds the deserialized JSON response from JIRA
type JIRAResponse struct {
	Key    string     `json:"key"`
	Fields JIRAFields `json:"fields"`
}

// This is a "glue" function.  It takes all of the more testable, behavioral
// functions and "glues" them together without any other inherint behavior
func getJiraTicket(jiraBaseURL, jiraAPIURI, ticket, jiraUserID, jiraPassword string) (JIRAResponse, error) {
	url := GenerateURL(jiraBaseURL, jiraAPIURI, ticket)
	jiraClient := http.Client{}
	req := BuildRequest(url, jiraUserID, jiraPassword)
	response, _ := jiraClient.Do(req)
	body := GetBody(response)
	logger.Print(body)
	return ParseJiraResponse(body)
}

// GenerateURL will construct the JIRA API call from components
func GenerateURL(jiraBaseURL, jiraAPIURI, ticket string) string {
	return jiraBaseURL + jiraAPIURI + ticket
}

// BuildRequest will build a new client and request with the proper
// headers, including basic authentication
func BuildRequest(url string, jiraUserID string, jiraPassword string) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.SetBasicAuth(jiraUserID, jiraPassword)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	return req
}

// GetBody will take an httpResponse and extract the body as a string
func GetBody(res *http.Response) string {
	defer checkedResponseBodyClose(res)
	body, _ := ioutil.ReadAll(res.Body)
	return string(body)
}

func checkedResponseBodyClose(response *http.Response) {
	err := response.Body.Close()
	if err != nil {
		logger.Println(err)
	}
}

// ParseJiraResponse will parse the Jira response into a JIRAResponse
func ParseJiraResponse(jsonData string) (JIRAResponse, error) {
	jiraResponse := JIRAResponse{}

	jsonErr := json.Unmarshal([]byte(jsonData), &jiraResponse)
	if jsonErr != nil {
		logger.Println(jsonErr)
		return JIRAResponse{}, jsonErr
	}

	return jiraResponse, nil
}
