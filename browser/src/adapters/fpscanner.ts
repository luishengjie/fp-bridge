import FingerprintScanner, {
    type Fingerprint,
} from "fpscanner";

import type { Collector } from "../collector";

export class FPScannerAdapter implements Collector<Fingerprint> {
	readonly name = "fpscanner";
	readonly version = "1.0.8";

	private readonly scanner = new FingerprintScanner();

	async collect(): Promise<Fingerprint> {
		return this.scanner.collectFingerprint({
			encrypt: false,
		}) as Promise<Fingerprint>;
	}
}
