import type { HttpClient } from '../http.js';
import { createMutation } from '../mutation.svelte.js';
import { prefetchQuery } from '../query-cache.js';
import { createQuery, whenPresent } from '../query.svelte.js';
import type { ExternalSquad, InternalSquad, Json, ReorderItem } from '../types.js';

export function squadsResource(http: HttpClient) {
	return {
		internal: {
			list: () =>
				createQuery({
					cacheKey: 'internal-squads.list',
					fn: () => http.get<{ total: number; internalSquads: InternalSquad[] }>('/api/internal-squads')
				}),
			prefetchList: () =>
				prefetchQuery({
					cacheKey: 'internal-squads.list',
					fn: () => http.get<{ total: number; internalSquads: InternalSquad[] }>('/api/internal-squads')
				}),
			byUuid: (uuid: () => string) =>
				createQuery({
					key: uuid,
					enabled: whenPresent(uuid),
					fn: () => http.get<InternalSquad>('/api/internal-squads/:uuid', undefined, { uuid: uuid() })
				}),
			accessibleNodes: (uuid: () => string) =>
				createQuery({
					key: uuid,
					enabled: whenPresent(uuid),
					fn: () =>
						http.get<Json>('/api/internal-squads/:uuid/accessible-nodes', undefined, { uuid: uuid() })
				}),
			create: () =>
				createMutation({ fn: (body: Json) => http.post<InternalSquad>('/api/internal-squads', body) }),
			update: () =>
				createMutation({ fn: (body: Json) => http.patch<InternalSquad>('/api/internal-squads', body) }),
			remove: () =>
				createMutation({
					fn: (uuid: string) => http.delete<unknown>('/api/internal-squads/:uuid', { uuid })
				}),
			reorder: () =>
				createMutation({
					fn: (body: { items: ReorderItem[] }) =>
						http.post<unknown>('/api/internal-squads/actions/reorder', body)
				}),
			addUsers: () =>
				createMutation({
					fn: (vars: { uuid: string; body?: Json }) =>
						http.post<unknown>('/api/internal-squads/:uuid/bulk-actions/add-users', vars.body, {
							uuid: vars.uuid
						})
				}),
			removeUsers: () =>
				createMutation({
					fn: (vars: { uuid: string; body?: Json }) =>
						http.delete<unknown>('/api/internal-squads/:uuid/bulk-actions/remove-users', { uuid: vars.uuid }, vars.body)
				})
		},
		external: {
			list: () =>
				createQuery({
					cacheKey: 'external-squads.list',
					fn: () => http.get<{ total: number; externalSquads: ExternalSquad[] }>('/api/external-squads')
				}),
			prefetchList: () =>
				prefetchQuery({
					cacheKey: 'external-squads.list',
					fn: () => http.get<{ total: number; externalSquads: ExternalSquad[] }>('/api/external-squads')
				}),
			byUuid: (uuid: () => string) =>
				createQuery({
					key: uuid,
					enabled: whenPresent(uuid),
					fn: () => http.get<ExternalSquad>('/api/external-squads/:uuid', undefined, { uuid: uuid() })
				}),
			create: () =>
				createMutation({ fn: (body: Json) => http.post<ExternalSquad>('/api/external-squads', body) }),
			update: () =>
				createMutation({ fn: (body: Json) => http.patch<ExternalSquad>('/api/external-squads', body) }),
			remove: () =>
				createMutation({
					fn: (uuid: string) => http.delete<unknown>('/api/external-squads/:uuid', { uuid })
				}),
			reorder: () =>
				createMutation({
					fn: (body: { items: ReorderItem[] }) =>
						http.post<unknown>('/api/external-squads/actions/reorder', body)
				}),
			addUsers: () =>
				createMutation({
					fn: (vars: { uuid: string; body?: Json }) =>
						http.post<unknown>('/api/external-squads/:uuid/bulk-actions/add-users', vars.body, {
							uuid: vars.uuid
						})
				}),
			removeUsers: () =>
				createMutation({
					fn: (vars: { uuid: string; body?: Json }) =>
						http.delete<unknown>(
							'/api/external-squads/:uuid/bulk-actions/remove-users',
							{ uuid: vars.uuid },
							vars.body
						)
				})
		}
	};
}
