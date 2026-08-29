import type { HttpClient } from '../http.js';
import type { User } from '../types.js';

export interface LoginResponse {
	token: string;
	spaceId: string;
	spaceRole: string;
	user: User;
}

export function authResource(http: HttpClient) {
	return {
		async login(email: string, password: string): Promise<LoginResponse> {
			return http.post<LoginResponse>('/api/auth/login', { email, password });
		},
		async register(email: string, password: string, name: string): Promise<LoginResponse> {
			return http.post<LoginResponse>('/api/auth/register', { email, password, name });
		},
		async loginApiKey(secret: string): Promise<LoginResponse> {
			return http.post<LoginResponse>('/api/auth/api-key', { secret });
		},
		async switchSpace(spaceId: string): Promise<{ token: string; spaceId: string }> {
			return http.post<{ token: string; spaceId: string }>('/api/auth/switch-space', { spaceId });
		}
	};
}
