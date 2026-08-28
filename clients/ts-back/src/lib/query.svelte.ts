import {
	DEFAULT_STALE_TIME,
	ensureQuery,
	getQueryCacheEntry,
	resolveCacheKey,
	type CacheKeyOptions
} from './query-cache.js';

export type QueryStatus = 'idle' | 'loading' | 'success' | 'error';

export type QueryOptions<T> = CacheKeyOptions & {
	fn: () => Promise<T>;
	enabled?: () => boolean;
	staleTime?: number;
};

/** Skip a parameterized query until the route id is a non-empty value. */
export function whenPresent(get: () => unknown): () => boolean {
	return () => {
		const value = get();
		return value !== undefined && value !== null && String(value).trim() !== '';
	};
}

export type Query<T> = {
	readonly data: T | undefined;
	readonly error: Error | undefined;
	readonly status: QueryStatus;
	readonly loading: boolean;
	refetch: () => Promise<void>;
};

export function createQuery<T>(opts: QueryOptions<T>): Query<T> {
	const initialKey = resolveCacheKey(opts);
	const initial = initialKey ? getQueryCacheEntry(initialKey) : undefined;

	// Keep JSON payloads unproxied. Vincjo's $state.snapshot uses structuredClone,
	// which throws DataCloneError on Svelte 5 proxies.
	let data = $state.raw<T | undefined>(initial ? (initial.data as T) : undefined);
	let error = $state<Error | undefined>(undefined);
	let status = $state<QueryStatus>(initial ? 'success' : 'idle');
	let hydrated = Boolean(initial);

	async function refetch() {
		if (opts.enabled && !opts.enabled()) return;
		if (!hydrated) status = 'loading';
		const cacheKey = resolveCacheKey(opts);
		try {
			const next = cacheKey ? await ensureQuery(cacheKey, opts.fn, 0) : await opts.fn();
			hydrated = true;
			data = next;
			error = undefined;
			status = 'success';
		} catch (e) {
			error = e instanceof Error ? e : new Error(String(e));
			status = 'error';
		}
	}

	$effect(() => {
		opts.key?.();
		if (opts.enabled && !opts.enabled()) return;
		const cacheKey = resolveCacheKey(opts);
		const staleTime = opts.staleTime ?? DEFAULT_STALE_TIME;
		if (cacheKey) {
			const hit = getQueryCacheEntry(cacheKey);
			if (hit) {
				hydrated = true;
				data = hit.data as T;
				error = undefined;
				status = 'success';
				if (Date.now() - hit.at < staleTime) return;
			}
		}
		void refetch();
	});

	return {
		get data() {
			return data;
		},
		get error() {
			return error;
		},
		get status() {
			return status;
		},
		get loading() {
			return !hydrated && (status === 'loading' || status === 'idle');
		},
		refetch
	};
}
