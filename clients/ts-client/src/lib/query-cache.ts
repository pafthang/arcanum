export const DEFAULT_STALE_TIME = 15_000;

type CacheRecord = {
	data: unknown;
	at: number;
};

const cache = new Map<string, CacheRecord>();
const inflight = new Map<string, Promise<unknown>>();

export type CacheKeyOptions = {
	cacheKey?: string;
	key?: () => unknown;
};

export function resolveCacheKey(opts: CacheKeyOptions): string | null {
	if (!opts.cacheKey) return null;
	try {
		return JSON.stringify([opts.cacheKey, opts.key?.() ?? null]);
	} catch {
		return opts.cacheKey;
	}
}

export function readQueryCache<T>(key: string): T | undefined {
	return cache.get(key)?.data as T | undefined;
}

export function getQueryCacheEntry(key: string): CacheRecord | undefined {
	return cache.get(key);
}

export function writeQueryCache<T>(key: string, data: T): void {
	cache.set(key, { data, at: Date.now() });
}

export function clearQueryCache(key?: string): void {
	if (key) {
		cache.delete(key);
		inflight.delete(key);
		return;
	}
	cache.clear();
	inflight.clear();
}

export function ensureQuery<T>(
	key: string,
	fn: () => Promise<T>,
	staleTime = DEFAULT_STALE_TIME
): Promise<T> {
	const hit = cache.get(key);
	if (hit && Date.now() - hit.at < staleTime) {
		return Promise.resolve(hit.data as T);
	}
	const pending = inflight.get(key);
	if (pending) return pending as Promise<T>;

	const request = fn()
		.then((data) => {
			writeQueryCache(key, data);
			return data;
		})
		.finally(() => {
			if (inflight.get(key) === request) inflight.delete(key);
		});

	inflight.set(key, request);
	return request;
}

export function prefetchQuery<T>(
	opts: CacheKeyOptions & { fn: () => Promise<T>; staleTime?: number }
): void {
	const key = resolveCacheKey(opts);
	if (!key) return;
	void ensureQuery(key, opts.fn, opts.staleTime);
}
