import type { HttpClient } from '../http.js';
import { session } from '../session.svelte.js';
import type { Channel, Message, Reaction } from '../types.js';

export function commsResource(http: HttpClient) {
	const resolveSpace = (sid?: string) => sid || session.spaceId;

	return {
		async listChannels(spaceId?: string, teamId?: string): Promise<Channel[]> {
			const sid = resolveSpace(spaceId);
			const query = teamId ? `?teamId=${teamId}` : '';
			return http.get<Channel[]>(`/api/spaces/${sid}/channels${query}`);
		},
		async createChannel(
			data: { name: string; kind?: 'space' | 'team' | 'dm'; teamId?: string },
			spaceId?: string
		): Promise<Channel> {
			const sid = resolveSpace(spaceId);
			return http.post<Channel>(`/api/spaces/${sid}/channels`, data);
		},
		async getChannel(channelId: string, spaceId?: string): Promise<Channel> {
			const sid = resolveSpace(spaceId);
			return http.get<Channel>(`/api/spaces/${sid}/channels/${channelId}`);
		},
		async listMessages(channelId: string, spaceId?: string): Promise<Message[]> {
			const sid = resolveSpace(spaceId);
			return http.get<Message[]>(`/api/spaces/${sid}/channels/${channelId}/messages`);
		},
		async createMessage(
			channelId: string,
			data: { body: string; parentId?: string; blobId?: string },
			spaceId?: string
		): Promise<Message> {
			const sid = resolveSpace(spaceId);
			return http.post<Message>(`/api/spaces/${sid}/channels/${channelId}/messages`, data);
		},
		async listReactions(channelId: string, messageId: string, spaceId?: string): Promise<Reaction[]> {
			const sid = resolveSpace(spaceId);
			return http.get<Reaction[]>(
				`/api/spaces/${sid}/channels/${channelId}/messages/${messageId}/reactions`
			);
		},
		async addReaction(
			channelId: string,
			messageId: string,
			emoji: string,
			spaceId?: string
		): Promise<Reaction> {
			const sid = resolveSpace(spaceId);
			return http.post<Reaction>(
				`/api/spaces/${sid}/channels/${channelId}/messages/${messageId}/reactions`,
				{ emoji }
			);
		},
		getChannelWSUrl(channelId: string, spaceId?: string): string {
			const sid = resolveSpace(spaceId);
			const proto = typeof window !== 'undefined' && window.location.protocol === 'https:' ? 'wss:' : 'ws:';
			const host = typeof window !== 'undefined' ? window.location.host : '127.0.0.1:8080';
			return `${proto}//${host}/api/spaces/${sid}/channels/${channelId}/ws`;
		}
	};
}
