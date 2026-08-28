import type { HttpClient } from '../http.js';
import { createMutation } from '../mutation.svelte.js';
import { prefetchQuery } from '../query-cache.js';
import { createQuery, whenPresent } from '../query.svelte.js';
import type {
	AccessibleNodes,
	Json,
	ResolveUser,
	TableQuery,
	Tags,
	User,
	UserHistory,
	UserList,
	UserStream
} from '../types.js';

export function usersResource(http: HttpClient) {
	return {
		list: (query: () => TableQuery = () => ({})) =>
			createQuery({
				cacheKey: 'users.list',
				key: query,
				fn: () => http.get<UserList>('/api/users', query())
			}),
		prefetchList: (query: TableQuery = { start: 0, size: 20 }) =>
			prefetchQuery({
				cacheKey: 'users.list',
				key: () => query,
				fn: () => http.get<UserList>('/api/users', query)
			}),
		byId: (userId: () => number | string) =>
			createQuery({
				cacheKey: 'users.byId',
				key: userId,
				enabled: whenPresent(userId),
				fn: () => http.get<User>('/api/users/:userId', undefined, { userId: userId() })
			}),
		byUsername: (username: () => string) =>
			createQuery({
				key: username,
				enabled: whenPresent(username),
				fn: () => http.get<User>('/api/users/by-username/:username', undefined, { username: username() })
			}),
		byShortUuid: (shortUuid: () => string) =>
			createQuery({
				key: shortUuid,
				enabled: whenPresent(shortUuid),
				fn: () =>
					http.get<User>('/api/users/by-short-uuid/:shortUuid', undefined, { shortUuid: shortUuid() })
			}),
		tags: () => createQuery({ fn: () => http.get<Tags>('/api/users/tags') }),
		stream: (query: () => Record<string, unknown> = () => ({})) =>
			createQuery({
				key: query,
				fn: () => http.get<UserStream>('/api/users/stream', query())
			}),
		accessibleNodes: (userId: () => number | string) =>
			createQuery({
				key: userId,
				enabled: whenPresent(userId),
				fn: () =>
					http.get<AccessibleNodes>('/api/users/:userId/accessible-nodes', undefined, {
						userId: userId()
					})
			}),
		subscriptionHistory: (userId: () => number | string) =>
			createQuery({
				key: userId,
				enabled: whenPresent(userId),
				fn: () =>
					http.get<UserHistory>('/api/users/:userId/subscription-request-history', undefined, {
						userId: userId()
					})
			}),
		create: () => createMutation({ fn: (body: Json) => http.post<User>('/api/users', body) }),
		update: () => createMutation({ fn: (body: Json) => http.patch<User>('/api/users', body) }),
		remove: () =>
			createMutation({
				fn: (userId: number | string) => http.delete<unknown>('/api/users/:userId', { userId })
			}),
		enable: () =>
			createMutation({
				fn: (userId: number | string) =>
					http.post<User>('/api/users/:userId/actions/enable', undefined, { userId })
			}),
		disable: () =>
			createMutation({
				fn: (userId: number | string) =>
					http.post<User>('/api/users/:userId/actions/disable', undefined, { userId })
			}),
		resetTraffic: () =>
			createMutation({
				fn: (userId: number | string) =>
					http.post<User>('/api/users/:userId/actions/reset-traffic', undefined, { userId })
			}),
		revoke: () =>
			createMutation({
				fn: (vars: { userId: number | string; body?: Json }) =>
					http.post<User>('/api/users/:userId/actions/revoke', vars.body ?? {}, { userId: vars.userId })
			}),
		extend: () =>
			createMutation({
				fn: (vars: { userId: number | string; days: number }) =>
					http.post<User>('/api/users/:userId/actions/extend', { days: vars.days }, { userId: vars.userId })
			}),
		resolve: () =>
			createMutation({ fn: (body: Json) => http.post<ResolveUser>('/api/users/resolve', body) }),
		bulkDeleteByStatus: () =>
			createMutation({ fn: (body: Json) => http.post<unknown>('/api/users/bulk/delete-by-status', body) }),
		bulkUpdate: () =>
			createMutation({ fn: (body: Json) => http.post<unknown>('/api/users/bulk/update', body) }),
		bulkResetTraffic: () =>
			createMutation({ fn: (body: Json) => http.post<unknown>('/api/users/bulk/reset-traffic', body) }),
		bulkRevoke: () =>
			createMutation({
				fn: (body: Json) => http.post<unknown>('/api/users/bulk/revoke-subscription', body)
			}),
		bulkDelete: () =>
			createMutation({ fn: (body: Json) => http.post<unknown>('/api/users/bulk/delete', body) }),
		bulkUpdateSquads: () =>
			createMutation({ fn: (body: Json) => http.post<unknown>('/api/users/bulk/update-squads', body) }),
		bulkExtend: () =>
			createMutation({
				fn: (body: Json) => http.post<unknown>('/api/users/bulk/extend-expiration-date', body)
			}),
		bulkAllUpdate: () =>
			createMutation({ fn: (body: Json) => http.post<unknown>('/api/users/bulk/all/update', body) }),
		bulkAllResetTraffic: () =>
			createMutation({ fn: () => http.post<unknown>('/api/users/bulk/all/reset-traffic') }),
		bulkAllExtend: () =>
			createMutation({
				fn: (body: Json) => http.post<unknown>('/api/users/bulk/all/extend-expiration-date', body)
			})
	};
}
