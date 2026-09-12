package config

import (
	"fmt"
	"net/url"
	"os"
)

type Config struct {
	ListenAddress  string
	Upstream       *url.URL
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

	upstream, err := url.Parse(upstreamValue)
	if err != nil {
		return Config{}, fmt.Errorf(
			"parse FPBRIDGE_UPSTREAM %w", err,
		)
	}

	if upstream.Scheme != "http" && upstream.Scheme != "https" {
		return Config{}, fmt.Errorf("FPBRIDGE_UPSTREAM scheme must be http or https")
	}

	if upstream.Host == "" {
		return Config{}, fmt.Errorf("FPBRIDGE_UPSTREAM must include a host")
	}

	return Config{
		ListenAddress:  listenAddress,
		Upstream:       upstream,
		TLSCertificate: tlsCertificate,
		TLSPrivateKey:  tlsPrivateKey,
	}, nil

}
