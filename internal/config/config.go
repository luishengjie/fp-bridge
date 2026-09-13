package config

import (
	"fmt"
	"net/url"
	"os"
)

type Config struct {
	ListenAddress  string
	Upstream       *url.URL
	EventEndpoint  *url.URL
	TLSCertificate string
	TLSPrivateKey  string
}

func Load() (Config, error) {
	tlsCertificate := os.Getenv("FPBRIDGE_TLS_CERT")
	if tlsCertificate == "" {
		return Config{}, fmt.Errorf(
			"FPBRIDGE_TLS_CERT is required",
		)
	}
	tlsPrivateKey := os.Getenv("FPBRIDGE_TLS_KEY")
	if tlsPrivateKey == "" {
		return Config{}, fmt.Errorf(
			"FPBRIDGE_TLS_KEY is required",
		)
	}

	listenAddress := os.Getenv("FPBRIDGE_LISTEN_ADDRESS")
	if listenAddress == "" {
		listenAddress = ":8443"
	}

	upstreamValue := os.Getenv("FPBRIDGE_UPSTREAM")
	if upstreamValue == "" {
		upstreamValue = "http://127.0.0.1:9000"
	}

	upstream, err := parseHTTPURL("FPBRIDGE_UPSTREAM", upstreamValue)
	if err != nil {
		return Config{}, err
	}

	eventEndpointValue := os.Getenv("FPBRIDGE_EVENT_ENDPOINT")
	if eventEndpointValue == "" {
		return Config{}, fmt.Errorf("FPBRIDGE_EVENT_ENDPOINT is required")
	}

	eventEndpoint, err := parseHTTPURL(
		"FPBRIDGE_EVENT_ENDPOINT",
		eventEndpointValue,
	)
	if err != nil {
		return Config{}, err
	}

	return Config{
		ListenAddress:  listenAddress,
		Upstream:       upstream,
		EventEndpoint:  eventEndpoint,
		TLSCertificate: tlsCertificate,
		TLSPrivateKey:  tlsPrivateKey,
	}, nil
}

func parseHTTPURL(name string, value string) (*url.URL, error) {
	parsed, err := url.Parse(value)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", name, err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("%s scheme must be http or https", name)
	}

	if parsed.Host == "" {
		return nil, fmt.Errorf("%s must include a host", name)
	}

	return parsed, nil
}
