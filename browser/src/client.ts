import type { Collector } from "./collector";
import { FPBridgeError } from "./errors";
import type {
	BrowserSubmission,
	SubmissionResponse,
} from "./types";

export interface FPBridgeOptions<TPayload = unknown> {
	endpoint: string;
	collector: Collector<TPayload>;
	fetch?: typeof globalThis.fetch;
}

export class FPBridge<TPayload = unknown> {
	private readonly fetch: typeof globalThis.fetch;

	constructor(private readonly options: FPBridgeOptions<TPayload>) {
		if (!options.fetch && typeof globalThis.fetch !== "function") {
			throw new FPBridgeError("fetch is not available in this environment");
		}

		this.fetch = options.fetch ?? ((input, init) => globalThis.fetch(input, init));
	}

	async collectAndSubmit<TResult = unknown>(): Promise<SubmissionResponse<TResult>> {
		const payload = await this.options.collector.collect();

		const submission: BrowserSubmission<TPayload> = {
			schema_version: "1.0",
			collected_at: new Date().toISOString(),
			collector: {
				name: this.options.collector.name,
				version: this.options.collector.version,
			},
			payload,
		};

		const response = await this.fetch(this.options.endpoint, {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify(submission),
		});

		if (!response.ok) {
			throw new FPBridgeError(
				`FPBridge returned HTTP ${response.status}`,
				response.status,
			);
		}

		const result: unknown = await response.json();

		if (!isSubmissionResponse<TResult>(result)) {
			throw new FPBridgeError("FPBridge returned an invalid response");
		}

		return result;
	}
}

function isSubmissionResponse<TResult>(
	value: unknown,
): value is SubmissionResponse<TResult> {
	if (typeof value !== "object" || value === null) {
		return false;
	}

	const response = value as Record<string, unknown>;
	if (typeof response.event_id !== "string") {
		return false;
	}

	return true;
}
