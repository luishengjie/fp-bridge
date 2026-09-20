package ingest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/luishengjie/fp-bridge/internal/event"
	"github.com/luishengjie/fp-bridge/internal/forwarding"
	"github.com/luishengjie/fp-bridge/internal/network"
)

type fakeNetworkProvider struct {
	metadata network.Metadata
	err      error
}

func (f fakeNetworkProvider) FromRequest(_ *http.Request) (network.Metadata, error) {
	return f.metadata, f.err
}

type fakeForwarder struct {
	received event.Event
	result   forwarding.Result
	err      error
	called   bool
}

func (f *fakeForwarder) Forward(
	_ context.Context,
	linkedEvent event.Event,
) (forwarding.Result, error) {
	f.called = true
	f.received = linkedEvent
	return f.result, f.err
}

func TestHandlerLinksAndForwardsEvent(t *testing.T) {
	forwarder := &fakeForwarder{}
	handler := NewHandler(
		fakeNetworkProvider{metadata: network.Metadata{
			Source:      "test",
			Transport:   "tcp",
			TLSVersion:  "TLS 1.3",
			JA4:         "test-ja4",
			HTTPVersion: "HTTP/2.0",
		}},
		forwarder,
	)

	body := []byte(`{
		"schema_version": "1.0",
		"collected_at": "2026-09-13T10:00:00Z",
		"collector": {"name": "fpscanner", "version": "1.0.0"},
		"payload": {"fingerprint": "browser-123"}
	}`)
	request := httptest.NewRequest(http.MethodPost, "https://example.com/v1/events", bytes.NewReader(body))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if !forwarder.called {
		t.Fatal("forwarder was not called")
	}
	if forwarder.received.EventID == "" {
		t.Error("forwarded event ID is empty")
	}
	if forwarder.received.Network.JA4 != "test-ja4" {
		t.Errorf("forwarded JA4 = %q, want %q", forwarder.received.Network.JA4, "test-ja4")
	}
	var payload struct {
		Fingerprint string `json:"fingerprint"`
	}
	if err := json.Unmarshal(forwarder.received.Browser.Payload, &payload); err != nil {
		t.Fatalf("decode forwarded browser payload: %v", err)
	}
	if payload.Fingerprint != "browser-123" {
		t.Errorf("forwarded fingerprint = %q, want %q", payload.Fingerprint, "browser-123")
	}
	if forwarder.received.Browser.Trust != "client_reported" {
		t.Errorf("browser trust = %q, want %q", forwarder.received.Browser.Trust, "client_reported")
	}

	var responseBody response
	if err := json.NewDecoder(recorder.Body).Decode(&responseBody); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if responseBody.EventID != forwarder.received.EventID {
		t.Errorf("response event ID = %q, want %q", responseBody.EventID, forwarder.received.EventID)
	}
}

func TestHandlerRejectsInvalidJSON(t *testing.T) {
	forwarder := &fakeForwarder{}
	handler := NewHandler(fakeNetworkProvider{}, forwarder)
	request := httptest.NewRequest(http.MethodPost, "https://example.com/v1/events", strings.NewReader(`{invalid`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if forwarder.called {
		t.Error("forwarder was called for invalid JSON")
	}
}

func TestHandlerRejectsGET(t *testing.T) {
	forwarder := &fakeForwarder{}
	handler := NewHandler(fakeNetworkProvider{}, forwarder)
	request := httptest.NewRequest(http.MethodGet, "https://example.com/v1/events", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusMethodNotAllowed)
	}
	if forwarder.called {
		t.Error("forwarder was called for GET request")
	}
}

func TestHandlerReturnsErrorWhenNetworkMetadataIsUnavailable(t *testing.T) {
	forwarder := &fakeForwarder{}
	handler := NewHandler(fakeNetworkProvider{err: errors.New("metadata missing")}, forwarder)
	request := httptest.NewRequest(
		http.MethodPost,
		"https://example.com/v1/events",
		strings.NewReader(`{"schema_version":"1.0","payload":{}}`),
	)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if forwarder.called {
		t.Error("forwarder was called without network metadata")
	}
}

func TestHandlerReturnsBadGatewayWhenForwardingFails(t *testing.T) {
	forwarder := &fakeForwarder{err: errors.New("backend unavailable")}
	handler := NewHandler(
		fakeNetworkProvider{metadata: network.Metadata{JA4: "test-ja4"}},
		forwarder,
	)
	request := httptest.NewRequest(
		http.MethodPost,
		"https://example.com/v1/events",
		strings.NewReader(`{"schema_version":"1.0","payload":{}}`),
	)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadGateway)
	}
	if !forwarder.called {
		t.Error("forwarder was not called")
	}
}

func TestHandlerReturnsArbitraryBackendResult(t *testing.T) {
	forwarder := &fakeForwarder{
		result: forwarding.Result{
			Payload: json.RawMessage(`{
				"risk": 0.87,
				"show_captcha": true
			}`),
		},
	}
	handler := NewHandler(
		fakeNetworkProvider{metadata: network.Metadata{JA4: "test-ja4"}},
		forwarder,
	)
	request := httptest.NewRequest(
		http.MethodPost,
		"https://example.com/v1/events",
		strings.NewReader(`{
			"schema_version":"1.0",
			"collector":{"name":"fpscanner","version":"1.0.8"},
			"payload":{}
		}`),
	)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var responseBody response
	if err := json.NewDecoder(recorder.Body).Decode(&responseBody); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if responseBody.Result == nil {
		t.Fatal("result = nil, want backend result")
	}
	var backendResult struct {
		Risk        float64 `json:"risk"`
		ShowCaptcha bool    `json:"show_captcha"`
	}
	if err := json.Unmarshal(responseBody.Result, &backendResult); err != nil {
		t.Fatalf("decode backend result: %v", err)
	}
	if backendResult.Risk != 0.87 || !backendResult.ShowCaptcha {
		t.Errorf("result = %#v, want arbitrary backend fields", backendResult)
	}
}
