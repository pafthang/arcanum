import type { ComponentType } from 'svelte';
import {
	Activity,
	CreditCard,
	FileCode,
	Fingerprint,
	Globe,
	House,
	Layers,
	Monitor,
	PanelTop,
	Puzzle,
	Scissors,
	ScrollText,
	Server,
	Settings,
	Shield,
	Users
} from 'lucide-svelte';
import { SETTINGS_NAV, type SettingsNavItem } from '$lib/features/settings/nav';

export type NavItem = {
	label: string;
	href: string;
	id: string;
	icon: ComponentType;
};

export type NavSection = {
	header: string;
	id: string;
	items: NavItem[];
};

export const NAV: NavSection[] = [
	{
		header: 'Home',
		id: 'home',
		items: [{ label: 'Home', href: '/dashboard/home', id: 'home', icon: House }]
	},
	{
		header: 'Users',
		id: 'users',
		items: [
			{ label: 'Users', href: '/dashboard/management/users', id: 'users', icon: Users },
			{
				label: 'Internal',
				href: '/dashboard/management/internal-squads',
				id: 'internal-squads',
				icon: Shield
			},
			{
				label: 'External',
				href: '/dashboard/management/external-squads',
				id: 'external-squads',
				icon: Globe
			},
			{
				label: 'Devices',
				href: '/dashboard/management/devices',
				id: 'hwid',
				icon: Fingerprint
			}
		]
	},
	{
		header: 'Nodes',
		id: 'nodes',
		items: [
			{ label: 'Nodes', href: '/dashboard/management/nodes', id: 'nodes', icon: Server },
			{ label: 'Plugins', href: '/dashboard/management/plugins', id: 'plugins', icon: Puzzle },
			{
				label: 'Billing',
				href: '/dashboard/crm/infra-billing',
				id: 'infra-billing',
				icon: CreditCard
			}
		]
	},
	{
		header: 'Config',
		id: 'profiles',
		items: [
			{
				label: 'Profiles',
				href: '/dashboard/management/config-profiles',
				id: 'config-profiles',
				icon: FileCode
			},
			{ label: 'Snippets', href: '/dashboard/tools/snippets', id: 'snippets', icon: Scissors }
		]
	},
	{
		header: 'Subscription',
		id: 'subscription',
		items: [
			{ label: 'Hosts', href: '/dashboard/management/hosts', id: 'hosts', icon: Monitor },
			{ label: 'Templates', href: '/dashboard/templates', id: 'templates', icon: Layers },
			{ label: 'Page', href: '/dashboard/subpage', id: 'subpage', icon: PanelTop }
		]
	},
	{
		header: 'System',
		id: 'system',
		items: [
			{ label: 'Metrics', href: '/dashboard/system/metrics', id: 'system-metrics', icon: Activity },
			{ label: 'Logs', href: '/dashboard/system/logs', id: 'system-logs', icon: ScrollText }
		]
	}
];

export type HubTab = { value: string; label: string };

export const SYSTEM_HUBS: { href: string; defaultTab: string; tabs: HubTab[] }[] = [
	{
		href: '/dashboard/system/metrics',
		defaultTab: 'panel',
		tabs: [
			{ value: 'panel', label: 'Panel' },
			{ value: 'nodes', label: 'Nodes' },
			{ value: 'http', label: 'HTTP' }
		]
	},
	{
		href: '/dashboard/system/logs',
		defaultTab: 'srh',
		tabs: [
			{ value: 'srh', label: 'Subscription' },
			{ value: 'torrent', label: 'Torrent' },
			{ value: 'sessions', label: 'Sessions' }
		]
	}
];

export function systemHubByPath(pathname: string) {
	return SYSTEM_HUBS.find((hub) => pathname === hub.href);
}

export const SETTINGS_PATH = '/dashboard/management/settings';

export function isSettingsPath(pathname: string): boolean {
	return pathname === SETTINGS_PATH || pathname.startsWith(SETTINGS_PATH + '/');
}

export function settingsItemByHref(pathname: string): SettingsNavItem | undefined {
	const items = SETTINGS_NAV.flatMap((group) => group.items);
	return items.find((item) =>
		item.exact ? pathname === item.href : pathname === item.href || pathname.startsWith(item.href + '/')
	);
}

export const TEMPLATE_TYPES = [
	{ type: 'XRAY_JSON', label: 'Xray JSON' },
	{ type: 'MIHOMO', label: 'Mihomo' },
	{ type: 'STASH', label: 'Stash' },
	{ type: 'SINGBOX', label: 'Singbox' },
	{ type: 'CLASH', label: 'Clash' }
] as const;

export function navItemByHref(pathname: string): NavItem | undefined {
	if (isSettingsPath(pathname)) {
		const item = settingsItemByHref(pathname);
		if (item) {
			return { label: item.label, href: item.href, id: item.href, icon: item.icon };
		}
	}
	const items = NAV.flatMap((section) => section.items);
	return (
		items.find((item) => pathname === item.href || pathname.startsWith(item.href + '/')) ??
		items.find((item) => item.href === '/dashboard/home')
	);
}

export function navSectionByHref(pathname: string): NavSection | undefined {
	return NAV.find((section) =>
		section.items.some((item) => pathname === item.href || pathname.startsWith(item.href + '/'))
	);
}

export type NavCrumb = {
	label: string;
	href?: string;
	icon?: ComponentType;
};

export function navTrail(pathname: string, pageTitle?: string | null): NavCrumb[] {
	if (isSettingsPath(pathname)) {
		const item = settingsItemByHref(pathname);
		const crumbs: NavCrumb[] = [{ label: 'Settings', href: SETTINGS_PATH, icon: Settings }];
		if (item && item.label !== 'Settings') {
			crumbs.push({ label: item.label, href: item.href, icon: item.icon });
		}
		if (pageTitle && pageTitle !== item?.label && pageTitle !== 'Settings') {
			crumbs.push({ label: pageTitle });
		}
		return crumbs;
	}

	const section = navSectionByHref(pathname);
	const item = navItemByHref(pathname);
	if (!item) return [{ label: 'Home', href: '/dashboard/home', icon: House }];

	const crumbs: NavCrumb[] = [];
	if (section && section.header !== item.label) {
		crumbs.push({ label: section.header, href: section.items[0]?.href });
	}
	crumbs.push({ label: item.label, href: item.href, icon: item.icon });
	if (pageTitle && pageTitle !== item.label) {
		crumbs.push({ label: pageTitle });
	}
	return crumbs;
}
