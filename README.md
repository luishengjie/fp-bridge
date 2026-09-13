# FP-Bridge

FP-Bridge is a reverse proxy that links network fingerprints with browser fingerprints at the request level. 

It allows developers to use their browser fingerprint SDK of choice and forward the combined telemetry back to their backend.

## How it works
```text
Browser
   │ HTTPS POST /v1/events
   ▼
FPBridge
   ├── reads browser signals from the request body
   ├── observes TLS and HTTP fingerprints
   ├── creates a linked event
   └── forwards it to the event endpoint
```
Normal website requests are forwarded to `FPBRIDGE_UPSTREAM`. Linked fingerprint events are sent to `FPBRIDGE_EVENT_ENDPOINT`.

## Implementation

FP-Bridge uses [fingerproxy](https://github.com/wi1dcard/fingerproxy), licensed
under Apache License 2.0, to capture TLS and HTTP/2 metadata and calculate JA3,
JA4, and HTTP/2 fingerprints.

FPBridge links browser and network signals, normalizes them into a common event format, and forwards the combined event to the developer's backend.



## Requirements
- Go 1.26 or later
- A TLS certificate and private key


## Quick start

Start example event backend:

```bash
python3 examples/backend-python/server.py
```


Start FP-Bridge:

```bash
FPBRIDGE_LISTEN_ADDRESS=:8443 \
FPBRIDGE_UPSTREAM=http://127.0.0.1:9000 \
FPBRIDGE_EVENT_ENDPOINT=http://127.0.0.1:9100/events \
FPBRIDGE_TLS_CERT=certs/localhost-cert.pem \
FPBRIDGE_TLS_KEY=certs/localhost-key.pem \
go run ./cmd/fpbridge
```

| Variable | Purpose |
|---|---|
| `FPBRIDGE_LISTEN_ADDRESS` | HTTPS address FPBridge listens on |
| `FPBRIDGE_UPSTREAM` | Normal website/application endpoint |
| `FPBRIDGE_EVENT_ENDPOINT` | Backend endpoint that receives linked events |
| `FPBRIDGE_TLS_CERT` | TLS certificate file |
| `FPBRIDGE_TLS_KEY` | TLS private-key file |


Send a test event:

```bash
curl -k -i https://localhost:8443/v1/events \
  -H 'Content-Type: application/json' \
  -d '{
    "schema_version": "1.0",
    "collected_at": "2026-09-13T10:00:00Z",
    "collector": {
      "name": "dummy-sdk",
      "version": "0.1.0"
    },
    "payload": {
      "fingerprint": "browser-test-123",
      "bot": false
    }
  }'
```

Expected response:

```json
{
  "accepted": true,
  "event_id": "evt_..."
}
```

## Testing

```bash
go test ./...
go vet ./...
```

## License
Apache-2.0