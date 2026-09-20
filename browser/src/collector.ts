export interface Collector<TPayload = unknown> {
	readonly name: string;
	readonly version: string;
	collect(): Promise<TPayload>;
}
