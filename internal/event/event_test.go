package event

import (
	"encoding/json"
	"testing"
)

func TestDecodeBrowserSubmission(t *testing.T) {
	input := []byte(`{
		"schema_version": "1.0",
		"collected_at": "2026-09-13T10:00:00Z",
		"collector": {
			"name": "fpscanner",
			"version": "1.0.8"
		},
		"payload": {
			"fsid": "FS1_example"
		}
	}`)

	var submission BrowserSubmission

	if err := json.Unmarshal(input, &submission); err != nil {
		t.Fatalf("decode browser submission: %v", err)
	}

	if submission.SchemaVersion != "1.0" {
		t.Errorf(
			"SchemaVersion = %q, want %q",
			submission.SchemaVersion,
			"1.0",
		)
	}

	if submission.Collector.Name != "fpscanner" {
		t.Errorf(
			"collector name = %q, want %q",
			submission.Collector.Name,
			"fpscanner",
		)
	}

	if len(submission.Payload) == 0 {
		t.Error("payload is empty")
	}
}
