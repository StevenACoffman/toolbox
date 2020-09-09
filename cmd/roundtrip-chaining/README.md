### Alternative Functional version

```go
// NewRoundTripper returns an http.RoundTripper that is tooled for use in the app

func NewRoundTripper(original http.RoundTripper) http.RoundTripper {
	if original == nil {
		original = http.DefaultTransport
	}

	return roundTripperFunc(func(request *http.Request) (*http.Response, error) {
		response, err := original.RoundTrip(request)
		return response, err
	})
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

```

