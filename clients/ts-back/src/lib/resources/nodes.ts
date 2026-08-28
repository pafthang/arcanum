import type { HttpClient } from '../http.js';
import { createMutation } from '../mutation.svelte.js';
import { prefetchQuery } from '../query-cache.js';
import { createQuery, whenPresent } from '../query.svelte.js';
import type { Json, Node, ReorderItem, Tags } from '../types.js';

export function nodesResource(http: HttpClient) {
	return {
		list: () =>
			createQuery({
				cacheKey: 'nodes.list',
				fn: () => http.get<Node[]>('/api/nodes')
			}),
		prefetchList: () =>
			prefetchQuery({
				cacheKey: 'nodes.list',
				fn: () => http.get<Node[]>('/api/nodes')
			}),
		byUuid: (uuid: () => string) =>
			createQuery({
				key: uuid,
				enabled: whenPresent(uuid),
				fn: () => http.get<Node>('/api/nodes/:uuid', undefined, { uuid: uuid() })
			}),
		tags: () => createQuery({ fn: () => http.get<Tags>('/api/nodes/tags') }),
		create: () => createMutation({ fn: (body: Json) => http.post<Node>('/api/nodes', body) }),
		update: () => createMutation({ fn: (body: Json) => http.patch<Node>('/api/nodes', body) }),
		remove: () =>
			createMutation({ fn: (uuid: string) => http.delete<unknown>('/api/nodes/:uuid', { uuid }) }),
		enable: () =>
			createMutation({
				fn: (uuid: string) => http.post<Node>('/api/nodes/:uuid/actions/enable', undefined, { uuid })
			}),
		disable: () =>
			createMutation({
				fn: (uuid: string) => http.post<Node>('/api/nodes/:uuid/actions/disable', undefined, { uuid })
			}),
		restart: () =>
			createMutation({
				fn: (vars: { uuid: string; forceRestart?: boolean }) =>
					http.post<unknown>(
						'/api/nodes/:uuid/actions/restart',
						{ forceRestart: vars.forceRestart ?? false },
						{ uuid: vars.uuid }
					)
			}),
		resetTraffic: () =>
			createMutation({
				fn: (uuid: string) =>
					http.post<Node>('/api/nodes/:uuid/actions/reset-traffic', undefined, { uuid })
			}),
		restartAll: () =>
			createMutation({
				fn: (body: { forceRestart: boolean }) => http.post<unknown>('/api/nodes/actions/restart-all', body)
			}),
		reorder: () =>
			createMutation({
				fn: (body: { nodes: ReorderItem[] }) => http.post<Node[]>('/api/nodes/actions/reorder', body)
			}),
		bulkActions: () =>
			createMutation({ fn: (body: Json) => http.post<unknown>('/api/nodes/bulk-actions', body) }),
		bulkUpdate: () =>
			createMutation({ fn: (body: Json) => http.post<unknown>('/api/nodes/bulk-actions/update', body) }),
		bulkProfile: () =>
			createMutation({
				fn: (body: Json) => http.post<unknown>('/api/nodes/bulk-actions/profile-modification', body)
			})
	};
}
