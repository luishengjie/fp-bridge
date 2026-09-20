package forwarding

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/luishengjie/fp-bridge/internal/event"
	"github.com/luishengjie/fp-bridge/internal/network"
)

func TestHTTPForwarderSendsEvent(t *testing.T) {
	var received event.Event

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", contentType)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer backend.Close()

	endpoint, err := url.Parse(backend.URL)
	if err != nil {
		t.Fatalf("parse backend URL: %v", err)
	}

	forwarder := NewHTTPForwarder(endpoint, backend.Client())
	linkedEvent := event.Event{
		SchemaVersion: "1.0",
		EventID:       "evt_test123",
		ObservedAt:    time.Date(2026, 9, 13, 10, 0, 0, 0, time.UTC),
		Network: network.Metadata{
			JA4: "test-ja4",
		},
		Browser: event.BrowserObservation{
			Payload: json.RawMessage(`{"fingerprint":"browser-123"}`),
		},
	}

	result, err := forwarder.Forward(
		t.Context(),
		linkedEvent,
	)
	if err != nil {
		t.Fatalf("Forward() error = %v", err)
	}

	if result.Payload != nil {
		t.Errorf(
			"Payload = %s, want nil",
			result.Payload,
		)
	}

	if received.EventID != "evt_test123" {
		t.Errorf("received event ID = %q, want %q", received.EventID, "evt_test123")
	}
	if received.Network.JA4 != "test-ja4" {
		t.Errorf("received JA4 = %q, want %q", received.Network.JA4, "test-ja4")
	}
}

func TestHTTPForwarderRejectsNon2xxResponse(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "backend unavailable", http.StatusServiceUnavailable)
	}))
	defer backend.Close()

	endpoint, err := url.Parse(backend.URL)
	if err != nil {
		t.Fatalf("parse backend URL: %v", err)
	}

	forwarder := NewHTTPForwarder(endpoint, backend.Client())
	_, err = forwarder.Forward(t.Context(), event.Event{EventID: "evt_test123"})
	if err == nil {
		t.Fatal("Forward() error = nil, want non-nil error")
	}
}

func TestHTTPForwarderReturnsArbitraryJSON(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			_, _ = w.Write([]byte(`{
				"risk": 0.87,
				"show_captcha": true,
				"provider": {"name": "example"}
			}`))
		},
	))
	defer backend.Close()

	endpoint, err := url.Parse(backend.URL)
	if err != nil {
		t.Fatalf("parse backend URL: %v", err)
	}

	forwarder := NewHTTPForwarder(
		endpoint,
		backend.Client(),
	)

	result, err := forwarder.Forward(
		t.Context(),
		event.Event{EventID: "evt_test123"},
	)
	if err != nil {
		t.Fatalf("Forward() error = %v", err)
	}

	var backendResult struct {
		Risk        float64 `json:"risk"`
		ShowCaptcha bool    `json:"show_captcha"`
	}
	if err := json.Unmarshal(result.Payload, &backendResult); err != nil {
		t.Fatalf("decode result payload: %v", err)
	}
	if backendResult.Risk != 0.87 || !backendResult.ShowCaptcha {
		t.Errorf("result = %#v, want arbitrary backend fields", backendResult)
	}
}

func TestHTTPForwarderRejectsInvalidJSON(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`not-json`))
		},
	))
	defer backend.Close()

	endpoint, err := url.Parse(backend.URL)
	if err != nil {
		t.Fatalf("parse backend URL: %v", err)
	}

	forwarder := NewHTTPForwarder(endpoint, backend.Client())
	_, err = forwarder.Forward(t.Context(), event.Event{})
	if err == nil {
		t.Fatal("Forward() error = nil, want invalid JSON error")
	}
}
