package proxy

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestReverseProxyForwardsRequest(t *testing.T) {
	upstream := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/hello" {
				t.Errorf("path = %q, want /hello", r.URL.Path)
			}

			if r.URL.Query().Get("name") != "fpbridge" {
				t.Errorf(
					"name = %q, want fpbridge",
					r.URL.Query().Get("name"),
				)
			}

			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "response from upstream")
		}),
	)
	defer upstream.Close()

	upstreamURL, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatalf("parse upstream URL: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"http://fpbridge.test/hello?name=fpbridge",
		nil,
	)

	recorder := httptest.NewRecorder()

	New(upstreamURL).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"status = %d, want %d",
			recorder.Code,
			http.StatusOK,
		)
	}

	if body := recorder.Body.String(); body != "response from upstream" {
		t.Fatalf(
			"body = %q, want %q",
			body,
			"response from upstream",
		)
	}
}

func TestRemoveProtectedHeaders(t *testing.T) {
	header := http.Header{}

	header.Set("X-FPBridge-JA4", "attacker-value")
	header.Set("X-Forwarded-For", "198.51.100.10")
	header.Set("X-Normal-Header", "keep-me")

	removeProtectedHeaders(header)

	if value := header.Get("X-FPBridge-JA4"); value != "" {
		t.Errorf("X-FPBridge-JA4 survived with value %q", value)
	}

	if value := header.Get("X-Forwarded-For"); value != "" {
		t.Errorf("X-Forwarded-For survived with value %q", value)
	}

	if value := header.Get("X-Normal-Header"); value != "keep-me" {
		t.Errorf(
			"X-Normal-Header = %q, want keep-me",
			value,
		)
	}
}

func TestReverseProxyRemovesClientFingerprintHeader(t *testing.T) {
	var receivedJA4 string

	upstream := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedJA4 = r.Header.Get("X-FPBridge-JA4")
			w.WriteHeader(http.StatusNoContent)
		}),
	)
	defer upstream.Close()

	upstreamURL, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatalf("parse upstream URL: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"http://fpbridge.test/",
		nil,
	)
	request.Header.Set("X-FPBridge-JA4", "attacker-value")

	recorder := httptest.NewRecorder()
	New(upstreamURL).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf(
			"status = %d, want %d",
			recorder.Code,
			http.StatusNoContent,
		)
	}

	if receivedJA4 != "" {
		t.Fatalf(
			"upstream received client-supplied JA4 %q",
			receivedJA4,
		)
	}
}
