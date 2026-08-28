import { redirect } from '@sveltejs/kit';

export function load({ url }) {
	const next = new URL('/dashboard/system/logs', url.origin);
	next.searchParams.set('tab', 'sessions');
	for (const key of ['kind', 'id', 'drop'] as const) {
		const value = url.searchParams.get(key);
		if (value) next.searchParams.set(key, value);
	}
	redirect(302, next.pathname + next.search);
}
