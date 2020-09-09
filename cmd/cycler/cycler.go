///usr/bin/env go run "$0" "$@" ; exit "$?"
package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/StevenACoffman/toolbox/cmd/cycler/auth"
	"github.com/StevenACoffman/toolbox/cmd/cycler/httpdoer"
	"github.com/sethgrid/pester"
	"io/ioutil"
	logger "log"
	"net/http"
	"os"
	"os/user"
	"sort"
	"time"

	"github.com/rickar/cal/v2"
	"github.com/rickar/cal/v2/us"
)

func main() {

	c := cal.NewBusinessCalendar()
	c.Name = "Khan Academy"
	c.Description = "Default company calendar"

	// add holidays that the business observes

	// Standard US weekend substitution rules:
	//   Saturdays move to Friday
	//   Sundays move to Monday
	weekendAlt := []cal.AltDay{
		{Day: time.Saturday, Offset: -1},
		{Day: time.Sunday, Offset: 1},
	}
	Juneteenth := &cal.Holiday{
		Name:     "Juneteenth",
		Type:     cal.ObservancePublic,
		Month:    time.June,
		Day:      19,
		Observed: weekendAlt,
		Func:     cal.CalcDayOfMonth,
	}
	FridayAfterThanksGiving := &cal.Holiday{
		Name:     "FridayAfterThanksGiving",
		Type:     cal.ObservancePublic,
		Month:    time.November,
		Day:      27,
		Observed: weekendAlt,
		Func:     cal.CalcDayOfMonth,
	}

	NewYearsEve := &cal.Holiday{
		Name:     "NewYearsEve",
		Type:     cal.ObservancePublic,
		Month:    time.December,
		Day:      31,
		Observed: weekendAlt,
		Func:     cal.CalcDayOfMonth,
	}

	c.AddHoliday(
		us.NewYear,
		us.MlkDay,
		us.PresidentsDay,
		us.MemorialDay,
		Juneteenth,
		us.IndependenceDay,
		us.LaborDay,
		us.ThanksgivingDay,
		FridayAfterThanksGiving,
		us.ChristmasDay,
		NewYearsEve,
	)
	log := logger.New(os.Stderr, "", 0)

	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	//replace with what you actually have
	jiraUserID := user.Username + "@khanacademy.org"
	jiraPassword := os.ExpandEnv("${JIRA_API_TOKEN}")
	//jiraBaseURL := "https://khanacademy.atlassian.net"
	jiraBaseURL := "http://127.0.0.1:9000"

	var ticket string
	//I don't need the program name
	args := os.Args[1:]

	if len(args) == 0 {
		log.Println("Usage: wti <ticket>")
		panic(errors.New("Usage: wti <ticket>"))
	} else {
		ticket = args[0]
	}
	jiraIssueURI := fmt.Sprintf("/rest/api/2/issue/%s", ticket)

	header := make(http.Header)
	header.Set("Accept", "application/json")
	header.Set("Content-Type", "application/json")
	header.Set("Authorization", "Basic "+basicAuth(jiraUserID, jiraPassword))
	//rt := auth.NewRoundTripper(http.DefaultTransport, header)
	rt := auth.NewHeaderRoundTripper(http.DefaultTransport, header)

		//rt.BasicAuth(jiraUserID, jiraPassword)

	//jiraClient := &http.Client{Transport: rt}

	jiraClient := pester.New()
	//jiraClient.EmbedHTTPClient(authClient)
	////jiraClient := pester.NewExtendedClient(authClient)
	//
	jiraClient.Concurrency = 3
	jiraClient.MaxRetries = 5
	jiraClient.Backoff = pester.ExponentialBackoff
	jiraClient.LogHook = func(e pester.ErrEntry) { log.Println(jiraClient.FormatError(e)) }
	jiraClient.Transport = rt


	issue, _ := getJiraTicket(jiraClient, header, jiraBaseURL, jiraIssueURI)
	fmt.Println(jiraClient.LogString())
	changelog := Changelog{
		Values:     nil,
	}
	isLast := false
	startAt := 0
	for !isLast {
		jiraChangelogURI := fmt.Sprintf( "%s/changelog?maxResults=100&startAt=%d&maxResults=1", jiraIssueURI, startAt)
		currentChanges, changeErr := getJiraChangeLog(jiraClient, header, jiraBaseURL, jiraChangelogURI)
		fmt.Println(jiraClient.LogString())
		if changeErr != nil {
			fmt.Println(changeErr)
		}
		changelog.Values = append(changelog.Values, currentChanges.Values...)
		isLast = currentChanges.IsLast
		startAt = currentChanges.StartAt+currentChanges.MaxResults
	}



	var statusChanges []StatusChange

	fmt.Println(issue.Key)

	createdTime, createdErr := JIRATime(issue.JIRAFields.Created)
	if createdErr != nil {
		fmt.Println(createdErr)
	}
	createdChange := StatusChange{
		FromStatus: "Ex Nihilo",
		ToStatus:   "To Do",
		ChangeTime: createdTime,
	}
	statusChanges = append(statusChanges, createdChange)

	statusChanges = append(statusChanges,getStatusChanges(changelog)...)

	sort.Slice(statusChanges, func(i, j int) bool {
		return statusChanges[i].ChangeTime.Before(statusChanges[j].ChangeTime)
	})

	created := statusChanges[0]
	inProgress := statusChanges[1]
	lastChange := statusChanges[len(statusChanges)-1]

	for i := range statusChanges {
		sc := statusChanges[i]
		log.Println("STATUS CHANGE From:", sc.FromStatus, "To:", sc.ToStatus, "On:", sc.ChangeTime)
		// last transition to In Progress
		if sc.ToStatus == "In Progress" {
			inProgress = sc
		}

		// last transition to Done
		if sc.ToStatus == "Done" {
			lastChange = sc
		}
	}

	//waitTime := inProgress.ChangeTime.Sub(created.ChangeTime)
	//leadTime := lastChange.ChangeTime.Sub(created.ChangeTime)
	//cycleTime := lastChange.ChangeTime.Sub(inProgress.ChangeTime)

	waitTime := c.WorkdaysInRange(created.ChangeTime, inProgress.ChangeTime)
	leadTime := c.WorkdaysInRange(created.ChangeTime, lastChange.ChangeTime)
	cycleTime := c.WorkHoursInRange(inProgress.ChangeTime, lastChange.ChangeTime)
	fmt.Println("Wait Days:", waitTime, "Lead Days:", leadTime, "Cycle Hours:", cycleTime)
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

type withHeader struct {
	http.Header
	rt http.RoundTripper
}

func WithHeader(rt http.RoundTripper) withHeader {
	if rt == nil {
		rt = http.DefaultTransport
	}

	return withHeader{Header: make(http.Header), rt: rt}
}

func (h withHeader) RoundTrip(req *http.Request) (*http.Response, error) {
	//for k, v := range h.Header {
	//	req.Header[k] = v
	//}
	//
	//return h.rt.RoundTrip(req)
	// Clone the request
	rClone := new(http.Request)
	*rClone = *req
	rClone.Header = make(http.Header, len(req.Header))
	for idx, header := range req.Header {
		rClone.Header[idx] = append([]string(nil), header...)
	}
	for k, v := range h.Header {
		rClone.Header[k] = v
	}
	fmt.Println("Making a request")
	return h.rt.RoundTrip(rClone)
}

func getStatusChanges(changelog Changelog) []StatusChange {
	var statusChanges []StatusChange
	for j := range changelog.Values {
		history := changelog.Values[j]
		for k := range history.Items {
			item := history.Items[k]
			if item.Field == "status" && item.FieldID == "status" && item.FieldType == "jira" {
				changeTime, timeErr := JIRATime(history.Created)
				if timeErr != nil {
					fmt.Println(timeErr)
				}
				statusChange := StatusChange{
					FromStatus: item.FromString,
					ToStatus:   item.ToString,
					ChangeTime: changeTime,
				}
				statusChanges = append(statusChanges, statusChange)
			}
		}
	}
	return statusChanges
}

// JIRATime will transform the Jira time into a time.Time
func JIRATime(s string) (time.Time, error) {
	// Ignore null, like in the main JSON package.
	if s == "null" {
		fmt.Println("null time")
		return time.Time{}, nil
	}

	return time.Parse("2006-01-02T15:04:05.999-0700", s)
}

type StatusChange struct {
	FromStatus string
	ToStatus   string
	ChangeTime time.Time
}

/*
   {
     "field": "status",
     "fieldtype": "jira",
     "fieldId": "status",
     "from": "10149",
     "fromString": "Landed",
     "to": "10001",
     "toString": "Done"
   }
*/

// This is a "glue" function.  It takes all of the more testable, behavioral
// functions and "glues" them together without any other inherent behavior
func getJiraTicket(jiraClient httpdoer.HttpRequestDoer, header http.Header, jiraBaseURL, jiraAPIURI string) (JIRAIssue, error) {
	url := jiraBaseURL + jiraAPIURI
	req := BuildRequest(url)
	response, respErr := jiraClient.Do(req)
	if respErr != nil {
		fmt.Println(respErr)
	}
	body := GetBody(response)
	return ParseJiraIssue(body)
}


func getJiraChangeLog(jiraClient httpdoer.HttpRequestDoer, header http.Header, jiraBaseURL, jiraAPIURI string) (Changelog, error){
	url := jiraBaseURL + jiraAPIURI
	req := BuildRequest(url)
	response, respErr := jiraClient.Do(req)
	if respErr != nil {
		fmt.Println(respErr)
	}
	body := GetBody(response)
	return ParseJiraChangelog(body)
}

// BuildRequest will build a new client and request with the proper
// headers, including basic authentication
func BuildRequest(url string) *http.Request {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println(err)
	}
	return req
}

