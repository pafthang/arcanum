import type { HttpClient } from '../http.js';
import { session } from '../session.svelte.js';
import type { Team } from '../types.js';

export function teamsResource(http: HttpClient) {
	const resolveSpace = (sid?: string) => sid || session.spaceId;

	return {
		async list(spaceId?: string): Promise<Team[]> {
			const sid = resolveSpace(spaceId);
			return http.get<Team[]>(`/api/spaces/${sid}/teams`);
		},
		async create(data: { name: string; description?: string }, spaceId?: string): Promise<Team> {
			const sid = resolveSpace(spaceId);
			return http.post<Team>(`/api/spaces/${sid}/teams`, data);
		},
		async get(teamId: string, spaceId?: string): Promise<Team> {
			const sid = resolveSpace(spaceId);
			return http.get<Team>(`/api/spaces/${sid}/teams/${teamId}`);
		},
		async update(teamId: string, data: { name?: string; description?: string }, spaceId?: string): Promise<Team> {
			const sid = resolveSpace(spaceId);
			return http.patch<Team>(`/api/spaces/${sid}/teams/${teamId}`, data);
		},
		async addMember(teamId: string, userId: string, spaceId?: string): Promise<{ ok: boolean }> {
			const sid = resolveSpace(spaceId);
			return http.post<{ ok: boolean }>(`/api/spaces/${sid}/teams/${teamId}/members`, { userId });
		}
	};
}
