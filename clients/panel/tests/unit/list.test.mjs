import assert from 'node:assert/strict';
import test from 'node:test';
import { asArray, pretty } from '../../src/lib/list.ts';

test('asArray returns arrays as-is', () => {
	assert.deepEqual(asArray([1, 2], ['items']), [1, 2]);
});

test('asArray unwraps the first matching object key', () => {
	assert.deepEqual(asArray({ users: [{ id: 1 }] }, ['items', 'users']), [{ id: 1 }]);
});

test('asArray returns empty for unknown shapes', () => {
	assert.deepEqual(asArray(null, ['items']), []);
	assert.deepEqual(asArray({ count: 2 }, ['items']), []);
});

test('pretty stringifies JSON with indentation', () => {
	assert.equal(pretty({ a: 1 }), '{\n  "a": 1\n}');
});
