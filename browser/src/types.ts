export interface BrowserSubmission<TPayload = unknown> {
	schema_version: "1.0";
	collected_at: string;
	collector: {
		name: string;
		version: string;
	};
	payload: TPayload;
}

export interface SubmissionResponse<TResult = unknown> {
	event_id: string;
	result?: TResult;
}
