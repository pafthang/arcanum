import type { HttpClient } from '../http.js';
import { createMutation } from '../mutation.svelte.js';
import { createQuery, whenPresent } from '../query.svelte.js';
import type { Json } from '../types.js';

export function metadataResource(http: HttpClient) {
	return {
		user: (userId: () => number | string) =>
			createQuery({
				key: userId,
				enabled: whenPresent(userId),
				fn: () => http.get<Json>('/api/metadata/user/:userId', undefined, { userId: userId() })
			}),
		putUser: () =>
			createMutation({
				fn: (vars: { userId: number | string; body: Json }) =>
					http.put<Json>('/api/metadata/user/:userId', vars.body, { userId: vars.userId })
			}),
		node: (uuid: () => string) =>
			createQuery({
				key: uuid,
				enabled: whenPresent(uuid),
				fn: () => http.get<Json>('/api/metadata/node/:uuid', undefined, { uuid: uuid() })
			}),
		putNode: () =>
			createMutation({
				fn: (vars: { uuid: string; body: Json }) =>
					http.put<Json>('/api/metadata/node/:uuid', vars.body, { uuid: vars.uuid })
			})
	};
}
