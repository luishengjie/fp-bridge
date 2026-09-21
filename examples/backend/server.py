import json
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from typing import Any

from profiles import lookup_ja4



MAX_EVENT_BYTES = 1024 * 1024


def as_dict(value: Any) -> dict[str, Any]:
    return value if isinstance(value, dict) else {}


def validate_event(event: dict[str, Any]) -> str | None:
    """ Validate event structure
    """
    browser = event.get("browser")
    if not isinstance(browser, dict):
        return "browser_must_be_an_object"

    if not isinstance(browser.get("payload"), dict):
        return "browser_payload_must_be_an_object"

    if not isinstance(event.get("network"), dict):
        return "network_must_be_an_object"

    return None


def evaluate_event(event: dict[str, Any]) -> dict[str, Any]:    
    """ Example demo rules
    """
    browser_observation = as_dict(event.get("browser"))
    payload = as_dict(browser_observation.get("payload"))
    signals = as_dict(payload.get("signals"))
    browser_signals = as_dict(signals.get("browser"))
    user_agent = str(browser_signals.get("userAgent", ""))

    
    network = as_dict(event.get("network"))
    observed_ja4 = str(network.get("ja4", ""))

    claims_browser = any(
        value in user_agent
        for value in ("Chrome/", "Firefox/", "Safari/")
    )
    network_profile = lookup_ja4(observed_ja4)
    reason_codes: list[str] = []

    if payload.get("fastBotDetection") is True:
        reason_codes.append("browser_automation_detected")

    if (
        claims_browser
        and network_profile["classification"] == "raw_client"
    ):
        reason_codes.append("browser_tls_mismatch")

    if reason_codes:
        return {
            "action": "challenge",
            "reason_codes": reason_codes,
        }

    return {
        "action": "allow",
        "reason_codes": [],
    }


def build_detection_result(
    event: dict[str, Any],
    decision: dict[str, Any],
) -> dict[str, Any]:
    """Build the detection result returned to FPBridge."""
    network = dict(as_dict(event.get("network")))
    network_profile = lookup_ja4(str(network.get("ja4", "")))
    network["profile"] = network_profile

    return {
        "schema_version": event.get("schema_version"),
        "observed_at": event.get("observed_at"),
        "request": event.get("request"),
        "browser": event.get("browser"),
        "network": network,
        "decision": decision,
    }


class EventHandler(BaseHTTPRequestHandler):
    def send_json(
        self,
        status: int,
        value: dict[str, Any],
    ) -> None:
        body = json.dumps(value).encode("utf-8")

        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_POST(self) -> None:
        if self.path != "/events":
            self.send_json(404, {"error": "not_found"})
            return

        if self.headers.get_content_type() != "application/json":
            self.send_json(415, {"error": "content_type_must_be_json"})
            return

        try:
            content_length = int(self.headers.get("Content-Length", "0"))
        except ValueError:
            self.send_json(400, {"error": "invalid_content_length"})
            return

        if content_length <= 0:
            self.send_json(400, {"error": "empty_request_body"})
            return

        if content_length > MAX_EVENT_BYTES:
            self.send_json(413, {"error": "event_too_large"})
            return

        try:
            event = json.loads(self.rfile.read(content_length))
        except (UnicodeDecodeError, json.JSONDecodeError):
            self.send_json(400, {"error": "invalid_json"})
            return

        if not isinstance(event, dict):
            self.send_json(400, {"error": "event_must_be_an_object"})
            return

        validation_error = validate_event(event)
        if validation_error is not None:
            self.send_json(400, {"error": validation_error})
            return

        decision = evaluate_event(event)
        assessment = build_detection_result(event, decision)

        print(
            json.dumps(
                {
                    "event_id": event.get("event_id"),
                    **assessment,
                },
                indent=2,
            )
        )

        self.send_json(200, assessment)


if __name__ == "__main__":
    server = ThreadingHTTPServer(("127.0.0.1", 9100), EventHandler)
    print(
        "Example detection backend listening on http://127.0.0.1:9100/events"
    )
    server.serve_forever()
