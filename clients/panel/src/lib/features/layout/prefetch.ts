import { rw } from '$lib/rw';

const USERS_QUERY = { start: 0, size: 20 };

export function prefetchRoute(href: string): void {
	switch (href) {
		case '/dashboard/home':
			rw.system.prefetchHome();
			break;
		case '/dashboard/management/users':
			rw.users.prefetchList(USERS_QUERY);
			break;
		case '/dashboard/management/internal-squads':
			rw.squads.internal.prefetchList();
			rw.configProfiles.prefetchAllInbounds();
			break;
		case '/dashboard/management/external-squads':
			rw.squads.external.prefetchList();
			break;
		case '/dashboard/management/nodes':
			rw.nodes.prefetchList();
			rw.configProfiles.prefetchList();
			break;
		case '/dashboard/management/plugins':
			rw.plugins.prefetchList();
			break;
		case '/dashboard/management/hosts':
			rw.hosts.prefetchList();
			rw.configProfiles.prefetchAllInbounds();
			break;
		case '/dashboard/management/config-profiles':
			rw.configProfiles.prefetchList();
			break;
	}
}
