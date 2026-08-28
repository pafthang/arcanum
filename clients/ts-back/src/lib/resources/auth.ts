import type { HttpClient } from '../http.js';
import { createMutation } from '../mutation.svelte.js';
import { createQuery } from '../query.svelte.js';
import { session } from '../session.svelte.js';
import type { AuthStatus, Credentials, Json, OAuth2Authorize, TokenResponse } from '../types.js';

export function authResource(http: HttpClient) {
	return {
		status: () =>
			createQuery({
				cacheKey: 'auth.status',
				fn: () => http.get<AuthStatus>('/api/auth/status')
			}),
		login: () =>
			createMutation({
				fn: async (body: Credentials) => {
					const res = await http.post<TokenResponse>('/api/auth/login', body);
					session.setToken(res.accessToken);
					return res;
				}
			}),
		register: () =>
			createMutation({
				fn: async (body: Credentials) => {
					const res = await http.post<TokenResponse>('/api/auth/register', body);
					session.setToken(res.accessToken);
					return res;
				}
			}),
		oauth2Authorize: () =>
			createMutation({
				fn: (body: OAuth2Authorize) =>
					http.post<{ authorizationUrl: string }>('/api/auth/oauth2/authorize', body)
			}),
		oauth2Callback: () =>
			createMutation({
				fn: async (body: OAuth2Authorize) => {
					const res = await http.post<TokenResponse>('/api/auth/oauth2/callback', body);
					session.setToken(res.accessToken);
					return res;
				}
			}),
		passkeyAuthenticationOptions: () =>
			createMutation({
				fn: () => http.get<Json>('/api/auth/passkey/authentication/options')
			}),
		passkeyAuthenticationVerify: () =>
			createMutation({
				fn: async (body: { response: Json }) => {
					const res = await http.post<TokenResponse>('/api/auth/passkey/authentication/verify', body);
					session.setToken(res.accessToken);
					return res;
				}
			}),
		logout: () => session.clear()
	};
}
