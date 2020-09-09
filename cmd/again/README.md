### Again - example on how to use backoff and return tuple
Keep in mind that http client documentation states:

> Returns an error if there were too many redirects or if there was an HTTP protocol error. A non-2xx response doesn’t cause an error.

```go
if err == nil && resp.StatusCode < http.StatusInternalServerError && (resp.StatusCode != http.StatusTooManyRequests) {
//... do something	
}
```