// GetBody will take an httpResponse and extract the body as a string
func GetBody(res *http.Response) string {
	defer checkedResponseBodyClose(res)
	body, _ := ioutil.ReadAll(res.Body)
	return string(body)
}

func checkedResponseBodyClose(response *http.Response) {
	if response != nil && response.Body != nil {
		err := response.Body.Close()
		if err != nil {
			logger.Println(err)
		}
	}
}

// ParseJiraIssue will parse the Jira response into a JIRAIssue
func ParseJiraIssue(jsonData string) (JIRAIssue, error) {
	issue := JIRAIssue{}

	jsonErr := json.Unmarshal([]byte(jsonData), &issue)
	if jsonErr != nil {
		logger.Println(jsonErr)
		return JIRAIssue{}, jsonErr
	}

	return issue, nil
}

// ParseJiraIssue will parse the Jira response into a JIRAIssue
func ParseJiraChangelog(jsonData string) (Changelog, error) {
	changelog := Changelog{}

	jsonErr := json.Unmarshal([]byte(jsonData), &changelog)
	if jsonErr != nil {
		logger.Println(jsonErr)
		return Changelog{}, jsonErr
	}

	return changelog, nil
}

type JIRAIssue struct {
	Expand     string     `json:"expand"`
	ID         string     `json:"id"`
	Self       string     `json:"self"`
	Key        string     `json:"key"`
	JIRAFields JIRAFields `json:"fields"`
}
type JIRAFields struct {
	Description string `json:"description"`
	Summary     string `json:"summary"`
	Created     string `json:"created"`
}

type ChangelogItems struct {
	Field      string `json:"field"`
	FieldType  string `json:"fieldtype"`
	FieldID    string `json:"fieldId"`
	To         string `json:"to,omitempty"`
	ToString   string `json:"toString,omitempty"`
	From       string `json:"from,omitempty"`
	FromString string `json:"fromString,omitempty"`
}

type Changelog struct {
	MaxResults int      `json:"maxResults"`
	StartAt    int      `json:"startAt"`
	Total      int      `json:"total"`
	IsLast     bool     `json:"isLast"`
	Values     []Values `json:"values"`
}

type Values struct {
	Created string  `json:"created"`
	Items   []ChangelogItems `json:"items"`
}