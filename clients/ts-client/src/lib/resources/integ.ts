import type { HttpClient } from '../http.js';
import { session } from '../session.svelte.js';
import type { Connector, GitHubRepo, Webhook, WebhookDelivery } from '../types.js';

export function integResource(http: HttpClient) {
	const resolveSpace = (sid?: string) => sid || session.spaceId;

	return {
		async listConnectors(spaceId?: string): Promise<Connector[]> {
			const sid = resolveSpace(spaceId);
			return http.get<Connector[]>(`/api/spaces/${sid}/integ/connectors`);
		},
		async createConnector(
			data: { kind: string; name: string; secret?: string; config?: Record<string, any> },
			spaceId?: string
		): Promise<Connector> {
			const sid = resolveSpace(spaceId);
			return http.post<Connector>(`/api/spaces/${sid}/integ/connectors`, data);
		},
		async getConnector(connectorId: string, spaceId?: string): Promise<Connector> {
			const sid = resolveSpace(spaceId);
			return http.get<Connector>(`/api/spaces/${sid}/integ/connectors/${connectorId}`);
		},
		async updateConnector(
			connectorId: string,
			data: { name?: string; status?: string; secret?: string; config?: Record<string, any> },
			spaceId?: string
		): Promise<Connector> {
			const sid = resolveSpace(spaceId);
			return http.patch<Connector>(`/api/spaces/${sid}/integ/connectors/${connectorId}`, data);
		},
		async listRepos(connectorId: string, spaceId?: string): Promise<GitHubRepo[]> {
			const sid = resolveSpace(spaceId);
			return http.get<GitHubRepo[]>(`/api/spaces/${sid}/integ/connectors/${connectorId}/repos`);
		},
		async attachRepo(
			connectorId: string,
			data: { owner: string; name: string; installationId?: string },
			spaceId?: string
		): Promise<GitHubRepo> {
			const sid = resolveSpace(spaceId);
			return http.post<GitHubRepo>(`/api/spaces/${sid}/integ/connectors/${connectorId}/repos`, data);
		},
		async listWebhooks(spaceId?: string): Promise<Webhook[]> {
			const sid = resolveSpace(spaceId);
			return http.get<Webhook[]>(`/api/spaces/${sid}/integ/webhooks`);
		},
		async createWebhook(
			data: { url: string; events: string[]; secret?: string },
			spaceId?: string
		): Promise<Webhook> {
			const sid = resolveSpace(spaceId);
			return http.post<Webhook>(`/api/spaces/${sid}/integ/webhooks`, data);
		},
		async listDeliveries(spaceId?: string): Promise<WebhookDelivery[]> {
			const sid = resolveSpace(spaceId);
			return http.get<WebhookDelivery[]>(`/api/spaces/${sid}/integ/deliveries`);
		},
		getHookURL(connectorId: string): string {
			return `/api/integ/hooks/${connectorId}`;
		},
		getGitHubHookURL(connectorId: string): string {
			return `/api/integ/hooks/github/${connectorId}`;
		},
		getTelegramHookURL(connectorId: string): string {
			return `/api/integ/hooks/telegram/${connectorId}`;
		}
	};
}
