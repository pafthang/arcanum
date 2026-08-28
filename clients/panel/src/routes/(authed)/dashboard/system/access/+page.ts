import { redirect } from '@sveltejs/kit';

export function load({ url }) {
	if (url.searchParams.get('tab') === 'hwid') {
		redirect(302, '/dashboard/management/devices');
	}
	const next = new URL('/dashboard/system/logs', url.origin);
	next.searchParams.set('tab', 'sessions');
	for (const key of ['kind', 'id', 'drop'] as const) {
		const value = url.searchParams.get(key);
		if (value) next.searchParams.set(key, value);
	}
	redirect(302, next.pathname + next.search);
}
