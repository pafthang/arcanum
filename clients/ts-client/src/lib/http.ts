import { ApiError, unwrapEnvelope } from './error.js';
import { createUrl } from './url.js';

export const ClientTypeHeader = 'X-Remnawave-Client-Type';
export const ClientTypeBrowser = 'browser';

export type HttpMethod = 'GET' | 'POST' | 'PATCH' | 'PUT' | 'DELETE';

export type ClientOptions = {
	baseUrl?: string;
	getToken?: () => string;
	onUnauthorized?: () => void;
	fetch?: typeof fetch;
	headers?: Record<string, string>;
};

export class HttpClient {
	readonly baseUrl: string;
	private readonly getToken: () => string;
	private readonly onUnauthorized?: () => void;
	private readonly fetchImpl: typeof fetch;
	private readonly extraHeaders: Record<string, string>;

	constructor(opts: ClientOptions = {}) {
		this.baseUrl = (opts.baseUrl ?? '').replace(/\/$/, '');
		this.getToken = opts.getToken ?? (() => '');
		this.onUnauthorized = opts.onUnauthorized;
		this.fetchImpl = opts.fetch ?? fetch;
		this.extraHeaders = opts.headers ?? {};
	}

	get<T>(path: string, query?: Record<string, unknown>, route?: Record<string, unknown>): Promise<T> {
		return this.request<T>('GET', path, { query, route });
	}

	post<T>(path: string, body?: unknown, route?: Record<string, unknown>, query?: Record<string, unknown>): Promise<T> {
		return this.request<T>('POST', path, { body, route, query });
	}

	patch<T>(path: string, body?: unknown, route?: Record<string, unknown>): Promise<T> {
		return this.request<T>('PATCH', path, { body, route });
	}

	put<T>(path: string, body?: unknown, route?: Record<string, unknown>): Promise<T> {
		return this.request<T>('PUT', path, { body, route });
	}

	delete<T>(path: string, route?: Record<string, unknown>, body?: unknown): Promise<T> {
		return this.request<T>('DELETE', path, { route, body });
	}

	async request<T>(
		method: HttpMethod,
		path: string,
		opts: { query?: Record<string, unknown>; route?: Record<string, unknown>; body?: unknown } = {}
	): Promise<T> {
		const url = this.baseUrl + createUrl(path, opts.query, opts.route);
		const headers = new Headers(this.extraHeaders);
		headers.set('Accept', 'application/json');
		headers.set(ClientTypeHeader, ClientTypeBrowser);
		const token = this.getToken();
		if (token) headers.set('Authorization', `Bearer ${token}`);
		if (opts.body !== undefined) headers.set('Content-Type', 'application/json');

		const res = await this.fetchImpl(url, {
			method,
			headers,
			body: opts.body === undefined ? undefined : JSON.stringify(opts.body)
		});

		const raw = await res.text();
		try {
			return unwrapEnvelope<T>(res.status, raw, path);
		} catch (err) {
			if (err instanceof ApiError && err.unauthorized) this.onUnauthorized?.();
			throw err;
		}
	}
}
