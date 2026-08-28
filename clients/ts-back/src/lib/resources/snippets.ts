import type { HttpClient } from '../http.js';
import { createMutation } from '../mutation.svelte.js';
import { createQuery } from '../query.svelte.js';
import type { Json, Snippet, SnippetList } from '../types.js';

export function snippetsResource(http: HttpClient) {
	return {
		list: () => createQuery({ fn: () => http.get<SnippetList>('/api/snippets') }),
		create: () => createMutation({ fn: (body: Json) => http.post<Snippet>('/api/snippets', body) }),
		update: () => createMutation({ fn: (body: Json) => http.patch<Snippet>('/api/snippets', body) }),
		remove: () => createMutation({ fn: (body: Json) => http.delete<unknown>('/api/snippets', undefined, body) }),
		sync: () => createMutation({ fn: (body?: Json) => http.post<unknown>('/api/snippets/actions/sync', body) })
	};
}
