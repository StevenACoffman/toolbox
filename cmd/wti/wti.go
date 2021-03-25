///usr/bin/env go run "$0" "$@" ; exit "$?"
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	logger "log"
	"net/http"
	"os"
	"strings"
)

func main() {
	log := logger.New(os.Stderr, "", 0)
	// replace with what you actually have
	jiraUserID := "steve@khanacademy.org"
	jiraPassword := os.ExpandEnv("${JIRA_API_TOKEN}")
	jiraBaseURL := "https://khanacademy.atlassian.net"
	jiraAPIURI := "/rest/api/2/issue/"

	ticket := ""
	// I don't need the program name
	args := os.Args[1:]
	var title, summary bool
	flag.BoolVar(&title, "title", false, "Just Print Title")
	flag.BoolVar(&summary, "summary", false, "Just Print Summary")

	flagErr := flag.CommandLine.Parse(getFlags())
	// check if we get a flag parse Error (e.g. missing required or
	// unrecognized)
	if flagErr != nil {
			fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
			panic(errors.New("Usage: wti <ticket>"))
			flag.PrintDefaults()
	}

	if len(args) == 0 {
		log.Println("Usage: wti <ticket>")
		panic(errors.New("Usage: wti <ticket>"))
	} else {
		ticket = args[0]
	}

	response, _ := getJiraTicket(jiraBaseURL, jiraAPIURI, ticket, jiraUserID, jiraPassword)

	if !title && !summary {
		log.Printf("%s - %s\n\n", response.Key, response.Fields.Summary)
		log.Println(response.Fields.Description)
	} else if title {
		log.Printf("%s - %s\n\n", response.Key, response.Fields.Summary)
	} else if summary {
		log.Println(response.Fields.Description)
	}

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

// flags but no args
func getFlags() []string {
	var args []string
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-") {
			args = append(args, arg)
		}
	}
	return args
}

