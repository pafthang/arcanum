import assert from 'node:assert/strict';
import { test } from 'node:test';

import { ApiError, unwrapEnvelope } from './lib/error.ts';

test('unwraps success envelope', () => {
	const got = unwrapEnvelope<{ accessToken: string }>(
		200,
		JSON.stringify({ response: { accessToken: 't' } }),
		'/api/auth/login'
	);
	assert.equal(got.accessToken, 't');
});

test('maps Go error envelope', () => {
	assert.throws(
		() =>
			unwrapEnvelope(
				401,
				JSON.stringify({
					timestamp: '2026-01-01T00:00:00.000Z',
					path: '/api/users',
					message: 'Unauthorized',
					errorCode: 'A001'
				}),
				'/api/users'
			),
		(err: unknown) => {
			assert.ok(err instanceof ApiError);
			assert.equal(err.status, 401);
			assert.equal(err.errorCode, 'A001');
			assert.equal(err.unauthorized, true);
			return true;
		}
	);
});

test('403 forbidden does not count as session logout', () => {
	assert.throws(
		() =>
			unwrapEnvelope(
				403,
				JSON.stringify({
					timestamp: '2026-01-01T00:00:00.000Z',
					path: '/api/auth/passkey/authentication/options',
					message: 'Forbidden',
					errorCode: 'A068'
				}),
				'/api/auth/passkey/authentication/options'
			),
		(err: unknown) => {
			assert.ok(err instanceof ApiError);
			assert.equal(err.status, 403);
			assert.equal(err.unauthorized, false);
			return true;
		}
	);
});
