import type { HttpClient } from '../http.js';
import { createMutation } from '../mutation.svelte.js';
import { prefetchQuery } from '../query-cache.js';
import { createQuery, whenPresent } from '../query.svelte.js';
import type { ConfigInbound, ConfigProfile, ConfigProfileList, Json, ReorderItem } from '../types.js';

export function configProfilesResource(http: HttpClient) {
	return {
		list: () =>
			createQuery({
				cacheKey: 'config-profiles.list',
				fn: () => http.get<ConfigProfileList>('/api/config-profiles')
			}),
		prefetchList: () =>
			prefetchQuery({
				cacheKey: 'config-profiles.list',
				fn: () => http.get<ConfigProfileList>('/api/config-profiles')
			}),
		byUuid: (uuid: () => string) =>
			createQuery({
				key: uuid,
				enabled: whenPresent(uuid),
				fn: () => http.get<ConfigProfile>('/api/config-profiles/:uuid', undefined, { uuid: uuid() })
			}),
		inbounds: (uuid: () => string) =>
			createQuery({
				key: uuid,
				enabled: whenPresent(uuid),
				fn: () =>
					http.get<{ total: number; inbounds: ConfigInbound[] }>(
						'/api/config-profiles/:uuid/inbounds',
						undefined,
						{ uuid: uuid() }
					)
			}),
		allInbounds: () =>
			createQuery({
				cacheKey: 'config-profiles.inbounds',
				fn: () => http.get<Json>('/api/config-profiles/inbounds')
			}),
		prefetchAllInbounds: () =>
			prefetchQuery({
				cacheKey: 'config-profiles.inbounds',
				fn: () => http.get<Json>('/api/config-profiles/inbounds')
			}),
		computedConfig: (uuid: () => string) =>
			createQuery({
				key: uuid,
				enabled: whenPresent(uuid),
				fn: () =>
					http.get<Json>('/api/config-profiles/:uuid/computed-config', undefined, { uuid: uuid() })
			}),
		create: () => createMutation({ fn: (body: Json) => http.post<ConfigProfile>('/api/config-profiles', body) }),
		update: () => createMutation({ fn: (body: Json) => http.patch<ConfigProfile>('/api/config-profiles', body) }),
		remove: () =>
			createMutation({
				fn: (uuid: string) => http.delete<unknown>('/api/config-profiles/:uuid', { uuid })
			}),
		reorder: () =>
			createMutation({
				fn: (body: { profiles?: ReorderItem[] } & Json) =>
					http.post<unknown>('/api/config-profiles/actions/reorder', body)
			})
	};
}
