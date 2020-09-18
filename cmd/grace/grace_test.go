package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"syscall"
	"testing"
)

func TestGrace(t *testing.T) {
	t.Run("Wait with func", func(t *testing.T) {
		var result int

		result = 1
		go func() {
			_ = runServer()
			result = 1
		}()
		syscall.Kill(os.Getpid(), syscall.SIGUSR1)

		if result != 1 {
			t.Error("Result is not equal 1")
		}
	})
}

func TestHealthCheckHandler(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HealthCheck)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	want := "text/plain"
	if contentType := rr.Header().Get("Content-Type"); contentType != want {
		t.Errorf("handler returned wrong status code: got %v want %v",
			contentType, want)
	}

	want = "0"
	if contentLength := rr.Header().Get("Content-Length"); contentLength != want {
		t.Errorf("handler returned wrong status code: got %v want %v",
			contentLength, want)
	}
}
