// Main Proxy
// Starts and configures the application

package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os/signal"
	"syscall"
	"time"

	"github.com/luishengjie/fp-bridge/internal/config"
	"github.com/luishengjie/fp-bridge/internal/forwarding"
	"github.com/luishengjie/fp-bridge/internal/ingest"
	"github.com/luishengjie/fp-bridge/internal/network"
	"github.com/luishengjie/fp-bridge/internal/proxy"
	"github.com/wi1dcard/fingerproxy/pkg/proxyserver"
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
func newHandler(
	upstream *url.URL,
	eventEndpoint *url.URL,
	httpClient *http.Client,
) http.Handler {
	mux := http.NewServeMux()
	networkProvider := network.NewFingerproxyProvider()
	eventForwarder := forwarding.NewHTTPForwarder(
		eventEndpoint,
		httpClient,
	)

	ingestHandler := ingest.NewHandler(
		networkProvider,
		eventForwarder,
	)

	mux.HandleFunc("/health", healthHandler) // Route registered for GET request

	mux.Handle(
		"/v1/events",
		ingestHandler,
	)

	mux.Handle("/", proxy.New(upstream))

	return mux
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load configuration: %v", err)

	}

	// Load certificate
	certificate, err := tls.LoadX509KeyPair(
		cfg.TLSCertificate,
		cfg.TLSPrivateKey,
	)
	if err != nil {
		log.Fatalf("load TLS certificate: %v", err)
	}

	// Create TLS config
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		MinVersion:   tls.VersionTLS12,
		MaxVersion:   tls.VersionTLS13,
		NextProtos:   []string{"h2", "http/1.1"},
	}

	// Shutdown context
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	// Create a new proxy server
	server := proxyserver.NewServer(
		ctx,
		newHandler(
			cfg.Upstream,
			cfg.EventEndpoint,
			&http.Client{Timeout: 5 * time.Second},
		),
		tlsConfig,
	)

	server.HTTPServer.ReadHeaderTimeout = 5 * time.Second
	server.HTTPServer.ReadTimeout = 10 * time.Second
	server.HTTPServer.WriteTimeout = 10 * time.Second
	server.HTTPServer.IdleTimeout = 60 * time.Second
	server.TLSHandshakeTimeout = 10 * time.Second

	log.Printf(
		"FPBridge listening on %s, proxying to %s, forwarding events to %s",
		cfg.ListenAddress,
		cfg.Upstream,
		cfg.EventEndpoint,
	)

	if err := server.ListenAndServe(cfg.ListenAddress); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("serve HTTPS: %v", err)
	}

}
