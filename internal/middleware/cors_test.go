package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const allowedOrigin = "http://localhost:5173"

func testCORSHandler(called *bool) http.Handler {
	return NewCORS(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			*called = true
			w.WriteHeader(http.StatusOK)
		}),
		[]string{allowedOrigin},
	)
}

func TestCORSHandlesAllowedPreflight(t *testing.T) {
	called := false
	handler := testCORSHandler(&called)
	request := httptest.NewRequest(http.MethodOptions, "/v1/events", nil)
	request.Header.Set("Origin", allowedOrigin)
	request.Header.Set("Access-Control-Request-Method", http.MethodPost)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if called {
		t.Error("next handler was called for preflight")
	}
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != allowedOrigin {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, allowedOrigin)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Methods"); got != "POST, OPTIONS" {
		t.Errorf("Access-Control-Allow-Methods = %q, want %q", got, "POST, OPTIONS")
	}
}

func TestCORSAllowsConfiguredOrigin(t *testing.T) {
	called := false
	handler := testCORSHandler(&called)
	request := httptest.NewRequest(http.MethodPost, "/v1/events", nil)
	request.Header.Set("Origin", allowedOrigin)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if !called {
		t.Error("next handler was not called")
	}
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != allowedOrigin {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, allowedOrigin)
	}
}

func TestCORSRejectsUnconfiguredOrigin(t *testing.T) {
	called := false
	handler := testCORSHandler(&called)
	request := httptest.NewRequest(http.MethodPost, "/v1/events", nil)
	request.Header.Set("Origin", "https://example.invalid")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
	if called {
		t.Error("next handler was called for an unconfigured origin")
	}
}

func TestCORSAllowsRequestWithoutOrigin(t *testing.T) {
	called := false
	handler := testCORSHandler(&called)
	request := httptest.NewRequest(http.MethodPost, "/v1/events", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if !called {
		t.Error("next handler was not called")
	}
}

func TestCORSDoesNotAffectOtherPaths(t *testing.T) {
	called := false
	handler := testCORSHandler(&called)
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	request.Header.Set("Origin", "https://example.invalid")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if !called {
		t.Error("next handler was not called")
	}
}
