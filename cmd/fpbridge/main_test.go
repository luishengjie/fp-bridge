package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func testUpstreamURL(t *testing.T) *url.URL {
	t.Helper()

	upstream, err := url.Parse("http://127.0.0.1:9000")
	if err != nil {
		t.Fatalf("parse test upstream: %v", err)
	}

	return upstream
}

func TestHealth(t *testing.T) {
	// Creates a fake request
	request := httptest.NewRequest(
		http.MethodGet,
		"/health",
		nil,
	)
	// Creates a fake resposne writer
	recorder := httptest.NewRecorder()
	// Sends fake request through the router
	newHandler(testUpstreamURL(t)).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"status = %d, want %d",
			recorder.Code,
			http.StatusOK,
		)
	}

	var response healthResponse

	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Status != "ok" {
		t.Fatalf(
			"response status = %q, want %q",
			response.Status,
			"ok",
		)
	}

}

func TestHealthRejectsPost(t *testing.T) {
	request := httptest.NewRequest(
		http.MethodPost,
		"/health",
		nil,
	)

	recorder := httptest.NewRecorder()

	newHandler(testUpstreamURL(t)).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"status = %d, want %d",
			recorder.Code,
			http.StatusMethodNotAllowed,
		)
	}
}
