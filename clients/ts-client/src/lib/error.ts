export class ApiError extends Error {
	readonly status: number;
	readonly errorCode: string;
	readonly path: string;
	readonly timestamp: string;

	constructor(init: { status: number; message: string; errorCode?: string; path?: string; timestamp?: string }) {
		super(init.message);
		this.name = 'ApiError';
		this.status = init.status;
		this.errorCode = init.errorCode ?? '';
		this.path = init.path ?? '';
		this.timestamp = init.timestamp ?? '';
	}

	get unauthorized(): boolean {
		return this.status === 401;
	}
}

export type Envelope<T> = { response: T };

export type ErrorBody = {
	timestamp?: string;
	path?: string;
	message?: string;
	errorCode?: string;
	statusCode?: number;
	error?: string;
};

export function unwrapEnvelope<T>(status: number, raw: string, path: string): T {
	let json: Envelope<T> | ErrorBody | null = null;
	if (raw) {
		try {
			json = JSON.parse(raw) as Envelope<T> | ErrorBody;
		} catch {
			json = null;
		}
	}
	if (status < 200 || status >= 300) {
		const errBody = (json ?? {}) as ErrorBody;
		throw new ApiError({
			status,
			message: errBody.message ?? `HTTP ${status}`,
			errorCode: errBody.errorCode ?? '',
			path: errBody.path ?? path,
			timestamp: errBody.timestamp ?? ''
		});
	}
	if (json && typeof json === 'object' && 'response' in json) {
		return (json as Envelope<T>).response;
	}
	return json as T;
}
