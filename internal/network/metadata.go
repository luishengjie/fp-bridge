package network

type Metadata struct {
	Source           string `json:"source"`
	Transport        string `json:"transport"`
	TLSVersion       string `json:"tls_version"`
	ServerName       string `json:"server_name,omitempty"`
	NegotiatedALPN   string `json:"negotiated_alpn,omitempty"`
	JA3              string `json:"ja3,omitempty"`
	JA4              string `json:"ja4,omitempty"`
	HTTPVersion      string `json:"http_version"`
	HTTP2Fingerprint string `json:"http2_fingerprint,omitempty"`
}
