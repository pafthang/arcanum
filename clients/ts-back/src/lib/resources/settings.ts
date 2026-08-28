import type { HttpClient } from '../http.js';
import { createMutation } from '../mutation.svelte.js';
import { createQuery } from '../query.svelte.js';
import type { Json, Keygen } from '../types.js';

export function settingsResource(http: HttpClient) {
	return {
		get: () =>
			createQuery({
				cacheKey: 'remnawave-settings',
				fn: () => http.get<Json>('/api/remnawave-settings')
			}),
		update: () => createMutation({ fn: (body: Json) => http.patch<Json>('/api/remnawave-settings', body) }),
		keygen: () => createQuery({ fn: () => http.get<Keygen>('/api/keygen') })
	};
}
