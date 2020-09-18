package auth

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"time"
)

// NewRoundTripper returns an http.RoundTripper that is tooled for use in the
// app
func NewRoundTripper(
	next http.RoundTripper,
	header http.Header,
) http.RoundTripper {
	if next == nil {
		next = http.DefaultTransport
	}

	return roundTripperFunc(
		func(request *http.Request) (*http.Response, error) {
			rClone := new(http.Request)
			*rClone = *request
			rClone.Header = make(http.Header, len(request.Header))
			for idx, header := range request.Header {
				rClone.Header[idx] = append([]string(nil), header...)
			}

			for k, v := range header {
				rClone.Header[k] = v
			}
			fmt.Println("We are tripping")

			return next.RoundTrip(rClone)
		},
	)
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(
	r *http.Request,
) (*http.Response, error) {
	return f(r)
}

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
	defer func(begin time.Time) {
		if resp != nil {
			fmt.Printf(
				"method=%s host=%s path=%s status_code=%d err=%v took=%s\n",
				req.Method,
				req.URL.Host,
				req.URL.Path,
				resp.StatusCode,
				err,
				time.Since(begin),
			)
		} else {
			fmt.Printf(
				"method=%s host=%s path=%s status_code=nil err=%v took=%s\n",
				req.Method, req.URL.Host, req.URL.Path, err, time.Since(begin),
			)
		}
	}(time.Now())
	// Clone the request
	//rClone := new(http.Request)
	//*rClone = *req
	//rClone.Header = make(http.Header, len(req.Header))
	//for idx, header := range req.Header {
	//	rClone.Header[idx] = append([]string(nil), header...)
	//}
	//
	//for k, v := range rt.Header {
	//	rClone.Header[k] = v
	//}
	//fmt.Println("We are tripping")
	//return rt.next.RoundTrip(rClone)

	//
	if rt.Header != nil {
		for k, v := range rt.Header {
			req.Header[k] = v
		}
	}
	fmt.Println("We are tripping")
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

// APIAuthTransport is an http.RoundTripper that authenticates all requests
// using HTTP Basic Authentication using the provided identifier and token.
type APIAuthTransport struct {
	APIIdentifier string      // API Identifier
	APIToken      string      // API Token
	Header        http.Header // HTTP Headers

	// Transport is the underlying HTTP transport to use when making requests.
	// It will default to http.DefaultTransport if nil.
	Transport http.RoundTripper
}

// RoundTrip implements the RoundTripper interface.
func (t *APIAuthTransport) RoundTrip(
	req *http.Request,
) (*http.Response, error) {
	// Clone the request
	rClone := new(http.Request)
	*rClone = *req
	rClone.Header = make(http.Header, len(req.Header))
	for idx, header := range req.Header {
		rClone.Header[idx] = append([]string(nil), header...)
	}
	rClone.SetBasicAuth(t.APIIdentifier, t.APIToken)
	for k, v := range t.Header {
		rClone.Header[k] = v
	}
	fmt.Println("Making a request")
	return t.GetTransport().RoundTrip(rClone)
}

// Client returns an *http.Client that makes requests that are authenticated
// using HTTP Basic Authentication.
func (t *APIAuthTransport) Client() *http.Client {
	return &http.Client{
		Transport: t,
	}
}

func (t *APIAuthTransport) GetTransport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}
	return http.DefaultTransport
}
