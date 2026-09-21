import json
import os
from datetime import datetime, timezone

import urllib3


ENDPOINT = os.environ.get(
    "FPBRIDGE_ENDPOINT",
    "https://localhost:8443/v1/events",
)


def synthetic_browser_submission() -> dict[str, object]:
    """ Return a payload that claims to be a Chrome client.
    """
    return {
        "schema_version": "1.0",
        "collected_at": datetime.now(timezone.utc).isoformat(),
        "collector": {
            "name": "spoof-browser-demo",
            "version": "1.0.0",
        },
        "payload": {
            "signals": {
                "automation": {
                    "webdriver": False,
                    "webdriverWritable": False,
                    "selenium": False,
                    "cdp": False,
                    "playwright": False,
                },
                "device": {"platform": "MacIntel"},
                "browser": {
                    "userAgent": (
                        "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) "
                        "AppleWebKit/537.36 (KHTML, like Gecko) "
                        "Chrome/151.0.0.0 Safari/537.36"
                    ),
                },
            },
            "fastBotDetection": False,
            "fastBotDetectionDetails": {},
        },
    }


def main() -> None:
    # Demo uses a self-signed certificate.
    # Do not disable certificate verification in production.    
    urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

    http = urllib3.PoolManager(
        cert_reqs="CERT_NONE",
        timeout=urllib3.Timeout(total=10),
    )
    response = http.request(
        "POST",
        ENDPOINT,
        json=synthetic_browser_submission(),
        headers={"Content-Type": "application/json"},
    )

    print(f"HTTP {response.status}")
    try:
        result = json.loads(response.data.decode("utf-8"))
    except (UnicodeDecodeError, json.JSONDecodeError):
        print(response.data.decode("utf-8", errors="replace"))
        if response.status >= 400:
            raise RuntimeError(
                f"FPBridge returned HTTP {response.status}"
            )
        raise RuntimeError("FPBridge returned a non-JSON response")

    print(json.dumps(result, indent=2))
    if response.status >= 400:
        raise RuntimeError(f"FPBridge returned HTTP {response.status}")

    assessment = result.get("result", {})
    decision = assessment.get("decision", {})
    reasons = decision.get("reason_codes", [])
    if (
        decision.get("action") != "challenge"
        or "browser_tls_mismatch" not in reasons
    ):
        raise RuntimeError("expected challenge with browser_tls_mismatch")

    print(
        "PASS: the payload claimed human Chrome, while FPBridge observed "
        "Python's network stack."
    )


if __name__ == "__main__":
    main()
