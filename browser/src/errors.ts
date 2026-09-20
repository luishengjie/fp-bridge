export class FPBridgeError extends Error {
	constructor(
		message: string,
		readonly status?: number,
	) {
		super(message);
		this.name = "FPBridgeError";
	}
}
