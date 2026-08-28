import type { HttpClient } from '../http.js';
import { createMutation } from '../mutation.svelte.js';
import { createQuery } from '../query.svelte.js';
import type { ApiToken, ApiTokenList, Json, Ott } from '../types.js';

export function tokensResource(http: HttpClient) {
	return {
		list: () => createQuery({ fn: () => http.get<ApiTokenList>('/api/tokens') }),
		scopes: () => createQuery({ fn: () => http.get<Json>('/api/tokens/scopes') }),
		ott: () => createQuery({ fn: () => http.get<Ott>('/api/tokens/ott') }),
		create: () => createMutation({ fn: (body: Json) => http.post<ApiToken>('/api/tokens', body) }),
		remove: () =>
			createMutation({ fn: (uuid: string) => http.delete<unknown>('/api/tokens/:uuid', { uuid }) })
	};
}
