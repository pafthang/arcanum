import type { HttpClient } from '../http.js';
import { session } from '../session.svelte.js';
import type {
	Issue,
	IssueRelation,
	IssueActivity,
	Comment,
	Label,
	Cycle,
	Project,
	View,
	WorkOverview
} from '../types.js';

export function workResource(http: HttpClient) {
	const resolveSpace = (sid?: string) => sid || session.spaceId;

	return {
		async getOverview(spaceId?: string): Promise<WorkOverview> {
			const sid = resolveSpace(spaceId);
			return http.get<WorkOverview>(`/api/spaces/${sid}/work/overview`);
		},
		async listIssues(spaceId?: string): Promise<Issue[]> {
			const sid = resolveSpace(spaceId);
			return http.get<Issue[]>(`/api/spaces/${sid}/issues`);
		},
		async createIssue(
			data: {
				title: string;
				body?: string;
				status?: string;
				assigneeId?: string;
				priority?: string;
				dueAt?: string;
				labelIds?: string[];
			},
			spaceId?: string
		): Promise<Issue> {
			const sid = resolveSpace(spaceId);
			return http.post<Issue>(`/api/spaces/${sid}/issues`, data);
		},
		async getIssue(issueId: string, spaceId?: string): Promise<Issue> {
			const sid = resolveSpace(spaceId);
			return http.get<Issue>(`/api/spaces/${sid}/issues/${issueId}`);
		},
		async updateIssue(issueId: string, data: Partial<Issue>, spaceId?: string): Promise<Issue> {
			const sid = resolveSpace(spaceId);
			return http.patch<Issue>(`/api/spaces/${sid}/issues/${issueId}`, data);
		},
		async getActivity(issueId: string, spaceId?: string): Promise<IssueActivity[]> {
			const sid = resolveSpace(spaceId);
			return http.get<IssueActivity[]>(`/api/spaces/${sid}/issues/${issueId}/activity`);
		},
		async createRelation(
			issueId: string,
			data: { toId: string; kind: 'blocks' | 'blocked_by' | 'duplicate' | 'related' },
			spaceId?: string
		): Promise<IssueRelation> {
			const sid = resolveSpace(spaceId);
			return http.post<IssueRelation>(`/api/spaces/${sid}/issues/${issueId}/relations`, data);
		},
		async listComments(issueId: string, spaceId?: string): Promise<Comment[]> {
			const sid = resolveSpace(spaceId);
			return http.get<Comment[]>(`/api/spaces/${sid}/issues/${issueId}/comments`);
		},
		async addComment(
			issueId: string,
			data: { body: string; blobId?: string },
			spaceId?: string
		): Promise<Comment> {
			const sid = resolveSpace(spaceId);
			return http.post<Comment>(`/api/spaces/${sid}/issues/${issueId}/comments`, data);
		},
		async listLabels(spaceId?: string): Promise<Label[]> {
			const sid = resolveSpace(spaceId);
			return http.get<Label[]>(`/api/spaces/${sid}/labels`);
		},
		async createLabel(data: { name: string; color?: string }, spaceId?: string): Promise<Label> {
			const sid = resolveSpace(spaceId);
			return http.post<Label>(`/api/spaces/${sid}/labels`, data);
		},
		async listCycles(spaceId?: string): Promise<Cycle[]> {
			const sid = resolveSpace(spaceId);
			return http.get<Cycle[]>(`/api/spaces/${sid}/work/cycles`);
		},
		async createCycle(
			data: { name: string; description?: string; status?: string; startDate?: string; endDate?: string },
			spaceId?: string
		): Promise<Cycle> {
			const sid = resolveSpace(spaceId);
			return http.post<Cycle>(`/api/spaces/${sid}/work/cycles`, data);
		},
		async listProjects(spaceId?: string): Promise<Project[]> {
			const sid = resolveSpace(spaceId);
			return http.get<Project[]>(`/api/spaces/${sid}/work/projects`);
		},
		async createProject(
			data: { name: string; key?: string; description?: string; status?: string; leadId?: string },
			spaceId?: string
		): Promise<Project> {
			const sid = resolveSpace(spaceId);
			return http.post<Project>(`/api/spaces/${sid}/work/projects`, data);
		},
		async listViews(spaceId?: string): Promise<View[]> {
			const sid = resolveSpace(spaceId);
			return http.get<View[]>(`/api/spaces/${sid}/work/views`);
		},
		async createView(
			data: { name: string; description?: string; query?: string; icon?: string },
			spaceId?: string
		): Promise<View> {
			const sid = resolveSpace(spaceId);
			return http.post<View>(`/api/spaces/${sid}/work/views`, data);
		}
	};
}
