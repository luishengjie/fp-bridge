# FPBridge

FPBridge links browser and network fingerprints at the request level. 

Existing fingerprint libraries typically focus on either browser or network-level signals. However, real-world detection system benefit from combining both signals. For example, an attacker may spoof browser attributes to appear legitimate, while their network fingerprint reveals inconsistencies with the claimed browser identity. Cross-layer fingerprinting helps identify these mismatches and can improve the detection of spoofed or automated clients.

FPBridge addresses this gap by linking browser and network fingerprints at the request level. It provides a simple way to correlated signals from both layers and make them available to the detection backend. 

FPBridge consists of two key modules:
1. **FPBridge Browser**: an SDK wrapper that allows developers to use their browser fingerprint SDK of choice.
2. **FPBridge Proxy**: a proxy that observes network fingerprints, links them with the browser signals captured by **FPBridge Browser**, and forwards them to the developer backend.

## Supported browser fingerprinting SDKs

| SDK | Adapter | Status | Payload |
|---|---|---|---|
| [FPScanner](https://github.com/antoinevastel/fpscanner) | `FPScannerAdapter` | Supported | FPScanner fingerprint and bot-detection signals |
<!-- | [FingerprintJS](https://github.com/fingerprintjs/fingerprintjs) | NIL | Planned | NIL | -->
<!-- | [BotD](https://github.com/fingerprintjs/BotD) | NIL | Planned | NIL | -->

## Network fingerprinting
FPBridge uses [fingerproxy](https://github.com/wi1dcard/fingerproxy) to obtain JA3, JA4, and HTTP/2 fingerprints.

fingerproxy is licensed under Apache License 2.0.



## Quick Start
### Dependencies

- Go 1.26 or later
- Node.js and npm
- A TLS certificate and private key
- [UV](https://docs.astral.sh/uv/) for example backend


### Installation
Clone the repository and install the following dependencies:

```bash
git clone https://github.com/luishengjie/fp-bridge.git
```

```bash
cd fp-bridge

go mod download

cd browser
npm install
npm run build
```

Install the python dependencies for the example backend:

```bash
cd examples
uv sync
```

Install the dependencies for the example web application

```bash
cd examples/web
npm install
cp .env.example .env
```


### FPBridge Browser
FPBridge Browser wraps a supported browser fingerprinting SDK and submits its payload to FPBridge Proxy.

```ts
import {
  FPBridge,
  FPScannerAdapter,
} from "@fpbridge/browser";

interface DetectionResult {
  action: "allow" | "challenge" | "block";
  reason_codes: string[];
}

const bridge = new FPBridge({
  endpoint: import.meta.env.VITE_FPBRIDGE_ENDPOINT,
  collector: new FPScannerAdapter(),
});

const response = await bridge.collectAndSubmit<DetectionResult>();

console.log(response.event_id);
console.log(response.result);
```

### Local TLS Certificate
FPBridge terminates TLS to observe the client handshake. For local development, a trusted certificate can be generated with [mkcert](https://github.com/FiloSottile/mkcert).

```bash
mkdir -p certs

mkcert \
  -cert-file certs/localhost-cert.pem \
  -key-file certs/localhost-key.pem \
  localhost 127.0.0.1 ::1
```


### FPBridge Proxy
The web application sends browser signals to the FPBridge Proxy endpoint configured by `VITE_FPBRIDGE_ENDPOINT`. FPBridge Proxy links those signals with the TLS and HTTP metadata observed in the same request and forwards the linked event to the detection backend configured by `FPBRIDGE_EVENT_ENDPOINT`.

First, start the example detection backend:

```bash
cd examples
uv sync
uv run python backend/server.py
```

In another terminal, start FPBridge Proxy:
```bash
FPBRIDGE_LISTEN_ADDRESS=:8443 \
FPBRIDGE_EVENT_ENDPOINT=http://127.0.0.1:9100/events \
FPBRIDGE_ALLOWED_ORIGINS=http://localhost:5173 \
FPBRIDGE_TLS_CERT=certs/localhost-cert.pem \
FPBRIDGE_TLS_KEY=certs/localhost-key.pem \
go run ./cmd/fpbridge
```

| Variable | Used by | Purpose | Example |
|---|---|---|---|
| `VITE_FPBRIDGE_ENDPOINT` | Web application | URL web application sends browser fingerprint signals | `https://localhost:8443/v1/events` |
| `FPBRIDGE_LISTEN_ADDRESS` | FPBridge Proxy | Port FPBridge listens to | `:8443` |
| `FPBRIDGE_EVENT_ENDPOINT` | FPBridge Proxy | URL which FPBridge Proxy sends linked events to | `http://127.0.0.1:9100/events` |
| `FPBRIDGE_ALLOWED_ORIGINS` | FPBridge Proxy | list the websites (separated by comma) allowed to send browser requests to FPBridge Proxy  | `http://localhost:5173` |
| `FPBRIDGE_TLS_CERT` | FPBridge Proxy | Path to TLS certificate | `certs/localhost-cert.pem` |
| `FPBRIDGE_TLS_KEY` | FPBridge Proxy | Path to TLS private key | `certs/localhost-key.pem` |


Confirm that FPBridge is running:

```bash
curl -k https://localhost:8443/health
```

Expected response:

```json
{
  "status": "ok"
}
```

### Example Web Application
With the example backend and FPBridge Proxy running, start the web application:

```bash
cd examples/web
npm run dev
```

Open [http://localhost:5173](http://localhost:5173) in a browser.

The example application automatically:
1. Runs FPScanner via **FPBridge Browser** to collect browser signals.
2. Sends the browser signals to **FPBridge Proxy**.
3. **FPBridge Proxy** obtains the network signals, links them with the browser signals and forwards them to the detection backend.
4. The detection backend processes the linked event and returns the results to **FPBridge Proxy**.
5. **FPBridge Proxy** forwards the result to the web application.

**Successful response:**

```json
{
  "event_id": "evt_123",
  "result": {
    "action": "allow",
    "reason_codes": []
  }
}
```

The `event_id` is generated by FPBridge and `result` is defined by the detection backend.





## License
Apache-2.0
