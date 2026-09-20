package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func testEventEndpointURL(t *testing.T) *url.URL {
	t.Helper()

	endpoint, err := url.Parse("http://127.0.0.1:9100/events")
	if err != nil {
		t.Fatalf("parse test event endpoint: %v", err)
	}

	return endpoint
}

func testHandler(t *testing.T) http.Handler {
	t.Helper()

	return newHandler(
		testEventEndpointURL(t),
		&http.Client{},
	)
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
	testHandler(t).ServeHTTP(recorder, request)

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

	testHandler(t).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"status = %d, want %d",
			recorder.Code,
			http.StatusMethodNotAllowed,
		)
	}
}

func TestUnknownRouteReturnsNotFound(t *testing.T) {
	request := httptest.NewRequest(
		http.MethodGet,
		"/products",
		nil,
	)
	recorder := httptest.NewRecorder()

	testHandler(t).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"status = %d, want %d",
			recorder.Code,
			http.StatusNotFound,
		)
	}
}
