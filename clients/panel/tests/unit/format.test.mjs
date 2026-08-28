import assert from 'node:assert/strict';
import test from 'node:test';
import { formatBytes, formatTrafficLimit, parsePrettyBytes, trafficPercent } from '../../src/lib/format.ts';

test('formatBytes uses binary units', () => {
	assert.equal(formatBytes(0), '0 B');
	assert.equal(formatBytes(1024), '1.00 KB');
	assert.equal(formatBytes(1024 ** 3), '1.00 GB');
});

test('formatTrafficLimit treats zero as unlimited', () => {
	assert.equal(formatTrafficLimit(0), 'Unlimited');
	assert.equal(formatTrafficLimit(1024 ** 2), '1.00 MB');
});

test('parsePrettyBytes reads unit suffixes', () => {
	assert.equal(parsePrettyBytes('1 GB'), 1024 ** 3);
	assert.equal(parsePrettyBytes(''), 0);
});

test('trafficPercent is null without a cap', () => {
	assert.equal(trafficPercent(100, 0), null);
	assert.equal(trafficPercent(50, 100), 50);
	assert.equal(trafficPercent(150, 100), 100);
});
