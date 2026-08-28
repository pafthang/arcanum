import type { HttpClient } from '../http.js';
import { createMutation } from '../mutation.svelte.js';
import { createQuery, whenPresent } from '../query.svelte.js';
import type { Json, NodesUsage } from '../types.js';

export function bandwidthResource(http: HttpClient) {
	return {
		nodes: (query: () => Record<string, unknown> = () => ({})) =>
			createQuery({
				key: query,
				fn: () => http.get<NodesUsage>('/api/bandwidth-stats/nodes', query())
			}),
		nodeUsers: (uuid: () => string, query: () => Record<string, unknown> = () => ({})) =>
			createQuery({
				key: () => [uuid(), query()],
				enabled: whenPresent(uuid),
				fn: () =>
					http.get<Json>('/api/bandwidth-stats/nodes/:uuid/users', query(), { uuid: uuid() })
			}),
		nodesUsers: () =>
			createMutation({ fn: (body: Json) => http.post<Json>('/api/bandwidth-stats/nodes/users', body) }),
		nodesUsage: () =>
			createMutation({ fn: (body: Json) => http.post<Json>('/api/bandwidth-stats/nodes/usage', body) }),
		user: (userId: () => number | string, query: () => Record<string, unknown> = () => ({})) =>
			createQuery({
				key: () => [userId(), query()],
				enabled: whenPresent(userId),
				fn: () => http.get<Json>('/api/bandwidth-stats/users/:userId', query(), { userId: userId() })
			}),
		squad: (uuid: () => string, query: () => Record<string, unknown> = () => ({})) =>
			createQuery({
				key: () => [uuid(), query()],
				enabled: whenPresent(uuid),
				fn: () =>
					http.get<Json>('/api/bandwidth-stats/internal-squads/:uuid/usage', query(), { uuid: uuid() })
			}),
		squadUser: (
			route: () => { squadUuid: string; userId: number | string },
			query: () => Record<string, unknown> = () => ({})
		) =>
			createQuery({
				key: () => [route(), query()],
				enabled: () => Boolean(route().squadUuid && String(route().userId ?? '') !== ''),
				fn: () =>
					http.get<Json>(
						'/api/bandwidth-stats/internal-squads/:squadUuid/users/:userId/usage',
						query(),
						route()
					)
			})
	};
}
