import assert from 'node:assert/strict';
import { test } from 'node:test';

import { createUrl } from './lib/url.ts';

test('replaces route params', () => {
	assert.equal(createUrl('/api/users/:userId', undefined, { userId: 7 }), '/api/users/7');
});

test('appends query params and json-encodes objects', () => {
	const url = createUrl('/api/users', { start: 0, size: 25, filters: [{ id: 'username', value: 'a' }] });
	assert.equal(url.includes('start=0'), true);
	assert.equal(url.includes('size=25'), true);
	assert.equal(url.includes('filters='), true);
});

test('skips empty query values', () => {
	assert.equal(createUrl('/api/x', { a: '', b: null, c: undefined, d: 'ok' }), '/api/x?d=ok');
});

test('rejects empty route params instead of emitting a trailing slash', () => {
	assert.throws(() => createUrl('/api/subscription-page-configs/:uuid', undefined, { uuid: '' }), /missing route param "uuid"/);
	assert.throws(() => createUrl('/api/users/:userId', undefined, { userId: undefined }), /missing route param "userId"/);
});
