import type { HttpClient } from '../http.js';
import { session } from '../session.svelte.js';
import type { AgentRun, Memory, Skill } from '../types.js';

export function agentsResource(http: HttpClient) {
	const resolveSpace = (sid?: string) => sid || session.spaceId;

	return {
		async listRuns(spaceId?: string): Promise<AgentRun[]> {
			const sid = resolveSpace(spaceId);
			return http.get<AgentRun[]>(`/api/spaces/${sid}/agents/runs`);
		},
		async createRun(
			data: { agentId: string; input: string; issueId?: string },
			spaceId?: string
		): Promise<AgentRun> {
			const sid = resolveSpace(spaceId);
			return http.post<AgentRun>(`/api/spaces/${sid}/agents/runs`, data);
		},
		async getRun(runId: string, spaceId?: string): Promise<AgentRun> {
			const sid = resolveSpace(spaceId);
			return http.get<AgentRun>(`/api/spaces/${sid}/agents/runs/${runId}`);
		},
		async cancelRun(runId: string, spaceId?: string): Promise<AgentRun> {
			const sid = resolveSpace(spaceId);
			return http.post<AgentRun>(`/api/spaces/${sid}/agents/runs/${runId}/cancel`, {});
		},
		async listMemory(agentId: string, spaceId?: string): Promise<Memory[]> {
			const sid = resolveSpace(spaceId);
			return http.get<Memory[]>(`/api/spaces/${sid}/agents/memory?agentId=${agentId}`);
		},
		async putMemory(
			agentId: string,
			data: { key: string; value: string; tier?: 'working' | 'episodic' | 'semantic' },
			spaceId?: string
		): Promise<Memory> {
			const sid = resolveSpace(spaceId);
			return http.post<Memory>(`/api/spaces/${sid}/agents/memory?agentId=${agentId}`, data);
		},
		async listSkills(spaceId?: string): Promise<Skill[]> {
			const sid = resolveSpace(spaceId);
			return http.get<Skill[]>(`/api/spaces/${sid}/agents/skills`);
		},
		async createSkill(data: { name: string; body: string }, spaceId?: string): Promise<Skill> {
			const sid = resolveSpace(spaceId);
			return http.post<Skill>(`/api/spaces/${sid}/agents/skills`, data);
		}
	};
}
