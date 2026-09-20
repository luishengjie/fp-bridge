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

const maxBackendResponseBytes = 64 * 1024

type Result struct {
	Payload json.RawMessage
}

type Forwarder interface {
	Forward(ctx context.Context, linkedEvent event.Event) (Result, error)
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

func (f *HTTPForwarder) Forward(
	ctx context.Context,
	linkedEvent event.Event,
) (Result, error) {
	body, err := json.Marshal(linkedEvent)
	if err != nil {
		return Result{}, fmt.Errorf(
			"encode event: %w",
			err,
		)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		f.endpoint.String(),
		bytes.NewReader(body),
	)
	if err != nil {
		return Result{}, fmt.Errorf(
			"create forwarding request: %w",
			err,
		)
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := f.client.Do(request)
	if err != nil {
		return Result{}, fmt.Errorf(
			"send event: %w",
			err,
		)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNoContent {
		return Result{}, nil
	}

	// Any response outside 200–299 is a failure.
	if response.StatusCode < http.StatusOK ||
		response.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, response.Body)

		return Result{}, fmt.Errorf(
			"backend returned status %d",
			response.StatusCode,
		)
	}

	responseBody, err := io.ReadAll(io.LimitReader(
		response.Body,
		maxBackendResponseBytes+1,
	))
	if err != nil {
		return Result{}, fmt.Errorf(
			"read backend response: %w",
			err,
		)
	}

	if len(responseBody) > maxBackendResponseBytes {
		return Result{}, fmt.Errorf("backend response exceeds 64 KiB")
	}

	if !json.Valid(responseBody) {
		return Result{}, fmt.Errorf("backend returned invalid JSON")
	}

	return Result{Payload: json.RawMessage(responseBody)}, nil
}
