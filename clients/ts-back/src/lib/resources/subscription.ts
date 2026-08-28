import type { HttpClient } from '../http.js';
import { createMutation } from '../mutation.svelte.js';
import { createQuery, whenPresent } from '../query.svelte.js';
import type { Json, ReorderItem, TableQuery } from '../types.js';

export function subscriptionResource(http: HttpClient) {
	return {
		get: (shortUuid: () => string) =>
			createQuery({
				key: shortUuid,
				enabled: whenPresent(shortUuid),
				fn: () => http.get<string>('/api/sub/:shortUuid', undefined, { shortUuid: shortUuid() })
			}),
		info: (shortUuid: () => string) =>
			createQuery({
				key: shortUuid,
				enabled: whenPresent(shortUuid),
				fn: () => http.get<Json>('/api/sub/:shortUuid/info', undefined, { shortUuid: shortUuid() })
			}),
		clientType: (route: () => { shortUuid: string; clientType: string }) =>
			createQuery({
				key: route,
				enabled: () => Boolean(route().shortUuid && route().clientType),
				fn: () => http.get<string>('/api/sub/:shortUuid/:clientType', undefined, route())
			}),
		requestHistory: {
			list: (query: () => TableQuery = () => ({})) =>
				createQuery({
					key: query,
					fn: () => http.get<Json>('/api/subscription-request-history', query())
				}),
			stats: () => createQuery({ fn: () => http.get<Json>('/api/subscription-request-history/stats') })
		},
		pageConfigs: {
			list: () => createQuery({ fn: () => http.get<Json>('/api/subscription-page-configs') }),
			byUuid: (uuid: () => string) =>
				createQuery({
					key: uuid,
					enabled: whenPresent(uuid),
					fn: () => http.get<Json>('/api/subscription-page-configs/:uuid', undefined, { uuid: uuid() })
				}),
			create: () =>
				createMutation({ fn: (body: Json) => http.post<Json>('/api/subscription-page-configs', body) }),
			update: () =>
				createMutation({ fn: (body: Json) => http.patch<Json>('/api/subscription-page-configs', body) }),
			remove: () =>
				createMutation({
					fn: (uuid: string) => http.delete<unknown>('/api/subscription-page-configs/:uuid', { uuid })
				}),
			reorder: () =>
				createMutation({
					fn: (body: { configs?: ReorderItem[] } & Json) =>
						http.post<unknown>('/api/subscription-page-configs/actions/reorder', body)
				}),
			clone: () =>
				createMutation({
					fn: (body: Json) => http.post<Json>('/api/subscription-page-configs/actions/clone', body)
				})
		},
		templates: {
			list: () => createQuery({ fn: () => http.get<Json>('/api/subscription-templates') }),
			byUuid: (uuid: () => string) =>
				createQuery({
					key: uuid,
					enabled: whenPresent(uuid),
					fn: () => http.get<Json>('/api/subscription-templates/:uuid', undefined, { uuid: uuid() })
				}),
			create: () =>
				createMutation({ fn: (body: Json) => http.post<Json>('/api/subscription-templates', body) }),
			update: () =>
				createMutation({ fn: (body: Json) => http.patch<Json>('/api/subscription-templates', body) }),
			remove: () =>
				createMutation({
					fn: (uuid: string) => http.delete<unknown>('/api/subscription-templates/:uuid', { uuid })
				}),
			reorder: () =>
				createMutation({
					fn: (body: Json) => http.post<unknown>('/api/subscription-templates/actions/reorder', body)
				})
		},
		settings: {
			get: () => createQuery({ fn: () => http.get<Json>('/api/subscription-settings') }),
			update: () =>
				createMutation({ fn: (body: Json) => http.patch<Json>('/api/subscription-settings', body) })
		}
	};
}
