import { describe, expect, it, vi } from "vitest";

import { FPBridge, FPBridgeError } from "../src";
import type { Collector } from "../src";

interface TestPayload {
	fingerprint: string;
}

const collector: Collector<TestPayload> = {
	name: "test-collector",
	version: "1.0.0",
	collect: vi.fn(async () => ({ fingerprint: "browser-123" })),
};

describe("FPBridge", () => {
	it("collects and submits a browser fingerprint", async () => {
		const fetchMock = vi.fn<typeof globalThis.fetch>(async () => new Response(
			JSON.stringify({ event_id: "evt_test123" }),
			{ status: 200, headers: { "Content-Type": "application/json" } },
		));
		const bridge = new FPBridge({
			endpoint: "https://fp.example.com/v1/events",
			collector,
			fetch: fetchMock,
		});

		const result = await bridge.collectAndSubmit();

		expect(result).toEqual({ event_id: "evt_test123" });
		expect(fetchMock).toHaveBeenCalledOnce();

		const call = fetchMock.mock.calls[0];
		if (!call) {
			throw new Error("fetch was not called");
		}

		const [endpoint, init] = call;
		expect(endpoint).toBe("https://fp.example.com/v1/events");

		if (!init) {
			throw new Error("fetch options are missing");
		}

		expect(init.method).toBe("POST");

		if (typeof init.body !== "string") {
			throw new Error("fetch request body is missing");
		}

		const submission = JSON.parse(init.body);
		expect(submission.collector).toEqual({
			name: "test-collector",
			version: "1.0.0",
		});
		expect(submission.payload).toEqual({ fingerprint: "browser-123" });
	});

	it("throws an FPBridgeError for a non-success response", async () => {
		const bridge = new FPBridge({
			endpoint: "https://fp.example.com/v1/events",
			collector,
			fetch: vi.fn(async () => new Response(null, { status: 502 })),
		});

		await expect(bridge.collectAndSubmit()).rejects.toEqual(
			expect.objectContaining<Partial<FPBridgeError>>({ status: 502 }),
		);
	});

	it("rejects an invalid acknowledgement", async () => {
		const bridge = new FPBridge({
			endpoint: "https://fp.example.com/v1/events",
			collector,
			fetch: vi.fn(async () => new Response(
				JSON.stringify({ result: { status: "missing_event_id" } }),
				{ status: 200 },
			)),
		});

		await expect(bridge.collectAndSubmit()).rejects.toThrow(
			"FPBridge returned an invalid response",
		);
	});

	it("returns an arbitrary backend result", async () => {
		const fetchMock = vi.fn<typeof globalThis.fetch>(async () => new Response(
			JSON.stringify({
				event_id: "evt_test123",
				result: {
					risk: 0.87,
					show_captcha: true,
					provider: { name: "example" },
				},
			}),
			{ status: 200, headers: { "Content-Type": "application/json" } },
		));
		const bridge = new FPBridge({
			endpoint: "https://fp.example.com/v1/events",
			collector,
			fetch: fetchMock,
		});

		interface BackendResult {
			risk: number;
			show_captcha: boolean;
			provider: { name: string };
		}
		const result = await bridge.collectAndSubmit<BackendResult>();

		expect(result.result).toEqual({
			risk: 0.87,
			show_captcha: true,
			provider: { name: "example" },
		});
	});

	it("does not prescribe the backend result schema", async () => {
		const bridge = new FPBridge({
			endpoint: "https://fp.example.com/v1/events",
			collector,
			fetch: vi.fn(async () => new Response(
				JSON.stringify({
					event_id: "evt_test123",
					result: {
						custom_status: "developer_defined",
						values: [1, 2, 3],
					},
				}),
				{ status: 200 },
			)),
		});

		await expect(bridge.collectAndSubmit()).resolves.toEqual({
			event_id: "evt_test123",
			result: {
				custom_status: "developer_defined",
				values: [1, 2, 3],
			},
		});
	});
});
