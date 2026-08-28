import type { HttpClient } from '../http.js';
import { createMutation } from '../mutation.svelte.js';
import { prefetchQuery } from '../query-cache.js';
import { createQuery, whenPresent } from '../query.svelte.js';
import type { Host, Json, ReorderItem, Tags } from '../types.js';

export function hostsResource(http: HttpClient) {
	return {
		list: () =>
			createQuery({
				cacheKey: 'hosts.list',
				fn: () => http.get<Host[]>('/api/hosts')
			}),
		prefetchList: () =>
			prefetchQuery({
				cacheKey: 'hosts.list',
				fn: () => http.get<Host[]>('/api/hosts')
			}),
		byUuid: (uuid: () => string) =>
			createQuery({
				key: uuid,
				enabled: whenPresent(uuid),
				fn: () => http.get<Host>('/api/hosts/:uuid', undefined, { uuid: uuid() })
			}),
		tags: () => createQuery({ fn: () => http.get<Tags>('/api/hosts/tags') }),
		create: () => createMutation({ fn: (body: Json) => http.post<Host>('/api/hosts', body) }),
		update: () => createMutation({ fn: (body: Json) => http.patch<Host>('/api/hosts', body) }),
		remove: () =>
			createMutation({ fn: (uuid: string) => http.delete<unknown>('/api/hosts/:uuid', { uuid }) }),
		reorder: () =>
			createMutation({
				fn: (body: { hosts: ReorderItem[] }) => http.post<{ isUpdated: boolean }>('/api/hosts/actions/reorder', body)
			}),
		bulkEnable: () =>
			createMutation({ fn: (body: Json) => http.post<unknown>('/api/hosts/bulk/enable', body) }),
		bulkDisable: () =>
			createMutation({ fn: (body: Json) => http.post<unknown>('/api/hosts/bulk/disable', body) }),
		bulkDelete: () =>
			createMutation({ fn: (body: Json) => http.post<unknown>('/api/hosts/bulk/delete', body) }),
		bulkUpdate: () =>
			createMutation({ fn: (body: Json) => http.patch<unknown>('/api/hosts/bulk/update', body) })
	};
}
