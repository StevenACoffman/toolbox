package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// thanks to peter bourgon on Gophers slack

func main() {
	// replace with what you actually have
	jiraUserID := os.ExpandEnv("${JIRA_USERID}")
	jiraPassword := os.ExpandEnv("${JIRA_API_TOKEN}")
	jiraBaseURL := "https://example.atlassian.net/rest/agile/1.0/board/"

	var rt http.RoundTripper
	rt = http.DefaultTransport
	rt = NewDelayRoundTripper(rt, 123*time.Millisecond)
	rt = NewLoggingRoundTripper(rt, os.Stderr)
	header := make(http.Header)
	header.Set("Accept", "application/json")
	header.Set("Content-Type", "application/json")
	hrt := NewHeaderRoundTripper(rt, header)
	hrt.BasicAuth(jiraUserID, jiraPassword)

	client := &http.Client{
		Transport: hrt,
	}

	_, err := client.Get(jiraBaseURL)
	fmt.Printf("err=%v\n", err)
}

//
//
//

type DelayRoundTripper struct {
	next  http.RoundTripper
	delay time.Duration
}

func NewDelayRoundTripper(
	next http.RoundTripper,
	delay time.Duration,
) *DelayRoundTripper {
	return &DelayRoundTripper{
		next:  next,
		delay: delay,
	}
}

func (rt *DelayRoundTripper) RoundTrip(
	req *http.Request,
) (resp *http.Response, err error) {
	time.Sleep(rt.delay)
	return rt.next.RoundTrip(req)
}

//
//
//

type LoggingRoundTripper struct {
	next http.RoundTripper
	dst  io.Writer
}

func NewLoggingRoundTripper(
	next http.RoundTripper,
	dst io.Writer,
) *LoggingRoundTripper {
	return &LoggingRoundTripper{
		next: next,
		dst:  dst,
	}
}

func (rt *LoggingRoundTripper) RoundTrip(
	req *http.Request,
) (resp *http.Response, err error) {
	defer func(begin time.Time) {
		fmt.Fprintf(rt.dst,
			"method=%s host=%s status_code=%d err=%v took=%s\n",
			req.Method, req.URL.Host, resp.StatusCode, err, time.Since(begin),
		)
	}(time.Now())

	return rt.next.RoundTrip(req)
}

//
//
//

type HeaderRoundTripper struct {
	next   http.RoundTripper
	Header http.Header
}

func NewHeaderRoundTripper(
	next http.RoundTripper,
	Header http.Header,
) *HeaderRoundTripper {
	if next == nil {
		next = http.DefaultTransport
	}
	return &HeaderRoundTripper{
		next:   next,
		Header: Header,
	}
}

func (rt *HeaderRoundTripper) RoundTrip(
	req *http.Request,
) (resp *http.Response, err error) {
	if rt.Header != nil {
		for k, v := range rt.Header {
			req.Header[k] = v
		}
	}
	fmt.Println("HeaderRoundTrip")
	return rt.next.RoundTrip(req)
}

func (rt *HeaderRoundTripper) BasicAuth(username, password string) {
	if rt.Header == nil {
		rt.Header = make(http.Header)
	}

	auth := username + ":" + password
	base64Auth := base64.StdEncoding.EncodeToString([]byte(auth))
	rt.Header.Set("Authorization", "Basic "+base64Auth)
}

func (rt *HeaderRoundTripper) SetHeader(key, value string) {
	if rt.Header == nil {
		rt.Header = make(http.Header)
	}
	rt.Header.Set(key, value)
}
