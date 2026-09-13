package ingest

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/luishengjie/fp-bridge/internal/event"
	"github.com/luishengjie/fp-bridge/internal/forwarding"
	"github.com/luishengjie/fp-bridge/internal/network"
)

type response struct {
	Accepted bool   `json:"accepted"`
	EventID  string `json:"event_id"`
}

type Handler struct {
	networkProvider network.Provider // reads network metadata from a HTTP request
	forwarder       forwarding.Forwarder
}

func NewHandler(networkProvider network.Provider, forwarder forwarding.Forwarder) *Handler {
	// Creates a handler and gives it its network provider
	return &Handler{
		networkProvider: networkProvider,
		forwarder:       forwarder,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handling HTTP request
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var submission event.BrowserSubmission

	if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// retrieve network metadata
	networkMetadata, err := h.networkProvider.FromRequest(r)
	if err != nil {
		http.Error(w, "network metadata unavailable", http.StatusInternalServerError)
		return
	}

	eventID, err := event.NewID()
	if err != nil {
		http.Error(w, "create event ID", http.StatusInternalServerError)
		return
	}

	linkedEvent := event.Event{
		SchemaVersion: "1.0",
		EventID:       eventID,
		ObservedAt:    time.Now().UTC(),
		Request: event.Request{
			Method: r.Method,
			Host:   r.Host,
			Path:   r.URL.Path,
		},
		Network: networkMetadata,
		Browser: event.BrowserObservation{
			Trust:       "client_reported",
			CollectedAt: submission.CollectedAt,
			Collector:   submission.Collector,
			Payload:     submission.Payload,
		},
	}

	if err := h.forwarder.Forward(r.Context(), linkedEvent); err != nil {
		log.Printf(
			"event forwarding failed: event_id=%s error=%v",
			linkedEvent.EventID,
			err,
		)

		http.Error(
			w,
			"forward event",
			http.StatusBadGateway, // 502 Bad Gateway: FP-Bridge received valid request but dev backend failed to accept it
		)
		return
	}

	log.Printf("event forwarded: event_id=%s", linkedEvent.EventID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response{
		Accepted: true,
		EventID:  linkedEvent.EventID,
	}); err != nil {
		return
	}
}
