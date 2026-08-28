import assert from 'node:assert/strict';
import { readdirSync, readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';

import { GO_ROUTES } from './lib/paths.ts';

test('GO_ROUTES method+path is unique', () => {
	const seen = new Set<string>();
	for (const r of GO_ROUTES) {
		const key = `${r.method} ${r.path}`;
		assert.equal(seen.has(key), false, `duplicate ${key}`);
		seen.add(key);
	}
});

test('every Go route path is used by a resource', () => {
	const dir = join(dirname(fileURLToPath(import.meta.url)), 'lib/resources');
	let src = '';
	for (const name of readdirSync(dir)) {
		src += readFileSync(join(dir, name), 'utf8');
	}
	const missing = GO_ROUTES.filter((r) => !src.includes(`'${r.path}'`) && !src.includes(`"${r.path}"`)).map(
		(r) => `${r.method} ${r.path}`
	);
	assert.deepEqual(missing, []);
});
