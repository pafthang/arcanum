import type { HttpClient } from '../http.js';
import { createMutation } from '../mutation.svelte.js';
import { createQuery } from '../query.svelte.js';
import type { Json } from '../types.js';

export function connectionsResource(http: HttpClient) {
	return {
		startByUser: () =>
			createMutation({
				fn: (vars: { userId: number | string; body?: Json }) =>
					http.post<Json>('/api/connections/by-user/:userId', vars.body, { userId: vars.userId })
			}),
		resultByUser: (jobId: () => string) =>
			createQuery({
				key: jobId,
				enabled: () => Boolean(jobId()),
				fn: () => http.get<Json>('/api/connections/by-user/:jobId', undefined, { jobId: jobId() })
			}),
		startByNode: () =>
			createMutation({
				fn: (vars: { uuid: string; body?: Json }) =>
					http.post<Json>('/api/connections/by-node/:uuid', vars.body, { uuid: vars.uuid })
			}),
		resultByNode: (jobId: () => string) =>
			createQuery({
				key: jobId,
				enabled: () => Boolean(jobId()),
				fn: () => http.get<Json>('/api/connections/by-node/:jobId', undefined, { jobId: jobId() })
			}),
		startGeocheck: () =>
			createMutation({
				fn: (vars: { uuid: string; body?: Json }) =>
					http.post<Json>('/api/connections/geocheck/:uuid', vars.body, { uuid: vars.uuid })
			}),
		geocheckResult: (jobId: () => string) =>
			createQuery({
				key: jobId,
				enabled: () => Boolean(jobId()),
				fn: () => http.get<Json>('/api/connections/geocheck/:jobId', undefined, { jobId: jobId() })
			}),
		drop: () => createMutation({ fn: (body: Json) => http.post<Json>('/api/connections/drop', body) })
	};
}
