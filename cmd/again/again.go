package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"

	"github.com/cenkalti/backoff"
)

func main() {
	var responseCode int = 429
	http.HandleFunc(
		"/boom",
		func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(responseCode)
			responseCode = 200
			res.Write([]byte("Boom!"))
		},
	)
	go http.ListenAndServe(":9000", nil)

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 3 * time.Minute

	var (
		resp *http.Response
		err  error
	)
	body := func(response *http.Response) string {
		raw, RawErr := httputil.DumpResponse(response, true)

		if RawErr == nil {
			return string(raw)
		}
		return ""
	}
	retryable := func() error {
		resp, err = doSomething()

		if err == nil && resp != nil &&
			(resp.StatusCode == http.StatusTooManyRequests ||
				resp.StatusCode >= http.StatusInternalServerError) {
			respCode := resp.StatusCode
			err = BadHttpResponseCode{
				HttpResponseCode: respCode,
				Message: "(Intermittent) HTTP response code " + strconv.Itoa(
					respCode,
				) + "\n" + body(
					resp,
				),
			}
		} else {
			log.Printf("Hey it worked")
		}
		resp.Body.Close()
		return err
	}

	notify := func(err error, t time.Duration) {
		log.Printf("error: %v happened at time: %v", err, t)
	}

	err = backoff.RetryNotify(retryable, b, notify)
	if err != nil {
		log.Fatalf("error after retrying: %v", err)
	}
}

func doSomething() (*http.Response, error) {
	client := &http.Client{}
	return client.Get("http://127.0.0.1:9000/boom")
}

// Any non 2xx HTTP status code is considered a bad response code, and will
// result in a BadHttpResponseCode.
type BadHttpResponseCode struct {
	HttpResponseCode int
	Message          string
}

// Returns an error message for this bad HTTP response code
func (err BadHttpResponseCode) Error() string {
	return err.Message
}
