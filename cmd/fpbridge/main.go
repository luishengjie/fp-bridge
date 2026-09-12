// Main Proxy
// Starts and configures the application

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/luishengjie/fp-bridge/internal/config"
	"github.com/luishengjie/fp-bridge/internal/proxy"
)

type healthResponse struct {
	Status string `json:"status"`
}

// Endpoint that checks if FP Bridge running and able to handle HTTP request
func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(
			w,
			http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json") // Sets the response type
	w.WriteHeader(http.StatusOK)                       // Sets HTTP status 200

	response := healthResponse{ // Response Data
		Status: "ok",
	}

	// Error handler
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("encode health response: %v", err)
	}
}

// Create the Router
func newHandler(upstream *url.URL) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler) // Route registered for GET request
	mux.Handle("/", proxy.New(upstream))

	return mux
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load configuration: %v", err)

	}

	// Returns a pointer to the new http.Server
	server := &http.Server{
		Addr:              cfg.ListenAddress,
		Handler:           newHandler(cfg.Upstream),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf(
		"FPBridge listening on %s and forwarding to %s",
		server.Addr,
		cfg.Upstream,
	)

	if err := server.ListenAndServeTLS(
		cfg.TLSCertificate,
		cfg.TLSPrivateKey,
	); err != nil {
		log.Fatalf("serve HTTPS: %v", err)
	}

}
