import type { HttpClient } from '../http.js';
import { session } from '../session.svelte.js';
import type { ApiKey, Space, SpaceMember } from '../types.js';

export function spaceResource(http: HttpClient) {
	const resolveSpace = (sid?: string) => sid || session.spaceId;

	return {
		async list(): Promise<Space[]> {
			return http.get<Space[]>('/api/spaces');
		},
		async create(name: string): Promise<Space> {
			return http.post<Space>('/api/spaces', { name });
		},
		async get(spaceId?: string): Promise<Space> {
			const sid = resolveSpace(spaceId);
			return http.get<Space>(`/api/spaces/${sid}`);
		},
		async getMembers(spaceId?: string): Promise<SpaceMember[]> {
			const sid = resolveSpace(spaceId);
			return http.get<SpaceMember[]>(`/api/spaces/${sid}/members`);
		},
		async inviteMember(email: string, role?: string, spaceId?: string): Promise<SpaceMember> {
			const sid = resolveSpace(spaceId);
			return http.post<SpaceMember>(`/api/spaces/${sid}/members`, { email, role });
		},
		async updateMemberRole(userId: string, role: string, spaceId?: string): Promise<SpaceMember> {
			const sid = resolveSpace(spaceId);
			return http.patch<SpaceMember>(`/api/spaces/${sid}/members/${userId}`, { role });
		},
		async removeMember(userId: string, spaceId?: string): Promise<{ ok: boolean }> {
			const sid = resolveSpace(spaceId);
			return http.delete<{ ok: boolean }>(`/api/spaces/${sid}/members/${userId}`);
		},
		async listKeys(spaceId?: string): Promise<ApiKey[]> {
			const sid = resolveSpace(spaceId);
			return http.get<ApiKey[]>(`/api/spaces/${sid}/keys`);
		},
		async createKey(name: string, actorType?: 'user' | 'agent', spaceId?: string): Promise<ApiKey & { keySecret: string }> {
			const sid = resolveSpace(spaceId);
			return http.post<ApiKey & { keySecret: string }>(`/api/spaces/${sid}/keys`, { name, actorType });
		}
	};
}
