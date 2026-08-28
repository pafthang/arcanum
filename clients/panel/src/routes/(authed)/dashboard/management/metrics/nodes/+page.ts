import { redirect } from '@sveltejs/kit';

export function load() {
	redirect(302, '/dashboard/system/metrics?tab=nodes');
}
