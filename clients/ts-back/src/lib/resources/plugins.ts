import type { HttpClient } from '../http.js';
import { createMutation } from '../mutation.svelte.js';
import { prefetchQuery } from '../query-cache.js';
import { createQuery, whenPresent } from '../query.svelte.js';
import type { Json, NodePlugin, NodePluginList, ReorderItem, SharedList, SharedListList } from '../types.js';

export function pluginsResource(http: HttpClient) {
	return {
		list: () =>
			createQuery({
				cacheKey: 'node-plugins.list',
				fn: () => http.get<NodePluginList>('/api/node-plugins')
			}),
		prefetchList: () =>
			prefetchQuery({
				cacheKey: 'node-plugins.list',
				fn: () => http.get<NodePluginList>('/api/node-plugins')
			}),
		byUuid: (uuid: () => string) =>
			createQuery({
				key: uuid,
				enabled: whenPresent(uuid),
				fn: () => http.get<NodePlugin>('/api/node-plugins/:uuid', undefined, { uuid: uuid() })
			}),
		create: () => createMutation({ fn: (body: Json) => http.post<NodePlugin>('/api/node-plugins', body) }),
		update: () => createMutation({ fn: (body: Json) => http.patch<NodePlugin>('/api/node-plugins', body) }),
		remove: () =>
			createMutation({ fn: (uuid: string) => http.delete<unknown>('/api/node-plugins/:uuid', { uuid }) }),
		reorder: () =>
			createMutation({
				fn: (body: { plugins?: ReorderItem[] } & Json) =>
					http.post<unknown>('/api/node-plugins/actions/reorder', body)
			}),
		clone: () =>
			createMutation({ fn: (body: Json) => http.post<NodePlugin>('/api/node-plugins/actions/clone', body) }),
		sync: () =>
			createMutation({ fn: (body: Json) => http.post<unknown>('/api/node-plugins/actions/sync', body) }),
		executor: () =>
			createMutation({ fn: (body: Json) => http.post<Json>('/api/node-plugins/executor', body) }),
		sharedLists: {
			list: () => createQuery({ fn: () => http.get<SharedListList>('/api/node-plugins/shared-lists') }),
			byName: (name: () => string) =>
				createQuery({
					key: name,
					enabled: whenPresent(name),
					fn: () =>
						http.get<SharedList>('/api/node-plugins/shared-lists/:name', undefined, { name: name() })
				}),
			create: () =>
				createMutation({ fn: (body: Json) => http.post<SharedList>('/api/node-plugins/shared-lists', body) }),
			update: () =>
				createMutation({ fn: (body: Json) => http.patch<SharedList>('/api/node-plugins/shared-lists', body) }),
			remove: () =>
				createMutation({
					fn: (name: string) => http.delete<unknown>('/api/node-plugins/shared-lists/:name', { name })
				}),
			sync: () =>
				createMutation({
					fn: (body?: Json) => http.post<unknown>('/api/node-plugins/shared-lists/actions/sync', body)
				})
		},
		torrentBlocker: {
			reports: (query: () => Record<string, unknown> = () => ({})) =>
				createQuery({
					key: query,
					fn: () => http.get<Json>('/api/node-plugins/torrent-blocker', query())
				}),
			stats: () => createQuery({ fn: () => http.get<Json>('/api/node-plugins/torrent-blocker/stats') }),
			truncate: () =>
				createMutation({
					fn: () => http.delete<unknown>('/api/node-plugins/torrent-blocker/truncate')
				})
		}
	};
}
