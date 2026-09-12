package network

import "net/http"

type Provider interface {
	FromRequest(*http.Request) (Metadata, error)
}
