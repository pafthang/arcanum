import type { HttpClient } from '../http.js';
import { session } from '../session.svelte.js';
import type { Machine } from '../types.js';

export function runtimeResource(http: HttpClient) {
	const resolveSpace = (sid?: string) => sid || session.spaceId;

	return {
		async listMachines(spaceId?: string): Promise<{ items: Machine[] }> {
			const sid = resolveSpace(spaceId);
			return http.get<{ items: Machine[] }>(`/api/spaces/${sid}/machines`);
		},
		async createMachine(
			data: { name: string; image: string; agentId?: string },
			spaceId?: string
		): Promise<Machine> {
			const sid = resolveSpace(spaceId);
			return http.post<Machine>(`/api/spaces/${sid}/machines`, data);
		},
		async getMachine(machineId: string, spaceId?: string): Promise<Machine> {
			const sid = resolveSpace(spaceId);
			return http.get<Machine>(`/api/spaces/${sid}/machines/${machineId}`);
		},
		async stopMachine(machineId: string, spaceId?: string): Promise<Machine> {
			const sid = resolveSpace(spaceId);
			return http.post<Machine>(`/api/spaces/${sid}/machines/${machineId}/stop`, {});
		},
		async execMachine(
			machineId: string,
			cmd: string | string[],
			spaceId?: string
		): Promise<{ stdout: string; stderr: string; exitCode: number }> {
			const sid = resolveSpace(spaceId);
			return http.post<{ stdout: string; stderr: string; exitCode: number }>(
				`/api/spaces/${sid}/machines/${machineId}/exec`,
				{ cmd }
			);
		},
		getProxyURL(machineId: string, port: string, spaceId?: string): string {
			const sid = resolveSpace(spaceId);
			return `/api/spaces/${sid}/machines/${machineId}/proxy/${port}`;
		}
	};
}
