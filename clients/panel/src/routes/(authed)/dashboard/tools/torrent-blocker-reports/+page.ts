import { redirect } from '@sveltejs/kit';

export function load() {
	redirect(302, '/dashboard/system/logs?tab=torrent');
}
