import json
from collections import defaultdict
from pathlib import Path
from typing import Any


PROFILE_PATH = Path(__file__).parent / "data" / "ja4_profiles.jsonl"


def load_profiles(path: Path) -> dict[str, list[dict[str, Any]]]:
    """Load FPBridge demo profiles."""
    profiles: defaultdict[str, list[dict[str, Any]]] = defaultdict(list)

    with path.open(encoding="utf-8") as mapping_file:
        for line_number, line in enumerate(mapping_file, start=1):
            if not line.strip():
                continue

            try:
                record = json.loads(line)
            except json.JSONDecodeError as error:
                raise ValueError(
                    f"invalid local JA4 JSON on line {line_number}: {error}"
                ) from error

            if not isinstance(record, dict):
                raise ValueError(
                    f"local JA4 profile on line {line_number} must be an object"
                )

            required_fields = {"ja4", "classification", "label"}
            if not required_fields.issubset(record):
                raise ValueError(
                    f"local JA4 profile on line {line_number} is missing "
                    "ja4, classification, or label"
                )

            fingerprint = str(record.pop("ja4")).strip()
            if not fingerprint:
                raise ValueError(
                    f"local JA4 profile on line {line_number} has an empty ja4"
                )

            record.setdefault("source", "locally_observed")
            profiles[fingerprint].append(record)

    return dict(profiles)


JA4_LOOKUP = load_profiles(PROFILE_PATH)


def lookup_ja4(fingerprint: str) -> dict[str, Any]:
    matches = JA4_LOOKUP.get(fingerprint, [])
    if not matches:
        return {
            "classification": "unknown",
            "matches": [],
        }

    classifications = {match["classification"] for match in matches}
    if len(classifications) == 1:
        classification = next(iter(classifications))
    else:
        classification = "ambiguous"

    return {
        "classification": classification,
        "matches": matches,
    }
