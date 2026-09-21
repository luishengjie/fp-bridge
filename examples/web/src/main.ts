import {
	FPBridge,
	FPBridgeError,
	FPScannerAdapter,
} from "@fpbridge/browser";

import "./style.css";

const output = requireElement<HTMLElement>("#output");

interface DemoResult {
	schema_version: string;
	observed_at: string;
	request: Record<string, unknown>;
	browser: Record<string, unknown>;
	network: Record<string, unknown>;
	decision: {
		action: "allow" | "challenge" | "review" | "block";
		reason_codes: string[];
	};
}

const fpbridgeEndpoint = import.meta.env.VITE_FPBRIDGE_ENDPOINT;
if (!fpbridgeEndpoint) {
	throw new Error("VITE_FPBRIDGE_ENDPOINT is not configured");
}

const bridge = new FPBridge({
	endpoint: fpbridgeEndpoint,
	collector: new FPScannerAdapter(),
});

async function runAssessment(): Promise<void> {
	output.dataset.status = "pending";
	output.textContent = "Collecting and evaluating signals...";

	try {
		const result = await bridge.collectAndSubmit<DemoResult>();

		output.textContent = JSON.stringify(result, null, 2);

		switch (result.result?.decision.action) {
			case "allow":
				output.dataset.status = "success";
				break;

			case "challenge":
			case "review":
				output.dataset.status = "challenge";
				break;

			case "block":
				output.dataset.status = "error";
				break;

			default:
				output.dataset.status = "error";
				output.textContent += "\n\nNo demo action was returned.";
		}
	} catch (error) {
		output.dataset.status = "error";

		if (error instanceof FPBridgeError && error.status) {
			output.textContent = `${error.message}\nHTTP status: ${error.status}`;
		} else {
			output.textContent = error instanceof Error
				? error.message
				: String(error);
		}
	}
}

void runAssessment();

function requireElement<T extends Element>(selector: string): T {
	const element = document.querySelector<T>(selector);
	if (!element) {
		throw new Error(`missing required element: ${selector}`);
	}

	return element;
}
