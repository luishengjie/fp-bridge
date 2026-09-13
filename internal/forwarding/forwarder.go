package forwarding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/luishengjie/fp-bridge/internal/event"
)

type Forwarder interface {
	Forward(ctx context.Context, linkedEvent event.Event) error
}

type HTTPForwarder struct {
	endpoint *url.URL
	client   *http.Client
}

func NewHTTPForwarder(endpoint *url.URL, client *http.Client) *HTTPForwarder {
	return &HTTPForwarder{
		endpoint: endpoint,
		client:   client,
	}
}

func (f *HTTPForwarder) Forward(ctx context.Context, linkedEvent event.Event) error {
	body, err := json.Marshal(linkedEvent)
	if err != nil {
		return fmt.Errorf("encode event: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		f.endpoint.String(),
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("create forwarding request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := f.client.Do(request)
	if err != nil {
		return fmt.Errorf("send event: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, response.Body)
		return fmt.Errorf("backend returned status %d", response.StatusCode)
	}

	_, _ = io.Copy(io.Discard, response.Body)
	return nil
}
