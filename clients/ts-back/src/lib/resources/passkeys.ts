import type { HttpClient } from '../http.js';
import { createMutation } from '../mutation.svelte.js';
import { createQuery } from '../query.svelte.js';
import type { Json } from '../types.js';

export function passkeysResource(http: HttpClient) {
	return {
		list: () => createQuery({ fn: () => http.get<Json>('/api/passkeys') }),
		registrationOptions: () =>
			createQuery({ fn: () => http.get<Json>('/api/passkeys/registration/options') }),
		registrationVerify: () =>
			createMutation({ fn: (body: Json) => http.post<Json>('/api/passkeys/registration/verify', body) }),
		update: () => createMutation({ fn: (body: Json) => http.patch<Json>('/api/passkeys', body) }),
		remove: () => createMutation({ fn: (body?: Json) => http.delete<unknown>('/api/passkeys', undefined, body) })
	};
}
