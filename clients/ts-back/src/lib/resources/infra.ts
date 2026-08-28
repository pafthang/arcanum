import type { HttpClient } from '../http.js';
import { createMutation } from '../mutation.svelte.js';
import { createQuery, whenPresent } from '../query.svelte.js';
import type { Json } from '../types.js';

export function infraResource(http: HttpClient) {
	return {
		providers: {
			list: () => createQuery({ fn: () => http.get<Json>('/api/infra-billing/providers') }),
			byUuid: (uuid: () => string) =>
				createQuery({
					key: uuid,
					enabled: whenPresent(uuid),
					fn: () => http.get<Json>('/api/infra-billing/providers/:uuid', undefined, { uuid: uuid() })
				}),
			create: () =>
				createMutation({ fn: (body: Json) => http.post<Json>('/api/infra-billing/providers', body) }),
			update: () =>
				createMutation({ fn: (body: Json) => http.patch<Json>('/api/infra-billing/providers', body) }),
			remove: () =>
				createMutation({
					fn: (uuid: string) => http.delete<unknown>('/api/infra-billing/providers/:uuid', { uuid })
				})
		},
		nodes: {
			list: () => createQuery({ fn: () => http.get<Json>('/api/infra-billing/nodes') }),
			create: () => createMutation({ fn: (body: Json) => http.post<Json>('/api/infra-billing/nodes', body) }),
			update: () => createMutation({ fn: (body: Json) => http.patch<Json>('/api/infra-billing/nodes', body) }),
			remove: () =>
				createMutation({
					fn: (uuid: string) => http.delete<unknown>('/api/infra-billing/nodes/:uuid', { uuid })
				})
		},
		history: {
			list: () => createQuery({ fn: () => http.get<Json>('/api/infra-billing/history') }),
			create: () =>
				createMutation({ fn: (body: Json) => http.post<Json>('/api/infra-billing/history', body) }),
			remove: () =>
				createMutation({
					fn: (uuid: string) => http.delete<unknown>('/api/infra-billing/history/:uuid', { uuid })
				})
		}
	};
}
