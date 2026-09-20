package event

import (
	"encoding/json"
	"time"

	"github.com/luishengjie/fp-bridge/internal/network"
)

type Collector struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type BrowserSubmission struct {
	// What the browser wrapper sends to FPBridge
	SchemaVersion string          `json:"schema_version"`
	CollectedAt   time.Time       `json:"collected_at"`
	Collector     Collector       `json:"collector"` // which browser library generated the payload
	Payload       json.RawMessage `json:"payload"`   // preserves provider-specific JSON
}

type BrowserObservation struct {
	Trust       string          `json:"trust"`
	CollectedAt time.Time       `json:"collected_at"`
	Collector   Collector       `json:"collector"`
	Payload     json.RawMessage `json:"payload"` // preserves provider-specific JSON
}

type Request struct {
	// HTTP request used to submit the fingerprint
	Method string `json:"method"`
	Host   string `json:"host"`
	Path   string `json:"path"`
}

type Event struct {
	// Final output created by FPBridge
	SchemaVersion string             `json:"schema_version"`
	EventID       string             `json:"event_id"`
	ObservedAt    time.Time          `json:"observed_at"`
	Request       Request            `json:"request"`
	Network       network.Metadata   `json:"network"`
	Browser       BrowserObservation `json:"browser"`
}
