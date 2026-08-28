import type { HttpClient } from '../http.js';
import { createMutation } from '../mutation.svelte.js';
import { createQuery, whenPresent } from '../query.svelte.js';
import type { HwidDevice, HwidList, Json, TableQuery } from '../types.js';

export function hwidResource(http: HttpClient) {
	return {
		list: (query: () => TableQuery = () => ({})) =>
			createQuery({
				key: query,
				fn: () => http.get<HwidList>('/api/hwid/devices', query())
			}),
		byUser: (userId: () => number | string) =>
			createQuery({
				key: userId,
				enabled: whenPresent(userId),
				fn: () => http.get<HwidList>('/api/hwid/devices/:userId', undefined, { userId: userId() })
			}),
		stats: () => createQuery({ fn: () => http.get<Json>('/api/hwid/devices/stats') }),
		topUsers: () => createQuery({ fn: () => http.get<Json>('/api/hwid/devices/top-users') }),
		create: () => createMutation({ fn: (body: Json) => http.post<HwidDevice>('/api/hwid/devices', body) }),
		remove: () => createMutation({ fn: (body: Json) => http.post<unknown>('/api/hwid/devices/delete', body) }),
		removeAll: () =>
			createMutation({ fn: (body: Json) => http.post<unknown>('/api/hwid/devices/delete-all', body) })
	};
}
