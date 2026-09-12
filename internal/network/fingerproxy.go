package network

import (
	"crypto/tls"
	"fmt"
	"math"
	"net/http"

	fpfingerprint "github.com/wi1dcard/fingerproxy/pkg/fingerprint"
	fpmetadata "github.com/wi1dcard/fingerproxy/pkg/metadata"
)

type FingerproxyProvider struct {
	http2Fingerprinter *fpfingerprint.HTTP2FingerprintParam
}

func NewFingerproxyProvider() *FingerproxyProvider {
	return &FingerproxyProvider{
		http2Fingerprinter: &fpfingerprint.HTTP2FingerprintParam{
			MaxPriorityFrames: math.MaxUint,
		},
	}
}

func (provider *FingerproxyProvider) FromRequest(
	request *http.Request,
) (Metadata, error) {
	captured, ok := fpmetadata.FromContext(request.Context())
	if !ok {
		return Metadata{}, fmt.Errorf(
			"fingerproxy metadata is missing from request context",
		)
	}

	ja3, err := fpfingerprint.JA3Fingerprint(captured)
	if err != nil {
		return Metadata{}, fmt.Errorf("calculate JA3: %w", err)
	}

	ja4, err := fpfingerprint.JA4Fingerprint(captured)
	if err != nil {
		return Metadata{}, fmt.Errorf("calculate JA4: %w", err)
	}

	http2Fingerprint, err :=
		provider.http2Fingerprinter.HTTP2Fingerprint(captured)
	if err != nil {
		return Metadata{}, fmt.Errorf(
			"calculate HTTP/2 fingerprint: %w",
			err,
		)
	}

	return Metadata{
		Source:           "fingerproxy",
		Transport:        "tcp",
		TLSVersion:       tls.VersionName(captured.ConnectionState.Version),
		ServerName:       captured.ConnectionState.ServerName,
		NegotiatedALPN:   captured.ConnectionState.NegotiatedProtocol,
		JA3:              ja3,
		JA4:              ja4,
		HTTPVersion:      request.Proto,
		HTTP2Fingerprint: http2Fingerprint,
	}, nil
}
