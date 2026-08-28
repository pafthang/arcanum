import type { HttpClient } from '../http.js';
import { createMutation } from '../mutation.svelte.js';
import { prefetchQuery } from '../query-cache.js';
import { createQuery } from '../query.svelte.js';
import { session } from '../session.svelte.js';
import type { BandwidthStats, Json, SystemMetadata, SystemStats } from '../types.js';

function authed<T>(cacheKey: string, fn: () => Promise<T>, key?: () => unknown) {
	return createQuery({
		cacheKey,
		key: () => [session.token, key?.() ?? null],
		enabled: () => session.isAuthenticated,
		fn
	});
}

function prefetchAuthed<T>(cacheKey: string, fn: () => Promise<T>, key?: () => unknown) {
	if (!session.isAuthenticated) return;
	prefetchQuery({
		cacheKey,
		key: () => [session.token, key?.() ?? null],
		fn
	});
}

export function systemResource(http: HttpClient) {
	return {
		stats: (tz: () => string = () => '') =>
			authed('system.stats', () => http.get<SystemStats>('/api/system/stats', tz() ? { tz: tz() } : undefined), tz),
		metadata: () => authed('system.metadata', () => http.get<SystemMetadata>('/api/system/metadata')),
		configuration: () => authed('system.configuration', () => http.get<Json>('/api/system/configuration')),
		health: () => authed('system.health', () => http.get<Json>('/api/system/health')),
		bandwidth: (tz: () => string = () => '') =>
			authed(
				'system.bandwidth',
				() => http.get<BandwidthStats>('/api/system/stats/bandwidth', tz() ? { tz: tz() } : undefined),
				tz
			),
		nodesStatistics: (tz: () => string = () => '') =>
			authed(
				'system.nodesStatistics',
				() => http.get<Json>('/api/system/stats/nodes', tz() ? { tz: tz() } : undefined),
				tz
			),
		recap: () => authed('system.recap', () => http.get<Json>('/api/system/stats/recap')),
		digest: () => authed('system.digest', () => http.get<Json>('/api/system/stats/digest')),
		httpStats: () => authed('system.httpStats', () => http.get<Json>('/api/system/stats/http')),
		nodesMetrics: () => authed('system.nodesMetrics', () => http.get<Json>('/api/system/nodes/metrics')),
		x25519: () => authed('system.x25519', () => http.get<Json>('/api/system/tools/x25519/generate')),
		prefetchHome: (tz: string = '') => {
			const tzKey = () => tz;
			prefetchAuthed(
				'system.stats',
				() => http.get<SystemStats>('/api/system/stats', tz ? { tz } : undefined),
				tzKey
			);
			prefetchAuthed('system.metadata', () => http.get<SystemMetadata>('/api/system/metadata'));
			prefetchAuthed('system.health', () => http.get<Json>('/api/system/health'));
			prefetchAuthed(
				'system.bandwidth',
				() => http.get<BandwidthStats>('/api/system/stats/bandwidth', tz ? { tz } : undefined),
				tzKey
			);
			prefetchAuthed(
				'system.nodesStatistics',
				() => http.get<Json>('/api/system/stats/nodes', tz ? { tz } : undefined),
				tzKey
			);
		},
		srrMatcher: () =>
			createMutation({ fn: (body: Json) => http.post<Json>('/api/system/testers/srr-matcher', body) })
	};
}
