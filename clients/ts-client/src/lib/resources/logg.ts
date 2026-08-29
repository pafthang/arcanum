import type { HttpClient } from '../http.js';
import { session } from '../session.svelte.js';
import type { LogActivity } from '../types.js';

export function loggResource(http: HttpClient) {
	return {
		async getLogs(spaceId?: string): Promise<LogActivity[]> {
			const sid = spaceId || session.spaceId;
			const query = sid ? `?spaceId=${sid}` : '';
			return http.get<LogActivity[]>(`/api/logs${query}`);
		}
	};
}
