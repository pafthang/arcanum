import type { HttpClient } from '../http.js';
import { createMutation } from '../mutation.svelte.js';
import { createQuery, whenPresent } from '../query.svelte.js';
import type { Json } from '../types.js';

export function integrationsResource(http: HttpClient) {
	return {
		list: () => createQuery({ fn: () => http.get<Json>('/api/node-integrations') }),
		byUuid: (uuid: () => string) =>
			createQuery({
				key: uuid,
				enabled: whenPresent(uuid),
				fn: () => http.get<Json>('/api/node-integrations/:uuid', undefined, { uuid: uuid() })
			}),
		create: () => createMutation({ fn: (body: Json) => http.post<Json>('/api/node-integrations', body) }),
		update: () => createMutation({ fn: (body: Json) => http.patch<Json>('/api/node-integrations', body) }),
		remove: () =>
			createMutation({
				fn: (uuid: string) => http.delete<unknown>('/api/node-integrations/:uuid', { uuid })
			})
	};
}
