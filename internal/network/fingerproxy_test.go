package network

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFingerproxyProviderMissingMetadata(t *testing.T) {
	request := httptest.NewRequest(
		http.MethodGet,
		"https://example.test/",
		nil,
	)

	provider := NewFingerproxyProvider()

	_, err := provider.FromRequest(request)

	if err == nil {
		t.Fatal("expected error when metadata is missing")
	}

	if !strings.Contains(err.Error(), "metadata is missing") {
		t.Fatalf(
			"error = %q, want missing metadata error",
			err,
		)
	}
}
