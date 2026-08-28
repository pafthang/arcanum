import assert from 'node:assert/strict';
import { test } from 'node:test';

import {
	clearQueryCache,
	ensureQuery,
	getQueryCacheEntry,
	prefetchQuery,
	readQueryCache,
	resolveCacheKey
} from './lib/query-cache.ts';

test('resolveCacheKey requires cacheKey and serializes key()', () => {
	assert.equal(resolveCacheKey({}), null);
	assert.equal(resolveCacheKey({ key: () => 'users' }), null);
	assert.equal(resolveCacheKey({ cacheKey: 'users.list' }), JSON.stringify(['users.list', null]));
	assert.equal(
		resolveCacheKey({ cacheKey: 'users.list', key: () => ({ start: 0, size: 1000 }) }),
		JSON.stringify(['users.list', { start: 0, size: 1000 }])
	);
});

test('ensureQuery coalesces in-flight requests and serves fresh cache', async () => {
	clearQueryCache();
	let calls = 0;
	const fn = async () => {
		calls += 1;
		await new Promise((resolve) => setTimeout(resolve, 20));
		return calls;
	};

	const [a, b] = await Promise.all([ensureQuery('k', fn, 0), ensureQuery('k', fn, 0)]);
	assert.equal(a, 1);
	assert.equal(b, 1);
	assert.equal(calls, 1);
	assert.equal(readQueryCache('k'), 1);

	const cached = await ensureQuery('k', fn, 60_000);
	assert.equal(cached, 1);
	assert.equal(calls, 1);
	assert.ok((getQueryCacheEntry('k')?.at ?? 0) > 0);
});

test('prefetchQuery writes the same key createQuery would read', async () => {
	clearQueryCache();
	prefetchQuery({
		cacheKey: 'internal-squads.list',
		fn: async () => ({ internalSquads: [{ uuid: '1' }] })
	});
	const key = resolveCacheKey({ cacheKey: 'internal-squads.list' });
	assert.ok(key);
	const data = await ensureQuery(key, async () => ({ internalSquads: [] }), 60_000);
	assert.deepEqual(data, { internalSquads: [{ uuid: '1' }] });
});
